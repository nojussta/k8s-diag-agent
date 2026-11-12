package agent

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func fakeLocal(t *testing.T, handler func(n int, body map[string]any) (int, string)) *OpenAIChat {
	t.Helper()
	n := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Errorf("path %s", r.URL.Path)
		}
		b, _ := io.ReadAll(r.Body)
		var body map[string]any
		_ = json.Unmarshal(b, &body)
		n++
		code, resp := handler(n, body)
		w.WriteHeader(code)
		io.WriteString(w, resp)
	}))
	t.Cleanup(srv.Close)
	return NewLocal(srv.URL, "", []ToolDef{{Name: "list_pods", Description: "d", Parameters: map[string]any{"type": "object"}}},
		json.RawMessage(`{"type":"object"}`))
}

const localCall = `{"choices":[{"message":{"content":null,"tool_calls":[{"id":"call_1","type":"function","function":{"name":"list_pods","arguments":"{\"namespace\":\"demo\"}"}}]}}],"usage":{"prompt_tokens":210,"completion_tokens":20}}`

func TestLocalToolCallParsing(t *testing.T) {
	o := fakeLocal(t, func(_ int, body map[string]any) (int, string) {
		if body["model"] != DefaultLocalModel || body["temperature"] != float64(0) {
			t.Errorf("body: %v", body)
		}
		if _, ok := body["response_format"]; ok {
			t.Error("the tool phase must not send response_format (apfel would skip the tools)")
		}
		if len(body["tools"].([]any)) != 1 {
			t.Error("tools missing")
		}
		return 200, localCall
	})
	st, err := o.Next(context.Background(), []Turn{{Role: "user", Content: "alert"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(st.Calls) != 1 || st.Calls[0].Name != "list_pods" || st.Calls[0].Args["namespace"] != "demo" || st.Calls[0].ID != "call_1" {
		t.Fatalf("calls: %+v", st.Calls)
	}
	if st.InputTokens != 210 || st.OutputTokens != 20 {
		t.Errorf("usage: %d/%d", st.InputTokens, st.OutputTokens)
	}
}

func TestLocalTwoPhaseAnswer(t *testing.T) {
	o := fakeLocal(t, func(n int, body map[string]any) (int, string) {
		if n == 1 { // tool phase: the model answers with plain text and no calls
			return 200, `{"choices":[{"message":{"content":"I think it is OOM"}}],"usage":{"prompt_tokens":100,"completion_tokens":10}}`
		}
		if _, ok := body["tools"]; ok {
			t.Error("the answer phase must not send tools")
		}
		if _, ok := body["response_format"]; !ok {
			t.Error("the answer phase must send response_format")
		}
		msgs := body["messages"].([]any)
		last := msgs[len(msgs)-1].(map[string]any)
		if last["role"] != "user" || !strings.Contains(last["content"].(string), "final diagnosis") {
			t.Errorf("last message: %v", last)
		}
		return 200, `{"choices":[{"message":{"content":"{\"summary\":\"x\"}"}}],"usage":{"prompt_tokens":300,"completion_tokens":50}}`
	})
	st, err := o.Next(context.Background(), []Turn{{Role: "user", Content: "alert"}})
	if err != nil {
		t.Fatal(err)
	}
	if st.Text != `{"summary":"x"}` || len(st.Calls) != 0 {
		t.Fatalf("step: %+v", st)
	}
	if st.InputTokens != 400 || st.OutputTokens != 60 {
		t.Errorf("usage must add both phases: %d/%d", st.InputTokens, st.OutputTokens)
	}
}

func TestChatHistoryRoundTrip(t *testing.T) { // the non-flattened mode, for servers without the apfel quirk
	o := fakeLocal(t, func(_ int, body map[string]any) (int, string) {
		msgs := body["messages"].([]any)
		if len(msgs) == 2 {
			return 200, localCall
		}
		roles := []string{}
		for _, m := range msgs {
			roles = append(roles, m.(map[string]any)["role"].(string))
		}
		if strings.Join(roles, ",") != "system,user,assistant,tool" {
			t.Errorf("roles: %v", roles)
		}
		if msgs[2].(map[string]any)["tool_calls"] == nil {
			t.Error("assistant tool_calls were not echoed back")
		}
		if msgs[3].(map[string]any)["tool_call_id"] != "call_1" {
			t.Errorf("tool message: %v", msgs[3])
		}
		return 200, localCall
	})
	o.Flatten = false
	first, err := o.Next(context.Background(), []Turn{{Role: "user", Content: "alert"}})
	if err != nil {
		t.Fatal(err)
	}
	_, err = o.Next(context.Background(), []Turn{
		{Role: "user", Content: "alert"},
		{Role: "model", Raw: first.Raw},
		{Role: "tool", CallID: "call_1", Name: "list_pods", Content: "<tool_output id=\"t1\">x</tool_output>"},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestLocalFitShrinksLargestToolOutputFirst(t *testing.T) {
	o := NewLocal("http://x", "", nil, nil)
	big := "<tool_output id=\"t1\" untrusted=\"true\">\n" + strings.Repeat("log line\n", 1500) + "\n</tool_output>"
	small := "<tool_output id=\"t2\" untrusted=\"true\">\nshort\n</tool_output>"
	hist := []Turn{
		{Role: "user", Content: "alert"},
		{Role: "tool", CallID: "c1", Name: "a", Content: big},
		{Role: "tool", CallID: "c2", Name: "b", Content: small},
	}
	msgs, err := o.fitHistory(hist, 600, false)
	if err != nil {
		t.Fatal(err)
	}
	text := *msgs[1].Content
	if got := o.estimate(msgs[0]) + o.estimate(msgs[1]); got > o.ContextTokens-o.OutputReserve-600 {
		t.Errorf("still over budget: %d", got)
	}
	if !strings.Contains(text, small) {
		t.Error("the small output must be left alone")
	}
	if !strings.Contains(text, "truncated") || strings.Count(text, "</tool_output>") != 2 {
		t.Errorf("truncation marker or closing tag missing")
	}
	if hist[1].Content != big {
		t.Error("fitHistory must not modify the caller's history")
	}
}

func TestFlattenKeepsAlertResultsAndIds(t *testing.T) {
	o := NewLocal("http://x", "", nil, nil)
	raw := json.RawMessage(`{"role":"assistant","content":null,"tool_calls":[{"id":"call_1","type":"function","function":{"name":"get_pod_status","arguments":"{\"pod\":\"a\"}"}}]}`)
	msgs, err := o.flatten([]Turn{
		{Role: "user", Content: "Alert (JSON):\n{\"labels\":{\"pod\":\"a\",\"uid\":\"zzz\"},\"generatorURL\":\"http://x\"}"},
		{Role: "model", Raw: raw},
		{Role: "tool", CallID: "call_1", Name: "get_pod_status", Content: "<tool_output id=\"t1\">ok</tool_output>"},
	}, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(msgs) != 2 || msgs[0].Role != "system" || msgs[1].Role != "user" {
		t.Fatalf("a flattened step is exactly system + user: %+v", msgs)
	}
	u := *msgs[1].Content
	for _, want := range []string{`"pod":"a"`, `Called: get_pod_status {"pod":"a"}`, `<tool_output id="t1">`, "DONE"} {
		if !strings.Contains(u, want) {
			t.Errorf("flattened text lacks %q:\n%s", want, u)
		}
	}
	if strings.Contains(u, "zzz") || strings.Contains(u, "generatorURL") {
		t.Error("noise fields must be dropped from the alert")
	}
}

func TestLocalContextOverflowIsAReadableError(t *testing.T) {
	o := fakeLocal(t, func(int, map[string]any) (int, string) {
		return 400, `{"error":{"message":"Input exceeds the model's context window.","type":"context_length_exceeded"}}`
	})
	_, err := o.Next(context.Background(), []Turn{{Role: "user", Content: "alert"}})
	if err == nil || !strings.Contains(err.Error(), "context window") {
		t.Fatalf("err: %v", err)
	}
}

func TestLocalServerDown(t *testing.T) {
	o := NewLocal("http://127.0.0.1:1", "", nil, nil)
	_, err := o.Next(context.Background(), []Turn{{Role: "user", Content: "a"}})
	if err == nil || !strings.Contains(err.Error(), "apfel --serve") {
		t.Fatalf("err: %v", err)
	}
}

func TestLocalFitTerminatesWhenEverythingIsAlreadyMinimal(t *testing.T) {
	o := NewLocal("http://x", "", nil, nil)
	hist := []Turn{{Role: "user", Content: "alert"}}
	for i := 0; i < 40; i++ { // far more tool outputs than 4096 tokens can hold even at the minimum size
		hist = append(hist, Turn{Role: "tool", CallID: "c", Name: "a",
			Content: "<tool_output id=\"t\" untrusted=\"true\">\n" + strings.Repeat("x", 3000) + "\n</tool_output>"})
	}
	done := make(chan struct{})
	go func() { _, _ = o.fitHistory(hist, 1300, true); close(done) }()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("fitHistory does not terminate")
	}
}

func TestConstrainSchemaRestrictsCitationsToRealToolOutputs(t *testing.T) {
	schema := json.RawMessage(`{"type":"object","properties":{
	  "evidence":{"type":"array","items":{"type":"object","properties":{"tool_call":{"type":"string"}}}},
	  "next_steps":{"type":"array","items":{"type":"object","properties":{"requires_human_approval":{"type":"boolean"}}}}}}`)
	hist := []Turn{
		{Role: "user", Content: "alert"},
		{Role: "tool", Content: `<tool_output id="t1" untrusted="true">` + "\nok\n</tool_output>"},
		{Role: "tool", Content: `<tool_output id="t2" untrusted="true">` + "\n{\"ok\":false,\"error\":\"duplicate call, same as t1\"}\n</tool_output>"},
		{Role: "tool", Content: `<tool_output id="t3" untrusted="true">` + "\nok\n</tool_output>"},
	}
	got, err := constrainSchema(schema, hist)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := json.Marshal(got)
	for _, want := range []string{`"tool_call":{"enum":["t1","t3"],"type":"string"}`} {
		if !strings.Contains(string(b), want) {
			t.Errorf("schema lacks %s:\n%s", want, b)
		}
	}
}

func TestNewOpenAIUsesOrdinaryToolTurnsAndTheFullPrompt(t *testing.T) {
	c := NewOpenAI("http://127.0.0.1:1", "llama3.1", "tok", nil, nil, 0)
	if c.Flatten || c.System != SystemPrompt || c.ContextTokens != 0 || c.Token != "tok" {
		t.Fatalf("unexpected client config: %+v", c)
	}
	msgs, err := c.messages([]Turn{{Role: "user", Content: "Alert (JSON):\n{}"}}, false)
	if err != nil || len(msgs) != 2 || msgs[0].Role != "system" {
		t.Fatalf("got %+v, %v", msgs, err)
	}
}
