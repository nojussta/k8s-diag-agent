// Command agent investigates one alert: it starts the MCP server, gives the
// model the read-only tools, runs the investigation loop and writes
// diagnosis.json, metrics.json and trace.jsonl. Used by labctl:
//
//	labctl scenario s1-oomkilled --agent-cmd "build/agent --alert {alert} --out {out}"
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"

	"github.com/nojussta/k8s-diag-agent/docs"
	"github.com/nojussta/k8s-diag-agent/internal/agent"
	"github.com/nojussta/k8s-diag-agent/internal/alertmsg"
	"github.com/nojussta/k8s-diag-agent/internal/lab"
	"github.com/nojussta/k8s-diag-agent/internal/tools"
)

func main() {
	alertPath := flag.String("alert", "", "alert JSON (a single Alertmanager alert, or a webhook payload)")
	out := flag.String("out", ".", "output directory")
	server := flag.String("server", "", "mcp-server binary (default: <root>/build/mcp-server)")
	model := flag.String("model", "", "model id (default: $GEMINI_MODEL / "+agent.DefaultGeminiModel+" for gemini, "+agent.DefaultLocalModel+" for apfel)")
	provider := flag.String("provider", "", "auto | apfel | gemini (default: $LLM_PROVIDER or auto: local apfel when installed, else Gemini)")
	flag.Parse()
	if *alertPath == "" {
		fmt.Fprintln(os.Stderr, "usage: agent --alert FILE [--out DIR] [--model ID]")
		os.Exit(2)
	}
	if err := run(*alertPath, *out, *server, *model, *provider); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(alertPath, out, server, model, provider string) error {
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
	if provider != "apfel" && provider != "gemini" {
		return fmt.Errorf("unknown provider %q (use auto, apfel or gemini)", provider)
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
	} else if model == "" {
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

	mcpCmd := exec.Command(server)
	mcpCmd.Env = os.Environ()
	client, err := agent.Connect(ctx, mcpCmd)
	if err != nil {
		return err
	}
	defer client.Close()

	allowed := map[string]bool{}
	for _, n := range tools.Names {
		allowed[n] = true
	}
	defs, err := client.Defs(ctx, allowed)
	if err != nil {
		return err
	}

	traceFile, err := os.Create(filepath.Join(out, "trace.jsonl"))
	if err != nil {
		return err
	}
	defer traceFile.Close()

	var llm agent.Model
	limits := agent.Limits{}
	if provider == "apfel" {
		url, stopApfel, err := agent.StartApfel(ctx, os.Getenv("APFEL_URL"))
		if err != nil {
			return err
		}
		defer stopApfel()
		local := agent.NewLocal(url, model, defs, docs.DiagnosisSchema)
		if d := os.Getenv("LOCAL_LLM_DUMP"); d != "" {
			if f, err := os.Create(d); err == nil {
				defer f.Close()
				local.Dump = f
			}
		}
		llm = local
		limits = agent.LocalLimits
	} else {
		llm = agent.NewGemini(key, model, defs, docs.DiagnosisSchema)
	}
	fmt.Printf("provider=%s model=%s\n", provider, model)

	a := &agent.Agent{
		Model:     llm,
		Limits:    limits,
		ModelName: model,
		Tools:     client,
		Allowed:   allowed,
		Trace:     traceFile,
	}
	rep, runErr := a.Investigate(ctx, alert)

	if err := writeJSON(filepath.Join(out, "metrics.json"), rep.Metrics); err != nil {
		return err
	}
	if rep.Diagnosis != nil {
		if err := writeJSON(filepath.Join(out, "diagnosis.json"), rep.Diagnosis); err != nil {
			return err
		}
		d := rep.Diagnosis
		fmt.Printf("%s | %s | confidence %s\n%s\n", d.RootCause.Category, d.RootCause.Component, d.Confidence, d.Summary)
	}
	fmt.Printf("model=%s tool_calls=%d rejected=%d repairs=%d duration=%.1fs tokens in/out=%d/%d stop=%s\n",
		rep.Metrics.Model, rep.Metrics.ToolCalls, rep.Metrics.Rejected, rep.Metrics.Repairs, rep.Metrics.DurationS,
		rep.Metrics.InputTokens, rep.Metrics.OutputTokens, rep.Metrics.StopReason)
	return runErr
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
