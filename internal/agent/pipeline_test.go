package agent

import (
	"context"
	"strings"
	"testing"
)

const pipelineEvidence = `{"total_calls":20,"total_duration_s":0.07,"actions":[
 {"id":"A1","name":"Parse alert","source":"Alert","output":{"pod":"backend-1"}},
 {"id":"A2","name":"Pod and container status","source":"K8s API","output":{"lastState":"OOMKilled"}},
 {"id":"A5","name":"Rollout history","source":"K8s API","output":{"diff":"--stress-memory=256"}}]}`

func TestPipelineMakesOneModelCallOverTheFixedEvidence(t *testing.T) {
	m := &scripted{steps: []Step{{Text: diag("A2", "A5"), InputTokens: 100, OutputTokens: 20}}}
	p := &Pipeline{Model: m, ModelName: "fake"}
	rep, err := p.Diagnose(context.Background(), `{"labels":{"alertname":"X"}}`, []byte(pipelineEvidence))
	if err != nil {
		t.Fatal(err)
	}
	if rep.Diagnosis == nil || rep.Metrics.ModelSteps != 1 || rep.Metrics.ToolCalls != 20 || rep.Metrics.Mode != "pipeline" {
		t.Fatalf("got %+v", rep.Metrics)
	}
	if rep.Metrics.DurationS < 0.07 {
		t.Errorf("duration must include the collection time, got %v", rep.Metrics.DurationS)
	}
	in := m.history[0][0].Content
	for _, want := range []string{`<tool_output id="A2"`, "--stress-memory=256", `untrusted="true"`} {
		if !strings.Contains(in, want) {
			t.Errorf("model input lacks %q", want)
		}
	}
}

func TestPipelineRejectsCitationsOfActionsThatDidNotRun(t *testing.T) {
	m := &scripted{steps: []Step{{Text: diag("A9")}, {Text: diag("A9")}}}
	p := &Pipeline{Model: m}
	rep, err := p.Diagnose(context.Background(), `{}`, []byte(pipelineEvidence))
	if err == nil || rep.Diagnosis != nil || rep.Metrics.StopReason != "invalid_diagnosis" || rep.Metrics.Repairs != 1 {
		t.Fatalf("want rejection after one repair, got %+v err=%v", rep.Metrics, err)
	}
}

func TestPipelineToolTurnsForTheLocalModel(t *testing.T) {
	m := &scripted{steps: []Step{{Text: diag("A2")}}}
	p := &Pipeline{Model: m, ToolTurns: true, MaxActionBytes: 10}
	if _, err := p.Diagnose(context.Background(), `{}`, []byte(pipelineEvidence)); err != nil {
		t.Fatal(err)
	}
	h := m.history[0]
	if len(h) != 4 || h[1].Role != "tool" || !strings.Contains(h[2].Content, "[truncated]") {
		t.Fatalf("want alert + 3 tool turns with capped output, got %d turns: %+v", len(h), h)
	}
}

func TestCompactDigestKeepsTheDecisiveFactsFirst(t *testing.T) {
	ev := `{"total_calls":20,"actions":[
	 {"id":"A10","name":"Loki","output":{"pod_logs":["a","b","c","d","e","f"]}},
	 {"id":"A2","name":"Pod","output":{"containers":[{"lastState":{"reason":"OOMKilled","exitCode":137}}]}},
	 {"id":"A5","name":"Rollout","output":{"diff":"--stress-memory=256"}},
	 {"id":"A1","name":"Alert","output":{"pod":"backend-1"}}]}`
	m := &scripted{steps: []Step{{Text: diag("A2")}}}
	p := &Pipeline{Model: m, ToolTurns: true, Compact: true}
	if _, err := p.Diagnose(context.Background(), `{}`, []byte(ev)); err != nil {
		t.Fatal(err)
	}
	h := m.history[0]
	if len(h) != 5 || !strings.HasPrefix(h[1].Name, "A1") || !strings.HasPrefix(h[2].Name, "A2") || !strings.HasPrefix(h[3].Name, "A5") {
		t.Fatalf("want A1, A2, A5 first, got %v %v %v", h[1].Name, h[2].Name, h[3].Name)
	}
	if !strings.Contains(h[2].Content, "OOMKilled") || strings.Contains(h[4].Content, `\"a\"`) {
		t.Errorf("digest lost the fact or kept old log lines: %s / %s", h[2].Content, h[4].Content)
	}
}
