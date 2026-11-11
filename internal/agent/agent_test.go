package agent

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

type scripted struct {
	steps   []Step
	i       int
	history [][]Turn
}

func (s *scripted) Next(_ context.Context, h []Turn) (Step, error) {
	s.history = append(s.history, append([]Turn(nil), h...))
	if s.i >= len(s.steps) {
		return Step{}, errors.New("script exhausted")
	}
	st := s.steps[s.i]
	s.i++
	return st, nil
}

type fakeTools struct{ calls []string }

func (f *fakeTools) Call(_ context.Context, name string, _ map[string]any) (string, error) {
	f.calls = append(f.calls, name)
	return `{"ok":true}`, nil
}

func call(id, name string, kv ...any) Call {
	args := map[string]any{}
	for i := 0; i < len(kv); i += 2 {
		args[kv[i].(string)] = kv[i+1]
	}
	return Call{ID: id, Name: name, Args: args}
}

// diag returns a valid diagnosis citing the given tool-call ids.
func diag(cites ...string) string {
	ev := []string{}
	for _, c := range cites {
		ev = append(ev, fmt.Sprintf(`{"tool_call":%q,"source":"get_pod_status","fact":"lastState.terminated.reason=OOMKilled"}`, c))
	}
	return `{"summary":"backend was OOMKilled","root_cause":{"category":"OOMKilled","component":"demo/deployment/backend","explanation":"memory limit exceeded"},` +
		`"confidence":"high","evidence":[` + strings.Join(ev, ",") + `],"impact":"rollout stuck","next_steps":[{"action":"roll back","command":"kubectl rollout undo deployment/backend -n demo","requires_human_approval":true}],"unknowns":[]}`
}

func newAgent(m Model, t Tools) *Agent {
	return &Agent{Model: m, ModelName: "fake", Tools: t, Allowed: map[string]bool{"get_pod_status": true, "get_events": true}, Limits: DefaultLimits}
}

func TestHappyPathProducesDiagnosisAndMetrics(t *testing.T) {
	m := &scripted{steps: []Step{
		{Calls: []Call{call("g1", "get_pod_status", "pod", "a")}, InputTokens: 100, OutputTokens: 10},
		{Text: diag("t1"), InputTokens: 150, OutputTokens: 40},
	}}
	ft := &fakeTools{}
	var trace bytes.Buffer
	a := newAgent(m, ft)
	a.Trace = &trace
	rep, err := a.Investigate(context.Background(), `{"alertname":"X"}`)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Metrics.ToolCalls != 1 || rep.Metrics.InputTokens != 250 || rep.Metrics.OutputTokens != 50 || rep.Metrics.StopReason != "diagnosis" || rep.Metrics.Model != "fake" {
		t.Errorf("metrics: %+v", rep.Metrics)
	}
	if rep.Diagnosis == nil || rep.Diagnosis.RootCause.Category != "OOMKilled" {
		t.Errorf("diagnosis: %+v", rep.Diagnosis)
	}
	if n := strings.Count(trace.String(), "\n"); n != 4 { // alert, tool_call, final, end
		t.Errorf("trace has %d lines:\n%s", n, trace.String())
	}
}

func TestToolOutsideAllowListIsRejectedAndNeverExecuted(t *testing.T) {
	m := &scripted{steps: []Step{
		{Calls: []Call{call("g1", "delete_pod", "pod", "a")}},
		{Calls: []Call{call("g2", "get_pod_status", "pod", "a")}},
		{Text: diag("t2")},
	}}
	ft := &fakeTools{}
	rep, err := newAgent(m, ft).Investigate(context.Background(), "{}")
	if err != nil {
		t.Fatal(err)
	}
	if len(ft.calls) != 1 || ft.calls[0] != "get_pod_status" || rep.Metrics.Rejected != 1 {
		t.Errorf("calls=%v metrics=%+v", ft.calls, rep.Metrics)
	}
}

func TestDuplicateCallIsNotExecutedTwice(t *testing.T) {
	same := call("g1", "get_pod_status", "namespace", "demo", "pod", "a")
	again := call("g2", "get_pod_status", "pod", "a", "namespace", "demo") // same args, other order
	m := &scripted{steps: []Step{{Calls: []Call{same}}, {Calls: []Call{again}}, {Text: diag("t1")}}}
	ft := &fakeTools{}
	rep, err := newAgent(m, ft).Investigate(context.Background(), "{}")
	if err != nil || len(ft.calls) != 1 || rep.Metrics.Rejected != 1 {
		t.Errorf("err=%v calls=%v metrics=%+v", err, ft.calls, rep.Metrics)
	}
}

func TestCallBudgetStopsAnEndlessModel(t *testing.T) {
	var steps []Step
	for i := 0; i < 40; i++ {
		steps = append(steps, Step{Calls: []Call{call("g", "get_events", "object", fmt.Sprintf("o%d", i))}})
	}
	ft := &fakeTools{}
	rep, err := newAgent(&scripted{steps: steps}, ft).Investigate(context.Background(), "{}")
	if err == nil || rep.Metrics.StopReason != "call_budget" || len(ft.calls) != 15 {
		t.Errorf("err=%v reason=%s calls=%d", err, rep.Metrics.StopReason, len(ft.calls))
	}
}

type slowModel struct{}

func (slowModel) Next(ctx context.Context, _ []Turn) (Step, error) {
	<-ctx.Done()
	return Step{}, ctx.Err()
}

func TestTimeoutIsEnforced(t *testing.T) {
	a := newAgent(slowModel{}, &fakeTools{})
	a.Limits = Limits{MaxCalls: 15, Timeout: 50 * time.Millisecond}
	rep, err := a.Investigate(context.Background(), "{}")
	if err == nil || rep.Metrics.StopReason != "timeout" {
		t.Errorf("err=%v reason=%s", err, rep.Metrics.StopReason)
	}
}

func TestToolOutputIsMarkedUntrustedAndCarriesTheCallID(t *testing.T) {
	m := &scripted{steps: []Step{{Calls: []Call{call("g1", "get_events", "object", "a")}}, {Text: diag("t1")}}}
	newAgent(m, &fakeTools{}).Investigate(context.Background(), "{}")
	h := m.history[1]
	last := h[len(h)-1]
	if last.Role != "tool" || last.CallID != "g1" || !strings.Contains(last.Content, `untrusted="true"`) || !strings.Contains(last.Content, `id="t1"`) {
		t.Errorf("tool turn wrong: %+v", last)
	}
}

func TestEmptyStepIsAnError(t *testing.T) {
	_, err := newAgent(&scripted{steps: []Step{{}}}, &fakeTools{}).Investigate(context.Background(), "{}")
	if err == nil {
		t.Error("want error")
	}
}

// Evidence must cite a call that really ran . One repair is allowed.
func TestInventedEvidenceIsSentBackOnceThenAccepted(t *testing.T) {
	m := &scripted{steps: []Step{
		{Calls: []Call{call("g1", "get_pod_status", "pod", "a")}},
		{Text: diag("t9")}, // t9 never ran
		{Text: diag("t1")},
	}}
	rep, err := newAgent(m, &fakeTools{}).Investigate(context.Background(), "{}")
	if err != nil || rep.Metrics.Repairs != 1 || rep.Diagnosis == nil {
		t.Fatalf("err=%v metrics=%+v", err, rep.Metrics)
	}
	h := m.history[2]
	if last := h[len(h)-1]; last.Role != "user" || !strings.Contains(last.Content, "t9") {
		t.Errorf("repair message missing: %+v", last)
	}
}

func TestInvalidTwiceIsRejected(t *testing.T) {
	m := &scripted{steps: []Step{{Text: "not json"}, {Text: "still not json"}}}
	rep, err := newAgent(m, &fakeTools{}).Investigate(context.Background(), "{}")
	if err == nil || rep.Diagnosis != nil || rep.Metrics.StopReason != "invalid_diagnosis" {
		t.Errorf("err=%v rep=%+v", err, rep)
	}
}

func TestParseDiagnosisRules(t *testing.T) {
	ok := map[string]bool{"t1": true}
	good := diag("t1")
	if _, err := ParseDiagnosis([]byte("```json\n"+good+"\n```"), ok); err != nil {
		t.Errorf("fenced JSON should be accepted: %v", err)
	}
	cases := map[string]string{
		"unknown field":     strings.Replace(good, `"impact"`, `"extra":1,"impact"`, 1),
		"bad category":      strings.Replace(good, `"OOMKilled","component"`, `"Gremlins","component"`, 1),
		"bad confidence":    strings.Replace(good, `"high"`, `"certain"`, 1),
		"command w/o human": strings.Replace(good, `"requires_human_approval":true`, `"requires_human_approval":false`, 1),
		"empty evidence":    strings.Replace(good, `"evidence":[{"tool_call":"t1","source":"get_pod_status","fact":"lastState.terminated.reason=OOMKilled"}]`, `"evidence":[]`, 1),
		"uncited tool call": strings.Replace(good, `"tool_call":"t1"`, `"tool_call":"t7"`, 1),
	}
	for name, bad := range cases {
		if _, err := ParseDiagnosis([]byte(bad), ok); err == nil {
			t.Errorf("%s: should be rejected", name)
		}
	}
}

// finalScripted is a model that can also be asked for its final answer.
type finalScripted struct {
	scripted
	final  Step
	nFinal int
}

func (f *finalScripted) Final(_ context.Context, h []Turn) (Step, error) {
	f.nFinal++
	return f.final, nil
}

func TestRepeatedDuplicatesForceTheFinalAnswer(t *testing.T) {
	dup := Step{Calls: []Call{call("g", "get_pod_status", "pod", "a")}}
	m := &finalScripted{scripted: scripted{steps: []Step{dup, dup, dup, dup, dup}}, final: Step{Text: diag("t1")}}
	rep, err := newAgent(m, &fakeTools{}).Investigate(context.Background(), `{}`)
	if err != nil {
		t.Fatal(err)
	}
	if m.nFinal != 1 || rep.Diagnosis == nil || rep.Metrics.ToolCalls != 1 || rep.Metrics.Rejected != 2 {
		t.Errorf("finals=%d metrics=%+v diagnosis=%v", m.nFinal, rep.Metrics, rep.Diagnosis)
	}
	if m.i != 3 {
		t.Errorf("the loop must stop asking for calls after %d strikes; model was asked %d times", maxStrikes, m.i)
	}
}

func TestCallBudgetForcesTheFinalAnswerWhenTheModelCanFinalize(t *testing.T) {
	steps := []Step{}
	for i := 0; i < 5; i++ {
		steps = append(steps, Step{Calls: []Call{call("g", "get_pod_status", "pod", fmt.Sprint(i))}})
	}
	m := &finalScripted{scripted: scripted{steps: steps}, final: Step{Text: diag("t1")}}
	a := newAgent(m, &fakeTools{})
	a.Limits = Limits{MaxCalls: 3, Timeout: time.Minute}
	rep, err := a.Investigate(context.Background(), `{}`)
	if err != nil || rep.Diagnosis == nil || rep.Metrics.ToolCalls != 3 || m.nFinal != 1 {
		t.Fatalf("err=%v finals=%d metrics=%+v", err, m.nFinal, rep.Metrics)
	}
}
