package agent_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/nojussta/k8s-diag-agent/internal/agent"
	"github.com/nojussta/k8s-diag-agent/internal/lab"
	"github.com/nojussta/k8s-diag-agent/internal/mcpserver"
	"github.com/nojussta/k8s-diag-agent/internal/tools"
)

type noObs struct{}

func (noObs) PromQuery(context.Context, string) ([]lab.Sample, error) { return nil, nil }
func (noObs) LokiQuery(context.Context, string, time.Duration, int) ([]lab.LogLine, error) {
	return nil, nil
}

type script struct {
	steps []agent.Step
	i     int
	last  []agent.Turn
}

func (s *script) Next(_ context.Context, h []agent.Turn) (agent.Step, error) {
	s.last = h
	st := s.steps[s.i]
	s.i++
	return st, nil
}

// Orchestrator + real MCP client + real MCP server + fake cluster: the whole
// path except the language model.
func TestEndToEndThroughMCP(t *testing.T) {
	cs := fake.NewClientset(&corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "backend-x", Namespace: "demo"},
		Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: "backend"}}},
		Status: corev1.PodStatus{Phase: corev1.PodRunning, ContainerStatuses: []corev1.ContainerStatus{{
			Name: "backend", RestartCount: 4,
			LastTerminationState: corev1.ContainerState{Terminated: &corev1.ContainerStateTerminated{Reason: "OOMKilled", ExitCode: 137}}}}},
	})
	ct, st := mcp.NewInMemoryTransports()
	ctx := context.Background()
	if _, err := mcpserver.New(&tools.Toolset{CS: cs, Obs: noObs{}}).Connect(ctx, st, nil); err != nil {
		t.Fatal(err)
	}
	client, err := agent.ConnectTransport(ctx, ct)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	allowed := map[string]bool{}
	for _, n := range tools.Names {
		allowed[n] = true
	}
	defs, err := client.Defs(ctx, allowed)
	if err != nil || len(defs) != 10 {
		t.Fatalf("defs=%d err=%v", len(defs), err)
	}
	for _, d := range defs {
		if _, bad := d.Parameters["$schema"]; bad || d.Parameters["additionalProperties"] != nil {
			t.Errorf("%s: schema not cleaned: %v", d.Name, d.Parameters)
		}
		if d.Parameters["type"] != "object" || d.Description == "" {
			t.Errorf("%s: bad declaration", d.Name)
		}
	}

	diag := `{"summary":"s","root_cause":{"category":"OOMKilled","component":"demo/pod/backend-x","explanation":"e"},"confidence":"high",` +
		`"evidence":[{"tool_call":"t1","source":"get_pod_status","fact":"OOMKilled exit 137"}],"impact":"i","next_steps":[],"unknowns":[]}`
	m := &script{steps: []agent.Step{
		{Calls: []agent.Call{{ID: "g1", Name: "get_pod_status", Args: map[string]any{"namespace": "demo", "pod": "backend-x"}}}},
		{Text: diag},
	}}
	a := &agent.Agent{Model: m, ModelName: "script", Tools: client, Allowed: allowed}
	rep, err := a.Investigate(ctx, "{}")
	if err != nil {
		t.Fatal(err)
	}
	if rep.Diagnosis == nil || rep.Metrics.ToolCalls != 1 {
		t.Fatalf("%+v", rep)
	}
	var sawTool bool
	for _, turn := range m.last {
		if turn.Role == "tool" {
			sawTool = true
			if !strings.Contains(turn.Content, "OOMKilled") || !strings.Contains(turn.Content, `id="t1"`) {
				t.Errorf("model did not get the real tool result: %s", turn.Content)
			}
		}
	}
	if !sawTool {
		t.Error("no tool turn")
	}
	_ = json.Marshal
}
