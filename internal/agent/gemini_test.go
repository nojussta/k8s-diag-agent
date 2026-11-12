package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func fakeGemini(t *testing.T, handler func(n int, body map[string]any, hdr http.Header) (int, string)) (*Gemini, *int32) {
	t.Helper()
	var n int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		var body map[string]any
		_ = json.Unmarshal(b, &body)
		if r.URL.Path != "/v1beta/interactions" {
			t.Errorf("path %s", r.URL.Path)
		}
		code, resp := handler(int(atomic.AddInt32(&n, 1)), body, r.Header)
		w.WriteHeader(code)
		io.WriteString(w, resp)
	}))
	t.Cleanup(srv.Close)
	g := NewGemini("SECRET-KEY", "gemini-test", []ToolDef{{Name: "get_pod_status", Description: "d", Parameters: map[string]any{"type": "object"}}}, json.RawMessage(`{"type":"object"}`))
	g.BaseURL = srv.URL
	g.Sleep = func(time.Duration) {}
	return g, &n
}

const callResp = `{"steps":[{"type":"thought","signature":"SIG123"},{"type":"function_call","id":"fc1","name":"get_pod_status","arguments":{"namespace":"demo","pod":"a"}}],"usage":{"total_input_tokens":120,"initial_input_tokens":120,"total_output_tokens":15,"total_thought_tokens":30,"total_cached_tokens":50}}`

func TestRequestShapeAndFunctionCallParsing(t *testing.T) {
	g, _ := fakeGemini(t, func(_ int, body map[string]any, hdr http.Header) (int, string) {
		if hdr.Get("x-goog-api-key") != "SECRET-KEY" {
			t.Error("api key header missing")
		}
		if body["store"] != false || body["model"] != "gemini-test" || body["system_instruction"] == "" {
			t.Errorf("body: %v", body)
		}
		if _, ok := body["response_format"].(map[string]any)["schema"]; !ok {
			t.Error("response_format.schema missing")
		}
		tools := body["tools"].([]any)
		if tools[0].(map[string]any)["type"] != "function" || tools[0].(map[string]any)["name"] != "get_pod_status" {
			t.Errorf("tools: %v", tools)
		}
		in := body["input"].([]any)
		if len(in) != 1 || in[0].(map[string]any)["type"] != "user_input" {
			t.Errorf("input: %v", in)
		}
		return 200, callResp
	})
	st, err := g.Next(context.Background(), []Turn{{Role: "user", Content: "alert"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(st.Calls) != 1 || st.Calls[0].ID != "fc1" || st.Calls[0].Name != "get_pod_status" || st.Calls[0].Args["pod"] != "a" {
		t.Errorf("calls: %+v", st.Calls)
	}
	if st.InputTokens != 120 || st.OutputTokens != 45 {
		t.Errorf("tokens in=%d out=%d (thinking counts as output, cached and initial_* must not be double counted)", st.InputTokens, st.OutputTokens)
	}
}

// The model's own steps (with the thought signature) must come back verbatim,
// followed by our function_result, or multi-step function calling breaks.
func TestModelStepsAndSignaturesAreEchoedBack(t *testing.T) {
	var first Step
	g, _ := fakeGemini(t, func(n int, body map[string]any, _ http.Header) (int, string) {
		if n == 1 {
			return 200, callResp
		}
		in := body["input"].([]any)
		types := []string{}
		for _, s := range in {
			types = append(types, s.(map[string]any)["type"].(string))
		}
		if strings.Join(types, ",") != "user_input,thought,function_call,function_result" {
			t.Errorf("input types: %v", types)
		}
		if in[1].(map[string]any)["signature"] != "SIG123" {
			t.Error("thought signature not echoed")
		}
		fr := in[3].(map[string]any)
		if fr["call_id"] != "fc1" || fr["name"] != "get_pod_status" || !strings.Contains(fr["result"].([]any)[0].(map[string]any)["text"].(string), "ok") {
			t.Errorf("function_result: %v", fr)
		}
		return 200, `{"steps":[{"type":"model_output","content":[{"type":"text","text":"{\"done\":true}"}]}]}`
	})
	hist := []Turn{{Role: "user", Content: "alert"}}
	first, _ = g.Next(context.Background(), hist)
	hist = append(hist, Turn{Role: "model", Raw: first.Raw}, Turn{Role: "tool", CallID: "fc1", Name: "get_pod_status", Content: `{"ok":true}`})
	st, err := g.Next(context.Background(), hist)
	if err != nil || st.Text != `{"done":true}` || len(st.Calls) != 0 {
		t.Errorf("err=%v step=%+v", err, st)
	}
}

func TestModelOutputAsPlainStringContent(t *testing.T) {
	st, err := parseInteraction([]byte(`{"steps":[{"type":"model_output","content":"hello"}]}`))
	if err != nil || st.Text != "hello" {
		t.Errorf("%v %+v", err, st)
	}
}

func TestRetriesOn429ThenSucceeds(t *testing.T) {
	g, n := fakeGemini(t, func(n int, _ map[string]any, _ http.Header) (int, string) {
		if n < 3 {
			return 429, `{"error":{"message":"quota"}}`
		}
		return 200, callResp
	})
	if _, err := g.Next(context.Background(), []Turn{{Role: "user", Content: "x"}}); err != nil {
		t.Fatal(err)
	}
	if *n != 3 {
		t.Errorf("attempts=%d", *n)
	}
}

func TestClientErrorsAreNotRetriedAndNeverLeakTheKey(t *testing.T) {
	g, n := fakeGemini(t, func(int, map[string]any, http.Header) (int, string) {
		return 400, `{"error":{"message":"bad key SECRET-KEY"}}`
	})
	_, err := g.Next(context.Background(), []Turn{{Role: "user", Content: "x"}})
	if err == nil || *n != 1 {
		t.Fatalf("err=%v attempts=%d", err, *n)
	}
	if strings.Contains(err.Error(), "SECRET-KEY") {
		t.Errorf("key leaked: %v", err)
	}
}

func TestGivesUpAfterRetriesAndMissingKey(t *testing.T) {
	g, n := fakeGemini(t, func(int, map[string]any, http.Header) (int, string) { return 503, "overloaded" })
	if _, err := g.Next(context.Background(), []Turn{{Role: "user", Content: "x"}}); err == nil || *n != 5 {
		t.Errorf("err=%v attempts=%d", err, *n)
	}
	g.Key = ""
	if _, err := g.Next(context.Background(), []Turn{{Role: "user", Content: "x"}}); err == nil || !strings.Contains(err.Error(), "GEMINI_API_KEY") {
		t.Errorf("err=%v", err)
	}
}

func TestDailyQuotaFailsFastWithoutRetries(t *testing.T) {
	g, n := fakeGemini(t, func(int, map[string]any, http.Header) (int, string) {
		return 429, `{"error":{"message":"Rate limit exceeded (limit: 20 requests per day on Free Tier). Please retry in 23h33m50s"}}`
	})
	_, err := g.Next(context.Background(), []Turn{{Role: "user", Content: "x"}})
	if err == nil || *n != 1 || !strings.Contains(err.Error(), "daily quota") {
		t.Errorf("err=%v attempts=%d", err, *n)
	}
}

func TestFinalSendsNoToolsAndAsksForTheAnswer(t *testing.T) {
	g, _ := fakeGemini(t, func(_ int, body map[string]any, _ http.Header) (int, string) {
		if _, has := body["tools"]; has {
			t.Errorf("Final must not offer tools: %v", body["tools"])
		}
		in := body["input"].([]any)
		last := in[len(in)-1].(map[string]any)
		if !strings.Contains(fmt.Sprint(last["content"]), "no more tool calls") {
			t.Errorf("last input should ask for the answer, got %v", last)
		}
		return 200, `{"steps":[{"type":"model_output","content":[{"type":"text","text":"{}"}]}]}`
	})
	var _ Finalizer = g
	if _, err := g.Final(context.Background(), []Turn{{Role: "user", Content: "x"}}); err != nil {
		t.Fatal(err)
	}
}

func TestPerMinuteLimitWaitsForTheServerHint(t *testing.T) {
	var slept []time.Duration
	g, _ := fakeGemini(t, func(n int, _ map[string]any, _ http.Header) (int, string) {
		if n == 1 {
			return 429, `{"error":{"message":"Rate limit exceeded (limit: 15 requests per minute). Please retry in 43s"}}`
		}
		return 200, callResp
	})
	g.Sleep = func(d time.Duration) { slept = append(slept, d) }
	if _, err := g.Next(context.Background(), []Turn{{Role: "user", Content: "x"}}); err != nil {
		t.Fatal(err)
	}
	if len(slept) != 1 || slept[0] != 44*time.Second || g.WaitedSeconds() != 44 {
		t.Fatalf("want one 44 s wait recorded, got %v / %v", slept, g.WaitedSeconds())
	}
}
