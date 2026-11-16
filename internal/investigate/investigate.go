// Package investigate runs one diagnosis from start to finish: it picks the
// model client, starts the MCP server, runs the agent loop (or the pipeline),
// writes diagnosis.json, metrics.json, trace.jsonl and report.md, and posts the
// report to the chat webhook. The command-line agent (cmd/agent) and the
// in-cluster agent (cmd/diag-agent) both call Run, so an investigation started
// by hand and one started by an alert are the same code.
package investigate

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/nojussta/k8s-diag-agent/docs"
	"github.com/nojussta/k8s-diag-agent/internal/agent"
	"github.com/nojussta/k8s-diag-agent/internal/alertmsg"
	"github.com/nojussta/k8s-diag-agent/internal/notify"
	"github.com/nojussta/k8s-diag-agent/internal/tools"
)

// Config selects the method, the model and where the pieces are.
type Config struct {
	Provider string // auto | apfel | gemini | openai ("" = $LLM_PROVIDER, then auto)
	Model    string // "" = the provider's default
	Mode     string // agent | pipeline ("" = agent)
	Server   string // mcp-server binary (agent mode)
	Evidence string // pipeline mode: evidence.json of the fixed sequence
	Notify   string // chat webhook ("" = $NOTIFY_WEBHOOK_URL; empty = file only)
	Log      io.Writer
}

// Result is what Run reports besides the files it wrote.
type Result struct {
	Report   agent.Report
	Notified bool
	Text     string // the message for the engineer ("" when there is no diagnosis)
}

func (c Config) logf(format string, args ...any) {
	if c.Log != nil {
		fmt.Fprintf(c.Log, format+"\n", args...)
	}
}

// Resolve fills in the provider and the model from the environment and checks
// what the provider needs. It is separate from Run so a long-running service
// can fail at start-up instead of at the first alert.
func (c Config) Resolve() (Config, error) {
	if c.Mode == "" {
		c.Mode = "agent"
	}
	if c.Mode != "agent" && c.Mode != "pipeline" {
		return c, fmt.Errorf("unknown mode %q (use agent or pipeline)", c.Mode)
	}
	if c.Provider == "" {
		c.Provider = os.Getenv("LLM_PROVIDER")
	}
	if c.Provider == "" || c.Provider == "auto" {
		// Local first: it costs nothing and nothing leaves the machine.
		c.Provider = "gemini"
		if agent.ApfelAvailable(os.Getenv("APFEL_URL")) {
			c.Provider = "apfel"
		}
	}
	switch c.Provider {
	case "apfel":
		if c.Model == "" {
			c.Model = agent.DefaultLocalModel
		}
	case "openai":
		if os.Getenv("OPENAI_BASE_URL") == "" {
			return c, fmt.Errorf("provider openai needs OPENAI_BASE_URL (e.g. http://127.0.0.1:11434 for Ollama) and OPENAI_MODEL")
		}
		if c.Model == "" {
			c.Model = os.Getenv("OPENAI_MODEL")
		}
		if c.Model == "" {
			return c, fmt.Errorf("provider openai needs --model or OPENAI_MODEL")
		}
	case "gemini":
		if os.Getenv("GEMINI_API_KEY") == "" {
			return c, fmt.Errorf("GEMINI_API_KEY is not set (put it in .env, or use --provider apfel)")
		}
		if c.Model == "" {
			c.Model = os.Getenv("GEMINI_MODEL")
		}
		if c.Model == "" {
			c.Model = agent.DefaultGeminiModel
		}
	default:
		return c, fmt.Errorf("unknown provider %q (use auto, apfel, gemini or openai)", c.Provider)
	}
	return c, nil
}

// Run diagnoses one alert (alert is the JSON of a single Alertmanager alert)
// and writes the result files into out. A returned error means the
// investigation did not end with a diagnosis; the metrics are written anyway.
func Run(ctx context.Context, c Config, alert, out string) (Result, error) {
	var res Result
	c, err := c.Resolve()
	if err != nil {
		return res, err
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		return res, err
	}
	traceFile, err := os.Create(filepath.Join(out, "trace.jsonl"))
	if err != nil {
		return res, err
	}
	defer traceFile.Close()

	var defs []agent.ToolDef
	var client *agent.MCPTools
	allowed := map[string]bool{}
	if c.Mode == "agent" {
		mcpCmd := exec.Command(c.Server)
		mcpCmd.Env = os.Environ()
		if client, err = agent.Connect(ctx, mcpCmd); err != nil {
			return res, err
		}
		defer client.Close()
		for _, n := range tools.Names {
			allowed[n] = true
		}
		if defs, err = client.Defs(ctx, allowed); err != nil {
			return res, err
		}
	}

	var llm agent.Model
	limits := agent.Limits{}
	switch c.Provider {
	case "apfel":
		url, stopApfel, err := agent.StartApfel(ctx, os.Getenv("APFEL_URL"))
		if err != nil {
			return res, err
		}
		defer stopApfel()
		local := agent.NewLocal(url, c.Model, defs, docs.DiagnosisSchema)
		if c.Mode == "pipeline" {
			local.System = agent.LocalPipelinePrompt
		}
		if d := os.Getenv("LOCAL_LLM_DUMP"); d != "" {
			if f, err := os.Create(d); err == nil {
				defer f.Close()
				local.Dump = f
			}
		}
		llm, limits = local, agent.LocalLimits
	case "openai":
		window, _ := strconv.Atoi(os.Getenv("LLM_CONTEXT_TOKENS")) // 0 = no fitting
		oc := agent.NewOpenAI(os.Getenv("OPENAI_BASE_URL"), c.Model, os.Getenv("OPENAI_API_KEY"), defs, docs.DiagnosisSchema, window)
		if c.Mode == "pipeline" {
			oc.System = agent.PipelineSystemPrompt
		}
		llm = oc
	default:
		g := agent.NewGemini(os.Getenv("GEMINI_API_KEY"), c.Model, defs, docs.DiagnosisSchema)
		if d, err := time.ParseDuration(os.Getenv("GEMINI_MIN_INTERVAL")); err == nil {
			g.MinInterval = d // e.g. 4.2s on the free tier (15 requests per minute)
		}
		if t, err := strconv.ParseFloat(os.Getenv("GEMINI_TEMPERATURE"), 64); err == nil {
			g.Temperature = &t
		}
		if c.Mode == "pipeline" {
			g.System = agent.PipelineSystemPrompt
		} else if os.Getenv("AGENT_PROMPT") == "v1" {
			g.System = agent.SystemPromptV1 // ablation: the procedure without step f
		}
		llm = g
	}
	c.logf("mode=%s provider=%s model=%s", c.Mode, c.Provider, c.Model)

	var runErr error
	if c.Mode == "pipeline" {
		evRaw, err := os.ReadFile(c.Evidence)
		if err != nil {
			return res, fmt.Errorf("pipeline mode needs the fixed-sequence evidence: %w", err)
		}
		trace := func(event string, v any) { writeTrace(traceFile, event, v) }
		trace("alert", alert)
		p := &agent.Pipeline{Model: llm, ModelName: c.Model, ToolTurns: c.Provider == "apfel", Compact: c.Provider == "apfel",
			MaxActionBytes: 8 << 10, Trace: trace}
		timeout := limits.Timeout
		if timeout == 0 {
			timeout = agent.DefaultLimits.Timeout
		}
		pctx, cancel := context.WithTimeout(ctx, timeout)
		res.Report, runErr = p.Diagnose(pctx, alert, evRaw)
		cancel()
	} else {
		a := &agent.Agent{Model: llm, Limits: limits, ModelName: c.Model, Tools: client, Allowed: allowed, Trace: traceFile}
		res.Report, runErr = a.Investigate(ctx, alert)
	}

	rep := res.Report
	if err := writeJSON(filepath.Join(out, "metrics.json"), rep.Metrics); err != nil {
		return res, err
	}
	if d := rep.Diagnosis; d != nil {
		if err := writeJSON(filepath.Join(out, "diagnosis.json"), d); err != nil {
			return res, err
		}
		c.logf("%s | %s | confidence %s\n%s", d.RootCause.Category, d.RootCause.Component, d.Confidence, d.Summary)
		res.Text = notify.Text(alertName(alert), d, rep.Metrics)
		if err := os.WriteFile(filepath.Join(out, "report.md"), []byte(res.Text), 0o644); err != nil {
			return res, err
		}
		url := c.Notify
		if url == "" {
			url = os.Getenv("NOTIFY_WEBHOOK_URL")
		}
		if url != "" {
			if err := notify.Post(ctx, url, res.Text); err != nil {
				c.logf("notification failed (the report is still in report.md): %v", err)
			} else {
				res.Notified = true
				c.logf("notification delivered to the webhook")
			}
		}
	}
	m := rep.Metrics
	c.logf("mode=%s model=%s tool_calls=%d model_steps=%d rejected=%d repairs=%d duration=%.1fs tokens in/out=%d/%d stop=%s",
		m.Mode, m.Model, m.ToolCalls, m.ModelSteps, m.Rejected, m.Repairs, m.DurationS, m.InputTokens, m.OutputTokens, m.StopReason)
	return res, runErr
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

// AlertJSON accepts one alert or a webhook payload and returns one alert as JSON.
func AlertJSON(raw []byte) (string, error) {
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
