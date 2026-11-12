package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

// Pipeline is the "fixed sequence + one model call" method: the read-only
// action sequence A1-A11 (internal/evidence) has already collected the data,
// and the model is asked once to combine it into a diagnosis. It is the
// workflow counterpart of Investigate (where the model chooses the tools), so
// the two can be compared on the same faults.
type Pipeline struct {
	Model     Model
	ModelName string
	// ToolTurns passes every action as a separate tool-result turn (the local
	// client trims the largest ones to fit its 4096-token window); otherwise
	// all actions go into one user message (hosted models with a large window).
	ToolTurns bool
	// MaxActionBytes caps one action's output (0 = no cap).
	MaxActionBytes int
	// Compact turns every action into a short digest before the call (the
	// on-device model's 4096-token window cannot hold the full ~20 KB of
	// evidence; without a digest the client trims every block to its header).
	Compact bool
	Trace   func(event string, v any)
}

// fixedEvidence is the part of evidence.json the pipeline reads.
type fixedEvidence struct {
	TotalCalls     int     `json:"total_calls"`
	TotalDurationS float64 `json:"total_duration_s"`
	Actions        []struct {
		ID     string          `json:"id"`
		Name   string          `json:"name"`
		Source string          `json:"source"`
		Output json.RawMessage `json:"output"`
	} `json:"actions"`
}

// Diagnose runs the single model call for one alert and its evidence.json.
func (p *Pipeline) Diagnose(ctx context.Context, alert string, evidenceJSON []byte) (Report, error) {
	start := time.Now()
	var ev fixedEvidence
	if err := json.Unmarshal(evidenceJSON, &ev); err != nil {
		return Report{}, fmt.Errorf("evidence.json: %w", err)
	}
	m := Metrics{Model: p.ModelName, ToolCalls: ev.TotalCalls, Mode: "pipeline"}
	trace := func(e string, v any) {
		if p.Trace != nil {
			p.Trace(e, v)
		}
	}
	finish := func(reason string, d *Diagnosis, err error) (Report, error) {
		m.StopReason = reason
		// The collection already happened before the model call; both count.
		m.DurationS = ev.TotalDurationS + time.Since(start).Seconds()
		if w, ok := p.Model.(Waiter); ok {
			m.RateLimitWaitS = w.WaitedSeconds()
		}
		trace("end", m)
		return Report{Diagnosis: d, Metrics: m}, err
	}

	executed := map[string]bool{}
	history := []Turn{{Role: "user", Content: "Alert (JSON):\n" + alert}}
	var all strings.Builder
	all.WriteString("Alert (JSON):\n" + alert + "\n\nThe fixed read-only action sequence has already collected this data:")
	ids := make([]string, 0, len(ev.Actions))
	actions := ev.Actions
	if p.Compact {
		actions = compactActions(actions)
	}
	for _, a := range actions {
		out := string(a.Output)
		if p.MaxActionBytes > 0 && len(out) > p.MaxActionBytes {
			out = out[:p.MaxActionBytes] + "...[truncated]"
		}
		executed[a.ID] = true
		ids = append(ids, a.ID)
		block := untrusted(a.ID, fmt.Sprintf("%s (%s):\n%s", a.Name, a.Source, out))
		if p.ToolTurns {
			history = append(history, Turn{Role: "tool", Name: a.ID + " " + a.Name, Content: block})
		} else {
			all.WriteString("\n\n" + block)
		}
	}
	if !p.ToolTurns {
		history = []Turn{{Role: "user", Content: all.String()}}
	}
	sort.Strings(ids)
	trace("pipeline_input", map[string]any{"actions": ids, "collection_calls": ev.TotalCalls})

	fin, canFinalize := p.Model.(Finalizer)
	for {
		var step Step
		var err error
		if canFinalize {
			step, err = fin.Final(ctx, history)
		} else {
			step, err = p.Model.Next(ctx, history)
		}
		m.ModelSteps++
		if err != nil {
			reason := "model_error"
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				reason = "timeout"
			}
			return finish(reason, nil, err)
		}
		m.InputTokens += step.InputTokens
		m.OutputTokens += step.OutputTokens
		if len(step.Calls) > 0 {
			return finish("unexpected_tool_call", nil, errors.New("the model asked for a tool in pipeline mode"))
		}
		d, verr := ParseDiagnosis([]byte(step.Text), executed)
		if verr == nil {
			trace("final", d)
			return finish("diagnosis", d, nil)
		}
		trace("invalid_diagnosis", verr.Error())
		if m.Repairs >= maxRepairs {
			return finish("invalid_diagnosis", nil, verr)
		}
		m.Repairs++
		history = append(history, Turn{Role: "model", Raw: step.Raw, Content: step.Text},
			Turn{Role: "user", Content: "Your answer was rejected: " + verr.Error() + ". Reply again with only the corrected JSON object."})
	}
}

// digestBudget is the character budget of each action in a compact digest:
// the actions that usually decide the cause (pod status, rollout change,
// events, logs) keep more.
var digestBudget = map[string]int{"A2": 360, "A5": 380, "A3": 240, "A4": 160, "A6": 120, "A7": 140,
	"A8": 80, "A9": 100, "A10": 140, "A11": 120, "A1": 160}

// digestDrop removes fields of one action that never decide the cause, so the
// decisive part (container states, the template diff) fits the budget.
var digestDrop = map[string][]string{
	"A2": {"conditions", "node"},
	"A5": {"conditions", "replicas"},
	"A1": {"startsAt", "summary"},
}

// digestOrder puts the decisive actions first, so trimming cuts the rest.
var digestOrder = []string{"A1", "A2", "A5", "A3", "A4", "A6", "A7", "A11", "A8", "A10", "A9"}

type fixedAction = struct {
	ID     string          `json:"id"`
	Name   string          `json:"name"`
	Source string          `json:"source"`
	Output json.RawMessage `json:"output"`
}

// compactActions shortens every action: lists keep their last items, noisy
// fields (stack traces, timestamps of every line) are cut by the budget.
func compactActions(in []fixedAction) []fixedAction {
	byID := map[string]fixedAction{}
	for _, a := range in {
		byID[a.ID] = a
	}
	var out []fixedAction
	for _, id := range digestOrder {
		a, ok := byID[id]
		if !ok {
			continue
		}
		var v any
		if json.Unmarshal(a.Output, &v) == nil {
			if m, ok := v.(map[string]any); ok {
				for _, k := range digestDrop[id] {
					delete(m, k)
				}
				if id == "A5" {
					// newest revision and what it changed; older revisions are noise
					if revs, ok := m["revisions"].([]any); ok && len(revs) > 0 {
						m["revisions"] = revs[:1]
					}
				}
			}
			v = shrink(v)
		}
		b, _ := json.Marshal(v)
		s := string(b)
		if n := digestBudget[id]; n > 0 && len(s) > n {
			s = s[:n] + "...[cut]"
		}
		a.Output = json.RawMessage(s) // used as text only; a cut may leave it unbalanced
		out = append(out, a)
	}
	return out
}

// noisyKeys are labels and fields that cost tokens and never decide a cause
// (Prometheus target labels, ids, stack traces).
var noisyKeys = map[string]bool{"__name__": true, "endpoint": true, "instance": true, "job": true, "namespace": true,
	"service": true, "uid": true, "node": true, "unit": true, "stacktrace": true, "created": true, "name": false}

// shrink keeps the last 3 items of every list, drops noisy keys and cuts long
// strings (log lines with stack traces).
func shrink(v any) any {
	switch t := v.(type) {
	case []any:
		if len(t) > 3 {
			t = t[len(t)-3:]
		}
		for i := range t {
			t[i] = shrink(t[i])
		}
		return t
	case map[string]any:
		for k, x := range t {
			if noisyKeys[k] {
				delete(t, k)
				continue
			}
			t[k] = shrink(x)
		}
		return t
	case string:
		if i := strings.Index(t, `"stacktrace"`); i > 0 {
			t = t[:i]
		}
		if len(t) > 220 {
			t = t[:220] + "..."
		}
		return t
	}
	return v
}
