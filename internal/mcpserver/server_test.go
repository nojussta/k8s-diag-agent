package mcpserver

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/nojussta/k8s-diag-agent/internal/lab"
	"github.com/nojussta/k8s-diag-agent/internal/tools"
)

type fakeObs struct{ lines int }

func (f fakeObs) PromQuery(context.Context, string) ([]lab.Sample, error) {
	return []lab.Sample{{Metric: map[string]string{"pod": "backend-x"}, Value: 3}}, nil
}
func (f fakeObs) LokiQuery(context.Context, string, time.Duration, int) ([]lab.LogLine, error) {
	out := make([]lab.LogLine, f.lines)
	for i := range out {
		out[i] = lab.LogLine{Line: strings.Repeat("x", 400)}
	}
	return out, nil
}

func connect(t *testing.T, obs fakeObs) *mcp.ClientSession {
	t.Helper()
	cs := fake.NewClientset(
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "backend-x", Namespace: "demo", Labels: map[string]string{"app": "backend"}},
			Spec:       corev1.PodSpec{NodeName: "n1", Containers: []corev1.Container{{Name: "backend"}}},
			Status: corev1.PodStatus{Phase: corev1.PodRunning, ContainerStatuses: []corev1.ContainerStatus{{
				Name: "backend", RestartCount: 4,
				LastTerminationState: corev1.ContainerState{Terminated: &corev1.ContainerStateTerminated{Reason: "OOMKilled", ExitCode: 137}}}}},
		},
		&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "backend", Namespace: "demo"}},
	)
	srv := New(&tools.Toolset{CS: cs, Obs: obs})
	ct, st := mcp.NewInMemoryTransports()
	ctx := context.Background()
	if _, err := srv.Connect(ctx, st, nil); err != nil {
		t.Fatal(err)
	}
	sess, err := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "0"}, nil).Connect(ctx, ct, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { sess.Close() })
	return sess
}

func call(t *testing.T, s *mcp.ClientSession, name string, args map[string]any) (tools.Result, *mcp.CallToolResult) {
	t.Helper()
	res, err := s.CallTool(context.Background(), &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	var r tools.Result
	if len(res.Content) == 0 {
		return r, res
	}
	if err := json.Unmarshal([]byte(res.Content[0].(*mcp.TextContent).Text), &r); err != nil {
		t.Fatalf("%s: bad envelope: %v", name, err)
	}
	return r, res
}

// Security level 2: the server exposes exactly the ten read-only tools.
func TestServerExposesOnlyTheTenReadOnlyTools(t *testing.T) {
	s := connect(t, fakeObs{})
	list, err := s.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]bool{}
	for _, tl := range list.Tools {
		got[tl.Name] = true
		for _, bad := range []string{"delete", "exec", "apply", "patch", "scale", "create", "update", "restart", "write"} {
			if strings.Contains(tl.Name, bad) {
				t.Errorf("tool %q looks like a write tool", tl.Name)
			}
		}
	}
	if len(got) != 10 {
		t.Fatalf("want 10 tools, got %d: %v", len(got), got)
	}
	for _, n := range tools.Names {
		if !got[n] {
			t.Errorf("missing tool %s", n)
		}
	}
}

func TestGetPodStatusReturnsOOMKilled(t *testing.T) {
	s := connect(t, fakeObs{})
	r, res := call(t, s, "get_pod_status", map[string]any{"namespace": "demo", "pod": "backend-x"})
	if !r.OK || res.IsError {
		t.Fatalf("not ok: %+v", r)
	}
	b, _ := json.Marshal(r.Data)
	if !strings.Contains(string(b), "OOMKilled") || !strings.Contains(string(b), "137") {
		t.Errorf("OOMKilled evidence missing: %s", b)
	}
}

func TestUnknownWriteToolDoesNotExist(t *testing.T) {
	s := connect(t, fakeObs{})
	_, err := s.CallTool(context.Background(), &mcp.CallToolParams{Name: "delete_pod", Arguments: map[string]any{"pod": "x"}})
	if err == nil {
		t.Fatal("delete_pod must not exist")
	}
}

func TestInputValidationRejectsInjection(t *testing.T) {
	s := connect(t, fakeObs{})
	for _, ns := range []string{"../kube-system", "demo; rm -rf /", "DEMO", "", strings.Repeat("a", 300)} {
		r, res := call(t, s, "get_pod_status", map[string]any{"namespace": ns, "pod": "backend-x"})
		if r.OK || !res.IsError {
			t.Errorf("namespace %q should be rejected", ns)
		}
	}
}

func TestMissingPodIsAnErrorEnvelopeNotACrash(t *testing.T) {
	s := connect(t, fakeObs{})
	r, res := call(t, s, "get_pod_status", map[string]any{"namespace": "demo", "pod": "nope"})
	if r.OK || !res.IsError || r.Error == "" {
		t.Errorf("want error envelope, got %+v", r)
	}
}

func TestResultsAreCappedAt8KB(t *testing.T) {
	s := connect(t, fakeObs{lines: 50}) // 50 x 400 B = 20 KB of logs
	r, res := call(t, s, "query_logs", map[string]any{"logql": `{namespace="demo"}`})
	if !r.OK || !r.Truncated {
		t.Fatalf("want ok+truncated, got ok=%v truncated=%v", r.OK, r.Truncated)
	}
	if n := len(res.Content[0].(*mcp.TextContent).Text); n > tools.MaxBytes {
		t.Errorf("result is %d bytes, cap is %d", n, tools.MaxBytes)
	}
}

func TestQueryMetricsRowCapAndBounds(t *testing.T) {
	s := connect(t, fakeObs{})
	r, _ := call(t, s, "query_metrics", map[string]any{"promql": "up"})
	if !r.OK {
		t.Fatalf("%+v", r)
	}
	r, _ = call(t, s, "query_metrics", map[string]any{"promql": strings.Repeat("a", 600)})
	if r.OK {
		t.Error("oversized query must be rejected")
	}
}
