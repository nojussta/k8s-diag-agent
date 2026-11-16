// Command diag-agent is the agent as it runs inside the cluster: one process
// that receives Alertmanager webhooks, queues the firing alerts, investigates
// them one at a time with the read-only tools and posts each report to the
// chat webhook. Nothing is started by hand.
//
//	POST /alerts                         Alertmanager webhook (payload version "4")
//	GET  /alerts[?since=<unix ts>]       received notifications (used by labctl)
//	GET  /investigations                 every queued, finished or skipped alert
//	GET  /investigations/{id}/{file}     trace.jsonl, diagnosis.json, metrics.json or report.md
//	POST /sink, GET /messages            built-in stand-in for a chat channel (no NOTIFY_WEBHOOK_URL)
//	GET  /healthz                        liveness
//
// Environment:
//
//	AUTO_INVESTIGATE   "true" starts an investigation for every firing alert; anything else only records alerts
//	LLM_PROVIDER       gemini (default) or openai (any OpenAI-compatible server in or outside the cluster)
//	GEMINI_API_KEY     from a Secret; GEMINI_MODEL, GEMINI_MIN_INTERVAL, GEMINI_TEMPERATURE are optional
//	NOTIFY_WEBHOOK_URL chat webhook; empty = the built-in stand-in
//	QUEUE_BUDGET, QUEUE_WINDOW   investigations accepted per window (default 5 per 10m)
//	WORK_DIR           writable directory for traces (default /work; the root filesystem is read-only)
//	MCP_SERVER         path of the MCP server binary (default /mcp-server)
package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"time"

	// The image is built FROM scratch and has no system CA bundle; this embeds
	// one, so HTTPS to the model provider can be verified.
	_ "golang.org/x/crypto/x509roots/fallback"

	"github.com/nojussta/k8s-diag-agent/internal/alertmsg"
	"github.com/nojussta/k8s-diag-agent/internal/investigate"
	"github.com/nojussta/k8s-diag-agent/internal/notify"
	"github.com/nojussta/k8s-diag-agent/internal/receiver"
	"github.com/nojussta/k8s-diag-agent/internal/scenario"
	"github.com/nojussta/k8s-diag-agent/internal/watch"
)

var servedFiles = map[string]string{
	"trace.jsonl": "application/x-ndjson", "diagnosis.json": "application/json",
	"metrics.json": "application/json", "report.md": "text/markdown; charset=utf-8",
}

type service struct {
	cfg     investigate.Config
	queue   *watch.Queue
	workDir string
	notify  string
	logMu   sync.Mutex
	msgMu   sync.Mutex
	msgs    []scenario.Message
}

func (s *service) logEvent(event string, fields map[string]any) {
	fields["event"] = event
	s.logMu.Lock()
	defer s.logMu.Unlock()
	_ = json.NewEncoder(os.Stdout).Encode(fields)
}

func main() {
	port := envInt("PORT", 8080)
	auto := os.Getenv("AUTO_INVESTIGATE") == "true"
	window, _ := time.ParseDuration(os.Getenv("QUEUE_WINDOW"))
	s := &service{
		queue:   watch.NewQueue(envInt("QUEUE_BUDGET", 0), window, 0),
		workDir: envStr("WORK_DIR", "/work"),
		notify:  envStr("NOTIFY_WEBHOOK_URL", "http://127.0.0.1:"+strconv.Itoa(port)+"/sink"),
		cfg:     investigate.Config{Mode: "agent", Server: envStr("MCP_SERVER", "/mcp-server"), Log: os.Stdout},
	}
	if s.cfg.Provider = os.Getenv("LLM_PROVIDER"); s.cfg.Provider == "" {
		s.cfg.Provider = "gemini"
	}
	s.cfg.Notify = s.notify
	if auto {
		// Fail at start-up, not at the first alert, when the model is not configured.
		if _, err := s.cfg.Resolve(); err != nil {
			log.Fatalf("AUTO_INVESTIGATE=true but the model is not configured: %v", err)
		}
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	store := receiver.NewStore(envInt("MAX_STORED", 500))
	mux := receiver.NewMux(store, os.Stdout, func(p alertmsg.Payload) {
		if auto {
			s.offer(ctx, p.Alerts)
		}
	})
	mux.HandleFunc("GET /investigations", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, s.queue.Items())
	})
	mux.HandleFunc("GET /investigations/{id}/{file}", func(w http.ResponseWriter, r *http.Request) {
		ctype, ok := servedFiles[r.PathValue("file")]
		if !ok {
			http.NotFound(w, r)
			return
		}
		// filepath.Base keeps the id inside the work directory.
		b, err := os.ReadFile(filepath.Join(s.workDir, filepath.Base(r.PathValue("id")), r.PathValue("file")))
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", ctype)
		_, _ = w.Write(b)
	})
	mux.HandleFunc("POST /sink", func(w http.ResponseWriter, r *http.Request) {
		var p struct {
			Text string `json:"text"`
		}
		b, _ := io.ReadAll(io.LimitReader(r.Body, 1<<20))
		if json.Unmarshal(b, &p) != nil || p.Text == "" {
			http.Error(w, `expected {"text": "..."}`, http.StatusBadRequest)
			return
		}
		s.msgMu.Lock()
		s.msgs = append(s.msgs, scenario.Message{Received: time.Now().UTC().Format(time.RFC3339), Text: p.Text})
		if len(s.msgs) > 200 {
			s.msgs = s.msgs[len(s.msgs)-200:]
		}
		s.msgMu.Unlock()
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("GET /messages", func(w http.ResponseWriter, r *http.Request) {
		s.msgMu.Lock()
		defer s.msgMu.Unlock()
		writeJSON(w, append([]scenario.Message{}, s.msgs...))
	})

	srv := &http.Server{Addr: ":" + strconv.Itoa(port), Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdown)
	}()
	if auto {
		go s.work(ctx)
	}
	s.logEvent("started", map[string]any{"port": port, "auto_investigate": auto, "provider": s.cfg.Provider,
		"budget": s.queue.Budget, "window": s.queue.Window.String()})
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

// offer queues the alerts of one notification and reports the ones over the budget.
func (s *service) offer(ctx context.Context, alerts []alertmsg.Alert) {
	queued, skipped := s.queue.Offer(alerts, time.Now())
	for _, it := range queued {
		s.logEvent("queued", map[string]any{"id": it.ID, "alert": watch.Target(it.Alert)})
	}
	if len(skipped) == 0 {
		return
	}
	s.logEvent("storm", map[string]any{"queued": len(queued), "skipped": len(skipped), "reason": skipped[0].Reason})
	if err := notify.Post(ctx, s.notify, watch.StormText(queued, skipped)); err != nil {
		s.logEvent("notify_failed", map[string]any{"error": err.Error()})
	}
}

// work is the single worker: one investigation at a time, in arrival order.
func (s *service) work(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case it := <-s.queue.Pending():
			start := time.Now()
			s.queue.Update(it, func(i *watch.Item) { i.Status, i.Started = "running", &start })
			alert, _ := json.MarshalIndent(it.Alert, "", "  ")
			res, err := investigate.Run(ctx, s.cfg, string(alert), filepath.Join(s.workDir, it.ID))
			end := time.Now()
			s.queue.Update(it, func(i *watch.Item) {
				i.Finished, i.Notified = &end, res.Notified
				i.Result = map[string]any{"metrics": res.Report.Metrics, "diagnosis": res.Report.Diagnosis}
				i.Status = "done"
				if err != nil || res.Report.Diagnosis == nil {
					i.Status = "failed"
					if err != nil {
						i.Reason = err.Error()
					}
				}
			})
			s.logEvent("investigated", map[string]any{"id": it.ID, "alert": watch.Target(it.Alert), "status": it.Status,
				"seconds": end.Sub(start).Seconds(), "tool_calls": res.Report.Metrics.ToolCalls, "notified": res.Notified})
		}
	}
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func envInt(name string, def int) int {
	if v, err := strconv.Atoi(os.Getenv(name)); err == nil && v > 0 {
		return v
	}
	return def
}

func envStr(name, def string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return def
}
