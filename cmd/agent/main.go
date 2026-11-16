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
//
// The same investigation code runs inside the cluster as cmd/diag-agent, where
// it is started by the alert itself.
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"github.com/nojussta/k8s-diag-agent/internal/agent"
	"github.com/nojussta/k8s-diag-agent/internal/investigate"
	"github.com/nojussta/k8s-diag-agent/internal/lab"
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
	if err := run(*alertPath, *out, investigate.Config{Provider: *provider, Model: *model, Mode: *mode,
		Server: *server, Evidence: *evidencePath, Notify: *notifyURL, Log: os.Stdout}); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(alertPath, out string, cfg investigate.Config) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	root, err := lab.FindRoot()
	if err != nil {
		return err
	}
	loadDotEnv(filepath.Join(root, ".env"))
	if cfg.Server == "" {
		cfg.Server = filepath.Join(root, "build", "mcp-server")
	}
	if cfg.Evidence == "" {
		cfg.Evidence = filepath.Join(filepath.Dir(filepath.Clean(out)), "evidence.json")
	}
	raw, err := os.ReadFile(alertPath)
	if err != nil {
		return err
	}
	alert, err := investigate.AlertJSON(raw)
	if err != nil {
		return err
	}
	_, err = investigate.Run(ctx, cfg, alert, out)
	return err
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
