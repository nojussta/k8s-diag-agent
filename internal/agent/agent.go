// Package agent is the orchestrator: it runs the investigation loop between a
// language model and the MCP tools. Security level 3 of the design lives here:
// an allow-list of tools, a call and time budget, a duplicate-call filter and
// a rule that tool output is passed on as untrusted data. The final answer is
// validated (schema, enums, grounding of every evidence item in a real tool
// call, and "every suggested command needs human approval") before it is
// accepted.
package agent

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"
)

// Limits: tool calls and wall-clock time per investigation.
type Limits struct {
	MaxCalls int
	Timeout  time.Duration
}

// DefaultLimits are 15 tool calls and 180 s.
var DefaultLimits = Limits{MaxCalls: 15, Timeout: 180 * time.Second}

// Call is one tool call the model asks for.
type Call struct {
	ID   string // provider's call id, echoed back with the result
	Name string
	Args map[string]any
}

// Step is what the model returns: tool calls, or a final text (the diagnosis JSON).
type Step struct {
	Calls                     []Call
	Text                      string
	Raw                       json.RawMessage // provider-specific model steps, echoed back verbatim in the next request
	InputTokens, OutputTokens int
}

// Turn is one entry of the conversation passed back to the model.
type Turn struct {
	Role    string // "user", "model", "tool"
	Content string
	CallID  string          // tool: provider call id
	Name    string          // tool: tool name
	Raw     json.RawMessage // model: what the provider returned
}

// Model is any language model (the Gemini client, or a scripted fake in tests).
type Model interface {
	Next(ctx context.Context, history []Turn) (Step, error)
}

// Tools is the MCP client side.
type Tools interface {
	Call(ctx context.Context, name string, args map[string]any) (string, error)
}

// Metrics are filled by the orchestrator, never by the model.
type Metrics struct {
	ToolCalls    int     `json:"tool_calls"`
	Rejected     int     `json:"rejected_calls"`
	Repairs      int     `json:"repairs"`
	DurationS    float64 `json:"duration_s"`
	InputTokens  int     `json:"input_tokens"`
	OutputTokens int     `json:"output_tokens"`
	StopReason   string  `json:"stop_reason"`
	Model        string  `json:"model,omitempty"`
}

// Report is the result of one investigation.
type Report struct {
	Diagnosis *Diagnosis `json:"diagnosis,omitempty"`
	Metrics   Metrics    `json:"metrics"`
}

// Agent ties a model, a tool client and the guards together.
type Agent struct {
	Model     Model
	ModelName string
	Tools     Tools
	Allowed   map[string]bool
	Limits    Limits
	Trace     io.Writer // JSONL, one line per event; may be nil
}

func (a *Agent) trace(event string, v any) {
	if a.Trace == nil {
		return
	}
	b, _ := json.Marshal(map[string]any{"t": time.Now().UTC().Format(time.RFC3339Nano), "event": event, "data": v})
	fmt.Fprintf(a.Trace, "%s\n", b)
}

// maxRepairs is how many times an invalid final answer is sent back to the model.
const maxRepairs = 1

// Investigate runs the loop for one alert (alert is the raw JSON of the alert).
func (a *Agent) Investigate(ctx context.Context, alert string) (Report, error) {
	lim := a.Limits
	if lim.MaxCalls == 0 {
		lim = DefaultLimits
	}
	ctx, cancel := context.WithTimeout(ctx, lim.Timeout)
	defer cancel()

	start := time.Now()
	m := Metrics{Model: a.ModelName}
	finish := func(reason string, d *Diagnosis, err error) (Report, error) {
		m.StopReason = reason
		m.DurationS = time.Since(start).Seconds()
		a.trace("end", m)
		return Report{Diagnosis: d, Metrics: m}, err
	}

	history := []Turn{{Role: "user", Content: "Alert (JSON):\n" + alert}}
	seen := map[[32]byte]string{}
	executed := map[string]bool{}
	n := 0
	a.trace("alert", alert)

	for {
		step, err := a.Model.Next(ctx, history)
		if err != nil {
			reason := "model_error"
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				reason = "timeout"
			}
			return finish(reason, nil, err)
		}
		m.InputTokens += step.InputTokens
		m.OutputTokens += step.OutputTokens
		history = append(history, Turn{Role: "model", Raw: step.Raw, Content: step.Text})

		if len(step.Calls) == 0 {
			if step.Text == "" {
				return finish("empty_step", nil, errors.New("model returned neither tool calls nor a diagnosis"))
			}
			d, verr := ParseDiagnosis([]byte(step.Text), executed)
			if verr == nil {
				a.trace("final", d)
				return finish("diagnosis", d, nil)
			}
			a.trace("invalid_diagnosis", verr.Error())
			if m.Repairs >= maxRepairs {
				return finish("invalid_diagnosis", nil, verr)
			}
			m.Repairs++
			history = append(history, Turn{Role: "user", Content: "Your answer was rejected: " + verr.Error() +
				". Reply again with only the corrected JSON object."})
			continue
		}

		for _, c := range step.Calls {
			n++
			id := fmt.Sprintf("t%d", n)
			if m.ToolCalls >= lim.MaxCalls {
				return finish("call_budget", nil, fmt.Errorf("call budget of %d exhausted", lim.MaxCalls))
			}
			out := ""
			switch {
			case !a.Allowed[c.Name]:
				m.Rejected++
				out = fmt.Sprintf(`{"ok":false,"error":"tool %q is not allowed"}`, c.Name)
			default:
				key := sha256.Sum256([]byte(c.Name + "\x00" + canonical(c.Args)))
				if prev, dup := seen[key]; dup {
					m.Rejected++
					out = fmt.Sprintf(`{"ok":false,"error":"duplicate call, same as %s; use that result"}`, prev)
				} else {
					seen[key] = id
					m.ToolCalls++
					executed[id] = true
					res, err := a.Tools.Call(ctx, c.Name, c.Args)
					if err != nil {
						res = fmt.Sprintf(`{"ok":false,"error":%q}`, err.Error())
					}
					out = res
				}
			}
			a.trace("tool_call", map[string]any{"id": id, "name": c.Name, "args": c.Args, "result_bytes": len(out)})
			history = append(history, Turn{Role: "tool", CallID: c.ID, Name: c.Name, Content: untrusted(id, out)})
		}
	}
}

// untrusted marks tool output as data: logs and events are attacker-influenced
// text and must never be followed as instructions (OWASP LLM01). The id lets
// the model cite the call, and lets us check the citation.
func untrusted(id, s string) string {
	return fmt.Sprintf("<tool_output id=%q untrusted=\"true\">\n%s\n</tool_output>", id, s)
}

func canonical(args map[string]any) string {
	b, _ := json.Marshal(args) // encoding/json sorts map keys
	return string(b)
}
