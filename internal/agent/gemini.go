package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// DefaultGeminiModel is the prototype's default (configurable with GEMINI_MODEL
// or --model; every run records the exact model id in its metrics).
const DefaultGeminiModel = "gemini-3.5-flash"

// ToolDef is a function declaration derived from an MCP tool.
type ToolDef struct {
	Name        string
	Description string
	Parameters  map[string]any
}

// Gemini is a Model that talks to the Gemini API (Interactions endpoint) in
// stateless mode (store=false): every request carries the whole history, and
// the model's own steps (including thought signatures) are echoed back
// verbatim, as the API requires for multi-step function calling.
type Gemini struct {
	Key     string
	Model   string
	BaseURL string // default https://generativelanguage.googleapis.com
	HTTP    *http.Client
	System  string
	Tools   []ToolDef
	Schema  json.RawMessage // JSON schema of the final answer

	Retries int                 // extra attempts on 429/5xx (default 4)
	Sleep   func(time.Duration) // for tests
	Backoff func(attempt int) time.Duration

	// MinInterval spaces requests (free tier: 15 requests per minute, so about
	// 4 s); the pause counts as rate-limit waiting.
	MinInterval time.Duration

	waited time.Duration // time spent waiting on per-minute rate limits
	last   time.Time
}

// WaitedSeconds implements Waiter: the time spent waiting on rate limits,
// which the orchestrator reports separately from the diagnosis time.
func (g *Gemini) WaitedSeconds() float64 { return g.waited.Seconds() }

// finalInstruction ends the tool phase when the orchestrator forces an answer.
const finalInstruction = "The investigation is over: no more tool calls are possible. Give the final diagnosis now as one JSON object that matches the schema, using only the tool results above."

// Final implements Finalizer: the same conversation without tools and with an
// instruction to answer, so a model that keeps calling tools still ends with
// a diagnosis (validated like any other).
func (g *Gemini) Final(ctx context.Context, history []Turn) (Step, error) {
	h := append(append([]Turn(nil), history...), Turn{Role: "user", Content: finalInstruction})
	body, err := g.requestWith(h, false)
	if err != nil {
		return Step{}, err
	}
	raw, err := g.post(ctx, body)
	if err != nil {
		return Step{}, err
	}
	return parseInteraction(raw)
}

// NewGemini builds a client with the prototype's defaults.
func NewGemini(key, model string, tools []ToolDef, schema json.RawMessage) *Gemini {
	if model == "" {
		model = DefaultGeminiModel
	}
	return &Gemini{Key: key, Model: model, System: SystemPrompt, Tools: tools, Schema: schema}
}

func (g *Gemini) endpoint() string {
	base := g.BaseURL
	if base == "" {
		base = "https://generativelanguage.googleapis.com"
	}
	return strings.TrimRight(base, "/") + "/v1beta/interactions"
}

// request builds the JSON body for one turn of the loop.
func (g *Gemini) request(history []Turn) ([]byte, error) { return g.requestWith(history, true) }

func (g *Gemini) requestWith(history []Turn, withTools bool) ([]byte, error) {
	input := []json.RawMessage{}
	add := func(v any) {
		b, _ := json.Marshal(v)
		input = append(input, b)
	}
	for _, t := range history {
		switch t.Role {
		case "user":
			add(map[string]any{"type": "user_input", "content": t.Content})
		case "model":
			var steps []json.RawMessage
			if len(t.Raw) > 0 && json.Unmarshal(t.Raw, &steps) == nil {
				input = append(input, steps...)
			}
		case "tool":
			add(map[string]any{"type": "function_result", "name": t.Name, "call_id": t.CallID,
				"result": []map[string]string{{"type": "text", "text": t.Content}}})
		default:
			return nil, fmt.Errorf("unknown turn role %q", t.Role)
		}
	}
	tools := make([]map[string]any, 0, len(g.Tools))
	for _, d := range g.Tools {
		if !withTools {
			break
		}
		tools = append(tools, map[string]any{"type": "function", "name": d.Name, "description": d.Description, "parameters": d.Parameters})
	}
	body := map[string]any{
		"model":              g.Model,
		"store":              false,
		"system_instruction": g.System,
		"input":              input,
	}
	if len(tools) > 0 {
		body["tools"] = tools
	}
	if len(g.Schema) > 0 {
		var schema any
		if err := json.Unmarshal(g.Schema, &schema); err != nil {
			return nil, err
		}
		body["response_format"] = map[string]any{"type": "text", "mime_type": "application/json", "schema": schema}
	}
	return json.Marshal(body)
}

type interaction struct {
	Steps []json.RawMessage `json:"steps"`
	Usage map[string]any    `json:"usage"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// Next implements Model.
func (g *Gemini) Next(ctx context.Context, history []Turn) (Step, error) {
	body, err := g.request(history)
	if err != nil {
		return Step{}, err
	}
	raw, err := g.post(ctx, body)
	if err != nil {
		return Step{}, err
	}
	return parseInteraction(raw)
}

func parseInteraction(raw []byte) (Step, error) {
	var it interaction
	if err := json.Unmarshal(raw, &it); err != nil {
		return Step{}, fmt.Errorf("gemini: bad response: %v", err)
	}
	if it.Error != nil {
		return Step{}, fmt.Errorf("gemini: %s", it.Error.Message)
	}
	st := Step{}
	rawSteps, _ := json.Marshal(it.Steps)
	st.Raw = rawSteps
	for _, s := range it.Steps {
		var h struct {
			Type      string          `json:"type"`
			ID        string          `json:"id"`
			Name      string          `json:"name"`
			Arguments map[string]any  `json:"arguments"`
			Content   json.RawMessage `json:"content"`
		}
		if json.Unmarshal(s, &h) != nil {
			continue
		}
		switch h.Type {
		case "function_call":
			st.Calls = append(st.Calls, Call{ID: h.ID, Name: h.Name, Args: h.Arguments})
		case "model_output":
			st.Text += contentText(h.Content)
		}
	}
	st.InputTokens, st.OutputTokens = usageTokens(it.Usage)
	return st, nil
}

// contentText accepts "content" as a string or as a list of {text} parts.
func contentText(c json.RawMessage) string {
	var s string
	if json.Unmarshal(c, &s) == nil {
		return s
	}
	var parts []struct {
		Text string `json:"text"`
	}
	if json.Unmarshal(c, &parts) == nil {
		var b strings.Builder
		for _, p := range parts {
			b.WriteString(p.Text)
		}
		return b.String()
	}
	return ""
}

// usageTokens reads the Interactions API usage block: total_input_tokens, and
// total_output_tokens plus total_thought_tokens (thinking is billed as output).
func usageTokens(u map[string]any) (in, out int) {
	num := func(k string) int { f, _ := u[k].(float64); return int(f) }
	return num("total_input_tokens"), num("total_output_tokens") + num("total_thought_tokens")
}

func (g *Gemini) post(ctx context.Context, body []byte) ([]byte, error) {
	if g.Key == "" {
		return nil, fmt.Errorf("gemini: GEMINI_API_KEY is not set")
	}
	client := g.HTTP
	if client == nil {
		client = &http.Client{Timeout: 60 * time.Second} // a hung request is retried instead of eating the whole 180 s budget
	}
	retries := g.Retries
	if retries == 0 {
		retries = 4
	}
	if g.MinInterval > 0 && !g.last.IsZero() {
		if d := g.MinInterval - time.Since(g.last); d > 0 {
			g.waited += d
			select {
			case <-time.After(d):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
	}
	defer func() { g.last = time.Now() }()
	var last error
	var hint time.Duration
	for attempt := 0; attempt <= retries; attempt++ {
		if attempt > 0 {
			d := time.Duration(1<<attempt) * time.Second
			if g.Backoff != nil {
				d = g.Backoff(attempt)
			}
			if hint > d {
				d = hint // the server says when the per-minute window frees up
			}
			g.waited += d
			if g.Sleep != nil {
				g.Sleep(d)
			} else {
				select {
				case <-time.After(d):
				case <-ctx.Done():
					return nil, ctx.Err()
				}
			}
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.endpoint(), bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("x-goog-api-key", g.Key) // header only: never in the URL, never logged
		resp, err := client.Do(req)
		if err != nil {
			last = fmt.Errorf("gemini: request failed: %v", scrub(err.Error(), g.Key))
			continue
		}
		data, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
		resp.Body.Close()
		switch {
		case resp.StatusCode == http.StatusOK:
			return data, nil
		case resp.StatusCode == http.StatusTooManyRequests && dailyQuota(string(data)):
			// Retrying cannot help: the free tier is used up until tomorrow.
			return nil, fmt.Errorf("gemini: daily quota exhausted: %s", scrub(short(string(data)), g.Key))
		case resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500:
			hint = retryHint(string(data))
			last = fmt.Errorf("gemini: HTTP %d (will retry): %s", resp.StatusCode, scrub(short(string(data)), g.Key))
		default:
			return nil, fmt.Errorf("gemini: HTTP %d: %s", resp.StatusCode, scrub(short(string(data)), g.Key))
		}
	}
	return nil, last
}

func short(s string) string {
	if len(s) > 300 {
		return s[:300] + "..."
	}
	return s
}

func scrub(s, key string) string {
	if key == "" {
		return s
	}
	return strings.ReplaceAll(s, key, "***")
}

// dailyQuota recognises "limit: N requests per day" and retry hints of an hour or more.
func dailyQuota(body string) bool {
	b := strings.ToLower(body)
	return strings.Contains(b, "per day") || strings.Contains(b, "retry in") && (strings.Contains(b, "h") && strings.Contains(b, "m") && strings.Contains(b, "in ") && regexpHours.MatchString(b))
}

var regexpHours = regexp.MustCompile(`retry in \d+h`)

var retryIn = regexp.MustCompile(`retry in (\d+(?:\.\d+)?)s`)

// retryHint reads "Please retry in 43s" (per-minute limits); at most 65 s.
func retryHint(body string) time.Duration {
	m := retryIn.FindStringSubmatch(strings.ToLower(body))
	if m == nil {
		return 0
	}
	var sec float64
	fmt.Sscanf(m[1], "%g", &sec)
	if sec <= 0 || sec > 65 {
		return 0
	}
	return time.Duration((sec + 1) * float64(time.Second))
}
