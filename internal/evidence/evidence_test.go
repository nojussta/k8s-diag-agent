package evidence

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/utils/ptr"

	"github.com/nojussta/k8s-diag-agent/internal/alertmsg"
	"github.com/nojussta/k8s-diag-agent/internal/lab"
)

type fakeObs struct{ queries []string }

func (f *fakeObs) PromQuery(_ context.Context, expr string) ([]lab.Sample, error) {
	f.queries = append(f.queries, expr)
	return []lab.Sample{{Metric: map[string]string{"pod": "backend-new-x"}, Value: 1}}, nil
}
func (f *fakeObs) LokiQuery(_ context.Context, q string, _ time.Duration, _ int) ([]lab.LogLine, error) {
	f.queries = append(f.queries, q)
	return []lab.LogLine{{Line: "line"}}, nil
}

func replicaSet(name string, rev string, args []string, cause string) *appsv1.ReplicaSet {
	ann := map[string]string{"deployment.kubernetes.io/revision": rev}
	if cause != "" {
		ann["kubernetes.io/change-cause"] = cause
	}
	return &appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "demo", Annotations: ann,
			Labels:          map[string]string{"app": "backend"},
			OwnerReferences: []metav1.OwnerReference{{Kind: "Deployment", Name: "backend"}}},
		Spec: appsv1.ReplicaSetSpec{Template: corev1.PodTemplateSpec{Spec: corev1.PodSpec{
			Containers: []corev1.Container{{Name: "backend", Image: "podinfo:6", Args: args}}}}},
	}
}

func lab1() *fake.Clientset {
	ready := ptr.To(true)
	return fake.NewClientset(
		&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "backend", Namespace: "demo"},
			Spec: appsv1.DeploymentSpec{Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "backend"}}}},
		replicaSet("backend-old", "1", []string{"--port=9898"}, ""),
		replicaSet("backend-new", "2", []string{"--port=9898", "--stress-memory=256"}, "release 2.0"),
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "backend-new-x", Namespace: "demo", Labels: map[string]string{"app": "backend"},
				OwnerReferences: []metav1.OwnerReference{{Kind: "ReplicaSet", Name: "backend-new"}}},
			Spec: corev1.PodSpec{NodeName: "n1", Containers: []corev1.Container{{Name: "backend"}}},
			Status: corev1.PodStatus{Phase: corev1.PodRunning, ContainerStatuses: []corev1.ContainerStatus{{
				Name: "backend", RestartCount: 4,
				State: corev1.ContainerState{Waiting: &corev1.ContainerStateWaiting{Reason: "CrashLoopBackOff"}},
				LastTerminationState: corev1.ContainerState{Terminated: &corev1.ContainerStateTerminated{
					Reason: "OOMKilled", ExitCode: 137}}}}},
		},
		&corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "backend", Namespace: "demo"},
			Spec: corev1.ServiceSpec{Selector: map[string]string{"app": "backend"}}},
		&corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "other", Namespace: "demo"},
			Spec: corev1.ServiceSpec{Selector: map[string]string{"app": "other"}}},
		&discoveryv1.EndpointSlice{ObjectMeta: metav1.ObjectMeta{Name: "backend-abc", Namespace: "demo",
			Labels: map[string]string{"kubernetes.io/service-name": "backend"}},
			Endpoints: []discoveryv1.Endpoint{{Conditions: discoveryv1.EndpointConditions{Ready: ready}},
				{Conditions: discoveryv1.EndpointConditions{Ready: ptr.To(false)}}}},
		&corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "n1"},
			Status: corev1.NodeStatus{Conditions: []corev1.NodeCondition{{Type: corev1.NodeReady, Status: corev1.ConditionTrue}}}},
	)
}

func action(t *testing.T, e Evidence, id string) Action {
	t.Helper()
	for _, a := range e.Actions {
		if a.ID == id {
			return a
		}
	}
	t.Fatalf("action %s missing", id)
	return Action{}
}

func TestCollectOOMKilledScenario(t *testing.T) {
	obs := &fakeObs{}
	alert := alertmsg.Alert{Status: "firing", Labels: map[string]string{
		"alertname": "LabPodCrashLooping", "namespace": "demo", "pod": "backend-new-x", "container": "backend"}}
	e := Collect(context.Background(), lab1(), obs, alert)

	if len(e.Actions) != 11 {
		t.Fatalf("want 11 actions, got %d", len(e.Actions))
	}
	for _, a := range e.Actions {
		if !a.OK {
			t.Errorf("%s failed: %v", a.ID, a.Output)
		}
	}
	// The reference sequence needs 20 read-only calls for a pod-level alert
	// (the count recorded in results/s1-oomkilled).
	if e.TotalCalls != 20 {
		t.Errorf("want 20 calls, got %d", e.TotalCalls)
	}

	body, _ := json.Marshal(action(t, e, "A2").Output)
	for _, want := range []string{`"reason":"OOMKilled"`, `"exitCode":137`, `CrashLoopBackOff`} {
		if !strings.Contains(string(body), want) {
			t.Errorf("A2 missing %s in %s", want, body)
		}
	}
	body, _ = json.Marshal(action(t, e, "A5").Output)
	for _, want := range []string{`"revision":2`, `--stress-memory=256`, `release 2.0`, `template_diff_latest_vs_previous`} {
		if !strings.Contains(string(body), want) {
			t.Errorf("A5 missing %s in %s", want, body)
		}
	}
	body, _ = json.Marshal(action(t, e, "A8").Output)
	if !strings.Contains(string(body), `"endpoints":2,"ready":1,"service":"backend"`) || strings.Contains(string(body), "other") {
		t.Errorf("A8 unexpected: %s", body)
	}
	if !strings.Contains(e.Markdown(), "## A5 Owner, rollout history") {
		t.Error("markdown missing A5 heading")
	}
}

func TestDeploymentLevelAlertPicksLeastHealthyPod(t *testing.T) {
	cs := lab1()
	_, _ = cs.CoreV1().Pods("demo").Create(context.Background(), &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "backend-old-y", Namespace: "demo", Labels: map[string]string{"app": "backend"}},
		Status:     corev1.PodStatus{Phase: corev1.PodRunning},
	}, metav1.CreateOptions{})
	c := &collector{cs: cs, obs: &fakeObs{}}
	got, err := c.pickTarget(context.Background(), alertmsg.Alert{Labels: map[string]string{
		"alertname": "LabDeploymentUnavailable", "namespace": "demo", "deployment": "backend"}})
	if err != nil || got.Pod != "backend-new-x" {
		t.Fatalf("want backend-new-x (4 restarts), got %q err=%v", got.Pod, err)
	}
	if c.calls != 2 {
		t.Errorf("want 2 calls, got %d", c.calls)
	}
}

func TestDiffTemplatesIgnoresUnchangedAndNull(t *testing.T) {
	older := map[string]containerView{"b": {"image": "x", "args": []string{"a"}, "env": nil}}
	newer := map[string]containerView{"b": {"image": "x", "args": []string{"a", "b"}, "env": nil}}
	got := DiffTemplates(newer, older)
	if len(got) != 1 || got[0].Field != "args" {
		t.Fatalf("want one args change, got %+v", got)
	}
}

func TestMissingTargetIsReported(t *testing.T) {
	e := Collect(context.Background(), fake.NewClientset(), &fakeObs{}, alertmsg.Alert{Labels: map[string]string{"namespace": "demo"}})
	if action(t, e, "A1").OK {
		t.Fatal("A1 should fail without pod or deployment label")
	}
}
