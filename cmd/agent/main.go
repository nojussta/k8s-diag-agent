// Command agent diagnoses one alert and writes diagnosis.json, metrics.json,
// trace.jsonl and report.md. Two modes:
//
//   - agent (default): starts the MCP server, gives the model the read-only
//     tools and runs the investigation loop (the model chooses the tools);
//   - pipeline: the fixed read-only sequence (evidence.json) has already
//     collected the data; the model is called once to combine it.
//
// With --notify URL (or NOTIFY_WEBHOOK_URL) the report is also posted to a
// chat webhook. Used by labctl:
//
//	labctl scenario s1-oomkilled --method agent="build/agent --alert {alert} --out {out} --notify {notify}"
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/nojussta/k8s-diag-agent/docs"
	"github.com/nojussta/k8s-diag-agent/internal/agent"
	"github.com/nojussta/k8s-diag-agent/internal/alertmsg"
	"github.com/nojussta/k8s-diag-agent/internal/lab"
	"github.com/nojussta/k8s-diag-agent/internal/notify"
	"github.com/nojussta/k8s-diag-agent/internal/tools"
	"time"
)

func main() {
	alertPath := flag.String("alert", "", "alert JSON (a single Alertmanager alert, or a webhook payload)")
	out := flag.String("out", ".", "output directory")
	server := flag.String("server", "", "mcp-server binary (default: <root>/build/mcp-server)")
	model := flag.String("model", "", "model id (default: $GEMINI_MODEL / "+agent.DefaultGeminiModel+" for gemini, "+agent.DefaultLocalModel+" for apfel)")
	provider := flag.String("provider", "", "auto | apfel | gemini | openai (default: $LLM_PROVIDER or auto: local apfel when installed, else Gemini); openai = any OpenAI-compatible server, see OPENAI_BASE_URL")
	mode := flag.String("mode", "agent", "agent (the model chooses tools) | pipeline (one model call over --evidence)")
	evidencePath := flag.String("evidence", "", "pipeline mode: evidence.json from the fixed action sequence (default: <out>/../evidence.json)")
	notifyURL := flag.String("notify", "", "chat webhook for the report (default: $NOTIFY_WEBHOOK_URL; empty = file only)")
	flag.Parse()
	if *alertPath == "" {
		fmt.Fprintln(os.Stderr, "usage: agent --alert FILE [--out DIR] [--mode agent|pipeline] [--provider P] [--model ID] [--notify URL]")
		os.Exit(2)
	}
	if *mode != "agent" && *mode != "pipeline" {
		fmt.Fprintf(os.Stderr, "unknown mode %q (use agent or pipeline)\n", *mode)
		os.Exit(2)
	}
	if err := run(options{alertPath: *alertPath, out: *out, server: *server, model: *model, provider: *provider,
		mode: *mode, evidence: *evidencePath, notify: *notifyURL}); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

type options struct {
	alertPath, out, server, model, provider, mode, evidence, notify string
}

func run(o options) error {
	alertPath, out, server, model, provider := o.alertPath, o.out, o.server, o.model, o.provider
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	root, err := lab.FindRoot()
	if err != nil {
		return err
	}
	loadDotEnv(filepath.Join(root, ".env"))
	if provider == "" {
		provider = os.Getenv("LLM_PROVIDER")
	}
	if provider == "" || provider == "auto" {
		// Local first: it costs nothing and nothing leaves the machine.
		provider = "gemini"
		if agent.ApfelAvailable(os.Getenv("APFEL_URL")) {
			provider = "apfel"
		}
	}
	if provider != "apfel" && provider != "gemini" && provider != "openai" {
		return fmt.Errorf("unknown provider %q (use auto, apfel, gemini or openai)", provider)
	}
	if provider == "openai" {
		if os.Getenv("OPENAI_BASE_URL") == "" {
			return fmt.Errorf("provider openai needs OPENAI_BASE_URL (e.g. http://127.0.0.1:11434 for Ollama) and OPENAI_MODEL")
		}
		if model == "" {
			model = os.Getenv("OPENAI_MODEL")
		}
		if model == "" {
			return fmt.Errorf("provider openai needs --model or OPENAI_MODEL")
		}
	}
	if model == "" && provider == "gemini" {
		model = os.Getenv("GEMINI_MODEL")
	}
	key := ""
	if provider == "gemini" {
		key = os.Getenv("GEMINI_API_KEY")
		if key == "" {
			return fmt.Errorf("GEMINI_API_KEY is not set (put it in %s/.env, or use --provider apfel)", root)
		}
		if model == "" {
			model = agent.DefaultGeminiModel
		}
	} else if model == "" && provider == "apfel" {
		model = agent.DefaultLocalModel
	}
	if server == "" {
		server = filepath.Join(root, "build", "mcp-server")
	}

	raw, err := os.ReadFile(alertPath)
	if err != nil {
		return err
	}
	alert, err := alertJSON(raw)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		return err
	}

	traceFile, err := os.Create(filepath.Join(out, "trace.jsonl"))
	if err != nil {
		return err
	}
	defer traceFile.Close()

	var defs []agent.ToolDef
	var client *agent.MCPTools
	allowed := map[string]bool{}
	if o.mode == "agent" {
		mcpCmd := exec.Command(server)
		mcpCmd.Env = os.Environ()
		client, err = agent.Connect(ctx, mcpCmd)
		if err != nil {
			return err
		}
		defer client.Close()
		for _, n := range tools.Names {
			allowed[n] = true
		}
		if defs, err = client.Defs(ctx, allowed); err != nil {
			return err
		}
	}

	var llm agent.Model
	limits := agent.Limits{}
	if provider == "apfel" {
		url, stopApfel, err := agent.StartApfel(ctx, os.Getenv("APFEL_URL"))
		if err != nil {
			return err
		}
		defer stopApfel()
		local := agent.NewLocal(url, model, defs, docs.DiagnosisSchema)
		if o.mode == "pipeline" {
			local.System = agent.LocalPipelinePrompt
		}
		if d := os.Getenv("LOCAL_LLM_DUMP"); d != "" {
			if f, err := os.Create(d); err == nil {
				defer f.Close()
				local.Dump = f
			}
		}
		llm = local
		limits = agent.LocalLimits
	} else if provider == "openai" {
		window, _ := strconv.Atoi(os.Getenv("LLM_CONTEXT_TOKENS")) // 0 = no fitting
		c := agent.NewOpenAI(os.Getenv("OPENAI_BASE_URL"), model, os.Getenv("OPENAI_API_KEY"), defs, docs.DiagnosisSchema, window)
		if o.mode == "pipeline" {
			c.System = agent.PipelineSystemPrompt
		}
		llm = c
	} else {
		g := agent.NewGemini(key, model, defs, docs.DiagnosisSchema)
		if d, err := time.ParseDuration(os.Getenv("GEMINI_MIN_INTERVAL")); err == nil {
			g.MinInterval = d // e.g. 4s on the free tier (15 requests per minute)
		}
		if o.mode == "pipeline" {
			g.System = agent.PipelineSystemPrompt
		}
		llm = g
	}
	fmt.Printf("mode=%s provider=%s model=%s\n", o.mode, provider, model)

	var rep agent.Report
	var runErr error
	if o.mode == "pipeline" {
		evPath := o.evidence
		if evPath == "" {
			evPath = filepath.Join(filepath.Dir(filepath.Clean(out)), "evidence.json")
		}
		evRaw, err := os.ReadFile(evPath)
		if err != nil {
			return fmt.Errorf("pipeline mode needs the fixed-sequence evidence: %w", err)
		}
		trace := func(event string, v any) { writeTrace(traceFile, event, v) }
		trace("alert", alert)
		p := &agent.Pipeline{Model: llm, ModelName: model, ToolTurns: provider == "apfel", Compact: provider == "apfel",
			MaxActionBytes: 8 << 10, Trace: trace}
		pctx, cancel := context.WithTimeout(ctx, limitsTimeout(limits))
		rep, runErr = p.Diagnose(pctx, alert, evRaw)
		cancel()
	} else {
		a := &agent.Agent{
			Model:     llm,
			Limits:    limits,
			ModelName: model,
			Tools:     client,
			Allowed:   allowed,
			Trace:     traceFile,
		}
		rep, runErr = a.Investigate(ctx, alert)
	}

	if err := writeJSON(filepath.Join(out, "metrics.json"), rep.Metrics); err != nil {
		return err
	}
	if rep.Diagnosis != nil {
		if err := writeJSON(filepath.Join(out, "diagnosis.json"), rep.Diagnosis); err != nil {
			return err
		}
		d := rep.Diagnosis
		fmt.Printf("%s | %s | confidence %s\n%s\n", d.RootCause.Category, d.RootCause.Component, d.Confidence, d.Summary)
		text := notify.Text(alertName(alert), d, rep.Metrics)
		if err := os.WriteFile(filepath.Join(out, "report.md"), []byte(text), 0o644); err != nil {
			return err
		}
		url := o.notify
		if url == "" {
			url = os.Getenv("NOTIFY_WEBHOOK_URL")
		}
		if url != "" {
			if err := notify.Post(ctx, url, text); err != nil {
				fmt.Fprintln(os.Stderr, "notification failed (the report is still in report.md):", err)
			} else {
				fmt.Println("notification delivered to the webhook")
			}
		}
	}
	fmt.Printf("mode=%s model=%s tool_calls=%d model_steps=%d rejected=%d repairs=%d duration=%.1fs tokens in/out=%d/%d stop=%s\n",
		rep.Metrics.Mode, rep.Metrics.Model, rep.Metrics.ToolCalls, rep.Metrics.ModelSteps, rep.Metrics.Rejected, rep.Metrics.Repairs,
		rep.Metrics.DurationS, rep.Metrics.InputTokens, rep.Metrics.OutputTokens, rep.Metrics.StopReason)
	return runErr
}

func limitsTimeout(l agent.Limits) time.Duration {
	if l.Timeout > 0 {
		return l.Timeout
	}
	return agent.DefaultLimits.Timeout
}

func writeTrace(w io.Writer, event string, v any) {
	b, _ := json.Marshal(map[string]any{"t": time.Now().UTC().Format(time.RFC3339Nano), "event": event, "data": v})
	fmt.Fprintf(w, "%s\n", b)
}

func alertName(alert string) string {
	var a struct {
		Labels map[string]string `json:"labels"`
	}
	_ = json.Unmarshal([]byte(alert), &a)
	if n := a.Labels["alertname"]; n != "" {
		return n
	}
	return "alert"
}

// alertJSON accepts one alert or a webhook payload and returns one alert as JSON.
func alertJSON(raw []byte) (string, error) {
	var p alertmsg.Payload
	if json.Unmarshal(raw, &p) == nil && len(p.Alerts) > 0 {
		b, _ := json.MarshalIndent(p.Alerts[0], "", "  ")
		return string(b), nil
	}
	var a alertmsg.Alert
	if err := json.Unmarshal(raw, &a); err != nil || len(a.Labels) == 0 {
		return "", fmt.Errorf("alert file is neither an alert nor a webhook payload")
	}
	b, _ := json.MarshalIndent(a, "", "  ")
	return string(b), nil
}

func writeJSON(path string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

// loadDotEnv reads KEY=VALUE lines; variables already set are not overridden.
func loadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		k, v = strings.TrimSpace(k), strings.Trim(strings.TrimSpace(v), `"'`)
		if _, set := os.LookupEnv(k); !set && k != "" {
			os.Setenv(k, v)
		}
	}
}
