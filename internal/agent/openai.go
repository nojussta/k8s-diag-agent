package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// Local default: apfel (https://apfel.franzai.com/) serves the on-device Apple
// Intelligence model on macOS 26+ through an OpenAI-compatible API. No key,
// no network, no per-token cost. The model has a 4096-token window shared by
// the prompt, the tool declarations and the answer, so this client fits every
// request into a budget (see fit).
const (
	DefaultLocalURL   = "http://127.0.0.1:11434"
	DefaultLocalModel = "apple-foundationmodel"
)

// OpenAIChat is a Model for any OpenAI-compatible /v1/chat/completions server.
// Local models cannot be trusted to combine tools and a response schema in one
// request (apfel then ignores the tools), so the client works in two phases:
//
//  1. investigation: tools, no schema; the model asks for tool calls;
//  2. answer: when the model stops calling tools, one more request without
//     tools and with the diagnosis schema as response_format.
type OpenAIChat struct {
	BaseURL string
	Model   string
	Token   string // optional bearer token
	HTTP    *http.Client
	System  string
	Tools   []ToolDef
	Schema  json.RawMessage

	// ContextTokens is the model's window (apfel: 4096); OutputReserve is kept
	// free for the answer. 0 disables fitting (large-context servers).
	ContextTokens int
	OutputReserve int
	// CharsPerToken is the estimate used by fit (JSON and logs are token-dense).
	CharsPerToken float64
	// Dump, when set, receives every request and response as one JSONL line (debugging).
	Dump io.Writer
	// Flatten sends every step as a fresh two-message conversation (system +
	// one user message holding the alert and all tool results so far) instead
	// of assistant/tool turns. apfel 1.16 fails with HTTP 500 when a request
	// that declares tools contains an earlier tool result and the model wants
	// to call another tool; a fresh conversation does not have that problem.
	Flatten bool
	// NextPrompt tells a flattened step what to do with the results so far.
	NextPrompt string
	// FinalPrompt is appended as a user message in the answer phase.
	FinalPrompt string
}

// NewLocal builds the client for an apfel server with the 4096-token budget.
func NewLocal(baseURL, model string, tools []ToolDef, schema json.RawMessage) *OpenAIChat {
	if baseURL == "" {
		baseURL = DefaultLocalURL
	}
	if model == "" {
		model = DefaultLocalModel
	}
	return &OpenAIChat{
		BaseURL: baseURL, Model: model, System: LocalSystemPrompt, Tools: tools, Schema: schema,
		ContextTokens: 4096, OutputReserve: 700, CharsPerToken: 2.6, FinalPrompt: LocalFinalPrompt,
		Flatten: true, NextPrompt: LocalNextPrompt,
	}
}

type oaiMessage struct {
	Role       string          `json:"role"`
	Content    *string         `json:"content"`
	ToolCalls  json.RawMessage `json:"tool_calls,omitempty"`
	ToolCallID string          `json:"tool_call_id,omitempty"`
}

func str(s string) *string { return &s }

// messages converts the orchestrator history into chat messages.
func (o *OpenAIChat) messages(history []Turn, final bool) ([]oaiMessage, error) {
	if o.Flatten {
		return o.flatten(history, final)
	}
	msgs := []oaiMessage{{Role: "system", Content: str(o.System)}}
	for _, t := range history {
		switch t.Role {
		case "user":
			msgs = append(msgs, oaiMessage{Role: "user", Content: str(t.Content)})
		case "model":
			var m oaiMessage
			if len(t.Raw) == 0 || json.Unmarshal(t.Raw, &m) != nil {
				m = oaiMessage{Role: "assistant", Content: str(t.Content)}
			}
			m.Role = "assistant"
			if final && len(m.ToolCalls) == 0 {
				continue // the draft text of the investigation phase is not evidence
			}
			msgs = append(msgs, m)
		case "tool":
			msgs = append(msgs, oaiMessage{Role: "tool", ToolCallID: t.CallID, Content: str(t.Content)})
		default:
			return nil, fmt.Errorf("unknown turn role %q", t.Role)
		}
	}
	if final {
		msgs = append(msgs, oaiMessage{Role: "user", Content: str(o.FinalPrompt)})
	}
	return msgs, nil
}

// noiseLabels are alert fields that cost tokens and never help a diagnosis.
var noiseLabels = map[string]bool{"uid": true, "prometheus": true, "service": true, "endpoint": true,
	"instance": true, "job": true, "lab": true, "generatorURL": true, "fingerprint": true,
	"endsAt": true, "runbook_url": true}

// compactAlert drops the noisy fields of an alert JSON (4096 tokens is little).
func compactAlert(s string) string {
	const head = "Alert (JSON):\n"
	if !strings.HasPrefix(s, head) {
		return s
	}
	var a map[string]any
	if json.Unmarshal([]byte(strings.TrimPrefix(s, head)), &a) != nil {
		return s
	}
	for k := range a {
		if noiseLabels[k] {
			delete(a, k)
		}
	}
	for _, sec := range []string{"labels", "annotations"} {
		if m, ok := a[sec].(map[string]any); ok {
			for k := range m {
				if noiseLabels[k] {
					delete(m, k)
				}
			}
		}
	}
	b, err := json.Marshal(a)
	if err != nil {
		return s
	}
	return head + string(b)
}

// flatten turns the history into one user message: the alert, the calls made
// so far with their (id-tagged) results, and what to do next. final asks for
// the diagnosis instead of the next call.
func (o *OpenAIChat) flatten(history []Turn, final bool) ([]oaiMessage, error) {
	var b strings.Builder
	pending := map[string]string{} // provider call id -> "name {args}" of the latest model turn
	first := true
	for _, t := range history {
		switch t.Role {
		case "user":
			if first {
				b.WriteString(compactAlert(t.Content))
				first = false
			} else {
				b.WriteString("\n\n" + t.Content)
			}
		case "model":
			pending = map[string]string{}
			var m oaiMessage
			if len(t.Raw) > 0 && json.Unmarshal(t.Raw, &m) == nil && len(m.ToolCalls) > 0 {
				var calls []struct {
					ID       string                           `json:"id"`
					Function struct{ Name, Arguments string } `json:"function"`
				}
				if json.Unmarshal(m.ToolCalls, &calls) == nil {
					for _, c := range calls {
						pending[c.ID] = c.Function.Name + " " + c.Function.Arguments
					}
				}
			}
		case "tool":
			call := pending[t.CallID]
			if call == "" {
				call = t.Name
			}
			b.WriteString("\n\nCalled: " + call + "\n" + t.Content)
		default:
			return nil, fmt.Errorf("unknown turn role %q", t.Role)
		}
	}
	if final {
		b.WriteString("\n\n" + o.FinalPrompt)
	} else if o.NextPrompt != "" {
		b.WriteString("\n\n" + o.NextPrompt)
	}
	return []oaiMessage{{Role: "system", Content: str(o.System)}, {Role: "user", Content: str(b.String())}}, nil
}

// compactParams drops the per-parameter descriptions: the names say enough and
// the 4096-token window pays for every character of the tool declarations.
func compactParams(p map[string]any) map[string]any {
	out := map[string]any{}
	for k, v := range p {
		out[k] = v
	}
	if props, ok := p["properties"].(map[string]any); ok {
		np := map[string]any{}
		for name, v := range props {
			if m, ok := v.(map[string]any); ok {
				c := map[string]any{}
				for k, x := range m {
					if k != "description" {
						c[k] = x
					}
				}
				np[name] = c
			} else {
				np[name] = v
			}
		}
		out["properties"] = np
	}
	return out
}

func (o *OpenAIChat) tools() []map[string]any {
	out := make([]map[string]any, 0, len(o.Tools))
	for _, d := range o.Tools {
		params := d.Parameters
		if o.ContextTokens > 0 {
			params = compactParams(params)
		}
		out = append(out, map[string]any{"type": "function", "function": map[string]any{
			"name": d.Name, "description": d.Description, "parameters": params}})
	}
	return out
}

// estimate is a token estimate for a JSON-serialisable value.
func (o *OpenAIChat) estimate(v any) int {
	b, _ := json.Marshal(v)
	return int(float64(len(b))/o.CharsPerToken) + 1
}

// fitHistory returns a copy of history whose tool outputs are shrunk (largest
// first) until the request estimate fits ContextTokens-OutputReserve. fixed is
// the cost of the tool declarations or of the response schema.
func (o *OpenAIChat) fitHistory(history []Turn, fixed int, final bool) ([]oaiMessage, error) {
	h := append([]Turn(nil), history...)
	build := func() ([]oaiMessage, int, error) {
		msgs, err := o.messages(h, final)
		if err != nil {
			return nil, 0, err
		}
		n := 0
		for _, m := range msgs {
			n += o.estimate(m) + 4
		}
		return msgs, n, nil
	}
	msgs, cost, err := build()
	if err != nil || o.ContextTokens == 0 {
		return msgs, err
	}
	budget := o.ContextTokens - o.OutputReserve - fixed
	// Shrink in two rounds: first down to 240 chars per tool output, then to
	// 80 if the request still does not fit. A tool output that cannot get any
	// shorter is skipped, so the loop always ends.
	for _, floor := range []int{240, 80} {
		stuck := map[int]bool{}
		for cost > budget {
			idx, size := -1, floor
			for i, t := range h {
				if t.Role == "tool" && !stuck[i] && len(t.Content) > size {
					idx, size = i, len(t.Content)
				}
			}
			if idx < 0 {
				break
			}
			// Cut at least 20% or what is needed, whichever is larger.
			cut := int(float64(cost-budget) * o.CharsPerToken)
			if min := size / 5; cut < min {
				cut = min
			}
			keep := size - cut
			if keep < floor {
				keep = floor
			}
			short := truncateTool(h[idx].Content, keep)
			if len(short) >= size {
				stuck[idx] = true // the marker text alone is as long as what is left
				continue
			}
			h[idx].Content = short
			if msgs, cost, err = build(); err != nil {
				return nil, err
			}
		}
	}
	return msgs, nil
}

// truncateTool keeps the head of a (possibly already wrapped) tool output.
func truncateTool(s string, keep int) string {
	if len(s) <= keep {
		return s
	}
	wrapped := strings.HasPrefix(s, "<tool_output")
	suffix := "\n...[truncated to fit the 4096-token context]"
	if wrapped {
		suffix += "\n</tool_output>"
		if i := strings.LastIndex(s, "\n</tool_output>"); i >= 0 {
			s = s[:i]
		}
		keep -= len("\n</tool_output>")
	}
	if keep < 0 {
		keep = 0
	}
	if keep > len(s) {
		keep = len(s)
	}
	return s[:keep] + suffix
}

// errContext marks the server's "input exceeds the context window" answer.
type errContext struct{ msg string }

func (e errContext) Error() string { return e.msg }

// withRoom retries a request with a tighter token estimate while the server
// says the input does not fit (the real tokenizer is not available here).
func (o *OpenAIChat) withRoom(do func() (Step, error)) (Step, error) {
	orig := o.CharsPerToken
	defer func() { o.CharsPerToken = orig }()
	for attempt := 0; ; attempt++ {
		st, err := do()
		var ec errContext
		if !errors.As(err, &ec) || attempt >= 4 {
			return st, err
		}
		o.CharsPerToken *= 0.8
	}
}

// Next implements Model.
func (o *OpenAIChat) Next(ctx context.Context, history []Turn) (Step, error) {
	step, err := o.withRoom(func() (Step, error) {
		toolDecl := o.tools()
		msgs, err := o.fitHistory(history, o.estimate(toolDecl), false)
		if err != nil {
			return Step{}, err
		}
		return o.chat(ctx, map[string]any{"messages": msgs, "tools": toolDecl})
	})
	if err != nil || len(step.Calls) > 0 || len(o.Schema) == 0 {
		return step, err
	}
	// Phase 2: the model stopped calling tools; ask for the schema-bound answer.
	final, err := o.Final(ctx, history)
	final.InputTokens += step.InputTokens
	final.OutputTokens += step.OutputTokens
	return final, err
}

var toolOutputID = regexp.MustCompile(`<tool_output id="(t\d+)"`)

// constrainSchema tightens the diagnosis schema for constrained decoding, so
// that the small model cannot cite a tool call that never ran: every evidence
// item must use one of the tool-output ids that appear in the history. (apfel
// cannot enforce an enum on a boolean, so "requires_human_approval must be
// true" stays a prompt rule plus the orchestrator's validation.) The
// orchestrator still validates the whole answer.
func constrainSchema(raw json.RawMessage, history []Turn) (any, error) {
	var schema map[string]any
	if err := json.Unmarshal(raw, &schema); err != nil {
		return nil, err
	}
	var ids []any
	seen := map[string]bool{}
	for _, t := range history {
		if t.Role != "tool" {
			continue
		}
		if m := toolOutputID.FindStringSubmatch(t.Content); m != nil && !seen[m[1]] && !strings.Contains(t.Content, `"duplicate call`) {
			seen[m[1]] = true
			ids = append(ids, m[1])
		}
	}
	dig := func(m map[string]any, path ...string) map[string]any {
		for _, k := range path {
			next, ok := m[k].(map[string]any)
			if !ok {
				return nil
			}
			m = next
		}
		return m
	}
	if p := dig(schema, "properties", "evidence", "items", "properties"); p != nil && len(ids) > 0 {
		p["tool_call"] = map[string]any{"type": "string", "enum": ids}
	}
	return schema, nil
}

// Final implements Finalizer: one request without tools, with the diagnosis
// schema as response_format.
func (o *OpenAIChat) Final(ctx context.Context, history []Turn) (Step, error) {
	schema, err := constrainSchema(o.Schema, history)
	if err != nil {
		return Step{}, err
	}
	rf := map[string]any{"type": "json_schema", "json_schema": map[string]any{"name": "diagnosis", "schema": schema}}
	return o.withRoom(func() (Step, error) {
		msgs, err := o.fitHistory(history, o.estimate(rf), true)
		if err != nil {
			return Step{}, err
		}
		return o.chat(ctx, map[string]any{"messages": msgs, "response_format": rf})
	})
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content   *string         `json:"content"`
			ToolCalls json.RawMessage `json:"tool_calls"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		Prompt     int `json:"prompt_tokens"`
		Completion int `json:"completion_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

func (o *OpenAIChat) chat(ctx context.Context, body map[string]any) (Step, error) {
	body["model"] = o.Model
	body["temperature"] = 0
	b, err := json.Marshal(body)
	if err != nil {
		return Step{}, err
	}
	client := o.HTTP
	if client == nil {
		client = &http.Client{Timeout: 120 * time.Second}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(o.BaseURL, "/")+"/v1/chat/completions", bytes.NewReader(b))
	if err != nil {
		return Step{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	if o.Token != "" {
		req.Header.Set("Authorization", "Bearer "+o.Token)
	}
	resp, err := client.Do(req)
	if err != nil {
		return Step{}, fmt.Errorf("local model: request failed: %v (is `apfel --serve` running?)", err)
	}
	data, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	resp.Body.Close()
	if o.Dump != nil {
		fmt.Fprintf(o.Dump, "{\"request\":%s,\"status\":%d,\"response\":%q}\n", b, resp.StatusCode, data)
	}
	var cr chatResponse
	if json.Unmarshal(data, &cr) != nil {
		return Step{}, fmt.Errorf("local model: HTTP %d, unreadable body: %s", resp.StatusCode, short(string(data)))
	}
	if cr.Error != nil {
		if cr.Error.Type == "context_length_exceeded" {
			return Step{}, errContext{fmt.Sprintf("local model: HTTP %d: %s", resp.StatusCode, cr.Error.Message)}
		}
		return Step{}, fmt.Errorf("local model: HTTP %d: %s", resp.StatusCode, cr.Error.Message)
	}
	if resp.StatusCode != http.StatusOK || len(cr.Choices) == 0 {
		return Step{}, fmt.Errorf("local model: HTTP %d: %s", resp.StatusCode, short(string(data)))
	}
	msg := cr.Choices[0].Message
	st := Step{InputTokens: cr.Usage.Prompt, OutputTokens: cr.Usage.Completion}
	if msg.Content != nil {
		st.Text = *msg.Content
	}
	raw := oaiMessage{Role: "assistant", Content: msg.Content, ToolCalls: msg.ToolCalls}
	st.Raw, _ = json.Marshal(raw)
	if len(msg.ToolCalls) > 0 && string(msg.ToolCalls) != "null" {
		var calls []struct {
			ID       string `json:"id"`
			Function struct {
				Name      string `json:"name"`
				Arguments string `json:"arguments"`
			} `json:"function"`
		}
		if err := json.Unmarshal(msg.ToolCalls, &calls); err != nil {
			return Step{}, fmt.Errorf("local model: bad tool_calls: %v", err)
		}
		for _, c := range calls {
			args := map[string]any{}
			if c.Function.Arguments != "" {
				if err := json.Unmarshal([]byte(c.Function.Arguments), &args); err != nil {
					args = map[string]any{} // invalid arguments are rejected by the tool's own validation
				}
			}
			st.Calls = append(st.Calls, Call{ID: c.ID, Name: c.Function.Name, Args: args})
		}
	}
	return st, nil
}
