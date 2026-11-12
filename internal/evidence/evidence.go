// Package evidence is the executable form of the "automatable diagnostic
// action set" (A1-A11, see docs/architecture.md). It runs every action in a
// fixed order, exactly as a careful engineer would during first-line triage,
// and records what each action returned, how long it took and how many
// read-only calls it needed.
//
// It serves three purposes:
//  1. proves that every action in the set can be done with read-only access;
//  2. gives the agent prototype a reference for its MCP tool catalogue;
//  3. is the "fixed sequence" baseline the agent will be compared against
//     (the agent should reach the cause with fewer calls than this).
package evidence

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"

	"github.com/nojussta/k8s-diag-agent/internal/alertmsg"
	"github.com/nojussta/k8s-diag-agent/internal/lab"
)

const (
	// LogTail is the number of lines per log call; it keeps the evidence (and
	// later the model context) small.
	LogTail  = 40
	lookback = "15m"
)

// Observability is the read-only Prometheus and Loki access the actions need.
// *lab.Lab implements it; tests use a fake.
type Observability interface {
	PromQuery(ctx context.Context, expr string) ([]lab.Sample, error)
	LokiQuery(ctx context.Context, logql string, since time.Duration, limit int) ([]lab.LogLine, error)
}

// Action is the record of one action.
type Action struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Source    string  `json:"source"`
	OK        bool    `json:"ok"`
	Calls     int     `json:"calls"`
	DurationS float64 `json:"duration_s"`
	Output    any     `json:"output"`
}

// Target is what A1 resolves the alert to.
type Target struct {
	Alertname  string `json:"alertname"`
	Namespace  string `json:"namespace"`
	Pod        string `json:"pod"`
	Container  string `json:"container"`
	Deployment string `json:"deployment"`
	Service    string `json:"service,omitempty"`
	Summary    string `json:"summary"`
	StartsAt   string `json:"startsAt"`
}

// Evidence is the whole run.
type Evidence struct {
	Target         Target   `json:"target"`
	TotalCalls     int      `json:"total_calls"`
	TotalDurationS float64  `json:"total_duration_s"`
	Actions        []Action `json:"actions"`
}

type collector struct {
	cs      kubernetes.Interface
	obs     Observability
	calls   int
	actions []Action
}

// Collect runs A1-A11 for one alert.
func Collect(ctx context.Context, cs kubernetes.Interface, obs Observability, alert alertmsg.Alert) Evidence {
	c := &collector{cs: cs, obs: obs}
	var (
		t   Target
		pod *corev1.Pod
	)

	c.run("A1", "Parse alert and resolve target", "Alert payload + K8s API", func() (any, error) {
		var err error
		t, err = c.pickTarget(ctx, alert)
		return t, err
	})
	ns := t.Namespace

	c.run("A2", "Pod and container status", "K8s API: pods", func() (any, error) {
		c.calls++
		p, err := cs.CoreV1().Pods(ns).Get(ctx, t.Pod, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		pod = p
		return podStatus(p), nil
	})

	c.run("A3", "Events for the pod", "K8s API: events", func() (any, error) {
		c.calls++
		list, err := cs.CoreV1().Events(ns).List(ctx, metav1.ListOptions{FieldSelector: "involvedObject.name=" + t.Pod})
		if err != nil {
			return nil, err
		}
		items := list.Items
		sort.SliceStable(items, func(i, j int) bool { return eventTime(items[i]) < eventTime(items[j]) })
		if len(items) > 15 {
			items = items[len(items)-15:]
		}
		out := make([]string, 0, len(items))
		for _, e := range items {
			count := e.Count
			if count == 0 {
				count = 1
			}
			out = append(out, fmt.Sprintf("%s %s %s x%d: %s", eventTime(e), e.Type, e.Reason, count, e.Message))
		}
		return out, nil
	})

	container := t.Container
	if container == "" && pod != nil && len(pod.Spec.Containers) > 0 {
		container = pod.Spec.Containers[0].Name
	}

	c.run("A4", "Container logs (current and previous)", "K8s API: pods/log", func() (any, error) {
		tail := int64(LogTail)
		c.calls++
		cur, err := cs.CoreV1().Pods(ns).GetLogs(t.Pod, &corev1.PodLogOptions{Container: container, TailLines: &tail}).DoRaw(ctx)
		if err != nil {
			return nil, err
		}
		out := map[string]any{"current": lines(string(cur))}
		c.calls++
		prev, err := cs.CoreV1().Pods(ns).GetLogs(t.Pod, &corev1.PodLogOptions{Container: container, Previous: true, TailLines: &tail}).DoRaw(ctx)
		if err != nil {
			out["previous"] = "unavailable: " + truncate(lastLine(err.Error()), 200)
		} else {
			out["previous"] = lines(string(prev))
		}
		return out, nil
	})

	c.run("A5", "Owner, rollout history and recent change", "K8s API: replicasets, deployments", func() (any, error) {
		return c.rollout(ctx, pod)
	})

	sel := fmt.Sprintf(`namespace=%q, pod=%q`, ns, t.Pod)
	c.run("A6", "Restarts and last termination reason", "Prometheus (kube-state-metrics)", func() (any, error) {
		restarts, err := c.prom(ctx, fmt.Sprintf("increase(kube_pod_container_status_restarts_total{%s}[%s])", sel, lookback))
		if err != nil {
			return nil, err
		}
		reason, err := c.prom(ctx, fmt.Sprintf("kube_pod_container_status_last_terminated_reason{%s} == 1", sel))
		if err != nil {
			return nil, err
		}
		return map[string]any{"restarts_15m": restarts, "last_terminated_reason": reason}, nil
	})

	c.run("A7", "Resource usage vs. limits", "Prometheus (cAdvisor + kube-state-metrics)", func() (any, error) {
		mem, err := c.prom(ctx, fmt.Sprintf(`max_over_time(container_memory_working_set_bytes{%s, container!=""}[%s])`, sel, lookback))
		if err != nil {
			return nil, err
		}
		limit, err := c.prom(ctx, fmt.Sprintf(`kube_pod_container_resource_limits{%s, resource="memory"}`, sel))
		if err != nil {
			return nil, err
		}
		cpu, err := c.prom(ctx, fmt.Sprintf(`rate(container_cpu_usage_seconds_total{%s, container!=""}[5m])`, sel))
		if err != nil {
			return nil, err
		}
		return map[string]any{"memory_working_set_max_15m_bytes": mem, "memory_limit_bytes": limit, "cpu_usage_cores_5m": cpu}, nil
	})

	c.run("A8", "Service endpoints of the workload", "K8s API: services, endpointslices", func() (any, error) {
		return c.endpoints(ctx, pod)
	})

	c.run("A9", "Node conditions", "K8s API: nodes", func() (any, error) {
		if pod == nil || pod.Spec.NodeName == "" {
			return "pod not scheduled", nil
		}
		c.calls++
		n, err := cs.CoreV1().Nodes().Get(ctx, pod.Spec.NodeName, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		conds := map[string]string{}
		for _, x := range n.Status.Conditions {
			conds[string(x.Type)] = string(x.Status)
		}
		return map[string]any{"node": n.Name, "conditions": conds}, nil
	})

	c.run("A10", "Log and event history (Loki)", "Loki", func() (any, error) {
		logs, err := c.loki(ctx, fmt.Sprintf(`{namespace=%q, pod=%q}`, ns, t.Pod))
		if err != nil {
			return nil, err
		}
		events, err := c.loki(ctx, fmt.Sprintf(`{job="integrations/kubernetes/eventhandler"} |= %q`, t.Pod))
		if err != nil {
			return nil, err
		}
		return map[string]any{"pod_logs": logs, "events": events}, nil
	})

	c.run("A11", "Service golden signals (RED)", "Prometheus (application metrics)", func() (any, error) {
		s := sel
		if pod != nil {
			if app := pod.Labels["app"]; app != "" {
				s = fmt.Sprintf(`namespace=%q, pod=~"%s-.*"`, ns, app)
			}
		}
		s += `, path!~"healthz|readyz"` // probe traffic is not user traffic
		rate, err := c.prom(ctx, fmt.Sprintf("sum(rate(http_request_duration_seconds_count{%s}[5m]))", s))
		if err != nil {
			return nil, err
		}
		errs, err := c.prom(ctx, fmt.Sprintf(`sum(rate(http_request_duration_seconds_count{%s, status=~"5.."}[5m]))`, s))
		if err != nil {
			return nil, err
		}
		p95, err := c.prom(ctx, fmt.Sprintf("histogram_quantile(0.95, sum by (le) (rate(http_request_duration_seconds_bucket{%s}[5m])))", s))
		if err != nil {
			return nil, err
		}
		return map[string]any{"request_rate_rps": rate, "error_rate_5xx_rps": errs, "latency_p95_s": p95}, nil
	})

	total := 0.0
	for _, a := range c.actions {
		total += a.DurationS
	}
	return Evidence{Target: t, TotalCalls: c.calls, TotalDurationS: round(total, 2), Actions: c.actions}
}

func (c *collector) run(id, name, source string, fn func() (any, error)) {
	start, before := time.Now(), c.calls
	out, err := fn()
	ok := err == nil
	if err != nil {
		out = "ERROR: " + truncate(lastLine(err.Error()), 300)
	}
	c.actions = append(c.actions, Action{
		ID: id, Name: name, Source: source, OK: ok, Calls: c.calls - before,
		DurationS: round(time.Since(start).Seconds(), 2), Output: out,
	})
}

// Every read-only Prometheus and Loki call goes through prom and loki so it is counted.
func (c *collector) prom(ctx context.Context, expr string) ([]map[string]any, error) {
	c.calls++
	rows, err := c.obs.PromQuery(ctx, expr)
	if err != nil {
		return nil, err
	}
	out := make([]map[string]any, 0, len(rows))
	for _, r := range rows {
		m := map[string]any{}
		for k, v := range r.Metric {
			m[k] = v
		}
		m["value"] = r.Value
		out = append(out, m)
	}
	return out, nil
}

func (c *collector) loki(ctx context.Context, logql string) ([]string, error) {
	c.calls++
	rows, err := c.obs.LokiQuery(ctx, logql, 15*time.Minute, LogTail)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(rows))
	for _, r := range rows {
		out = append(out, r.Line)
	}
	return out, nil
}

// pickTarget is A1: resolve the alert's labels to a namespace, pod and container.
func (c *collector) pickTarget(ctx context.Context, alert alertmsg.Alert) (Target, error) {
	l := alert.Labels
	t := Target{
		Alertname: l["alertname"], Namespace: l["namespace"], Pod: l["pod"], Container: l["container"],
		Deployment: l["deployment"], Summary: alert.Annotations["summary"], StartsAt: alert.StartsAt,
	}
	var podSelector string
	switch {
	case t.Pod != "":
	case t.Deployment != "":
		// Deployment-level alert: inspect its least healthy pod.
		c.calls++
		dep, err := c.cs.AppsV1().Deployments(t.Namespace).Get(ctx, t.Deployment, metav1.GetOptions{})
		if err != nil {
			return t, err
		}
		podSelector = selector(dep)
	case l["service"] != "":
		// Service-level alert (e.g. latency or traffic of a service): inspect
		// the least healthy pod behind the service.
		t.Service = l["service"]
		c.calls++
		svc, err := c.cs.CoreV1().Services(t.Namespace).Get(ctx, t.Service, metav1.GetOptions{})
		if err != nil {
			return t, err
		}
		podSelector = labels.SelectorFromSet(svc.Spec.Selector).String()
	}
	if t.Pod == "" && podSelector != "" {
		c.calls++
		pods, err := c.cs.CoreV1().Pods(t.Namespace).List(ctx, metav1.ListOptions{LabelSelector: podSelector})
		if err != nil {
			return t, err
		}
		score := func(p corev1.Pod) int {
			n := 0
			for _, s := range p.Status.ContainerStatuses {
				n += int(s.RestartCount)
			}
			if p.Status.Phase != corev1.PodRunning {
				n += 1000
			}
			return n
		}
		items := pods.Items
		sort.SliceStable(items, func(i, j int) bool {
			if score(items[i]) != score(items[j]) {
				return score(items[i]) > score(items[j])
			}
			return items[i].Name < items[j].Name
		})
		if len(items) > 0 {
			t.Pod = items[0].Name
		}
	}
	if t.Pod == "" {
		return t, fmt.Errorf("alert has no pod, deployment or service label")
	}
	return t, nil
}

func podStatus(p *corev1.Pod) map[string]any {
	conds := map[string]string{}
	for _, x := range p.Status.Conditions {
		conds[string(x.Type)] = string(x.Status)
	}
	var containers []map[string]any
	for _, s := range append(append([]corev1.ContainerStatus{}, p.Status.InitContainerStatuses...), p.Status.ContainerStatuses...) {
		containers = append(containers, map[string]any{
			"name": s.Name, "ready": s.Ready, "restartCount": s.RestartCount, "image": s.Image,
			"state": stateView(s.State), "lastState": stateView(s.LastTerminationState),
		})
	}
	resources := map[string]corev1.ResourceRequirements{}
	for _, x := range p.Spec.Containers {
		resources[x.Name] = x.Resources
	}
	return map[string]any{"phase": p.Status.Phase, "node": p.Spec.NodeName, "conditions": conds,
		"containers": containers, "resources": resources}
}

func stateView(s corev1.ContainerState) map[string]any {
	ts := func(t metav1.Time) string { return t.UTC().Format(time.RFC3339) }
	switch {
	case s.Waiting != nil:
		return omitEmpty(map[string]any{"state": "waiting", "reason": s.Waiting.Reason, "message": s.Waiting.Message})
	case s.Running != nil:
		return map[string]any{"state": "running", "startedAt": ts(s.Running.StartedAt)}
	case s.Terminated != nil:
		t := s.Terminated
		return omitEmpty(map[string]any{"state": "terminated", "reason": t.Reason, "exitCode": t.ExitCode,
			"message": t.Message, "startedAt": ts(t.StartedAt), "finishedAt": ts(t.FinishedAt)})
	}
	return nil
}

func omitEmpty(m map[string]any) map[string]any {
	for k, v := range m {
		if s, ok := v.(string); ok && s == "" {
			delete(m, k)
		}
	}
	return m
}

func eventTime(e corev1.Event) string {
	switch {
	case !e.LastTimestamp.IsZero():
		return e.LastTimestamp.UTC().Format(time.RFC3339)
	case !e.EventTime.IsZero():
		return e.EventTime.UTC().Format(time.RFC3339)
	}
	return ""
}

func revision(rs appsv1.ReplicaSet) int {
	var n int
	fmt.Sscanf(rs.Annotations["deployment.kubernetes.io/revision"], "%d", &n)
	return n
}

func selector(dep *appsv1.Deployment) string {
	return labels.SelectorFromSet(dep.Spec.Selector.MatchLabels).String()
}

// rollout is A5: owner chain, recent revisions and what changed in the latest one.
func (c *collector) rollout(ctx context.Context, pod *corev1.Pod) (any, error) {
	if pod == nil {
		return nil, fmt.Errorf("pod unavailable")
	}
	rsName := ""
	for _, o := range pod.OwnerReferences {
		if o.Kind == "ReplicaSet" {
			rsName = o.Name
			break
		}
	}
	if rsName == "" {
		return map[string]any{"owner": pod.OwnerReferences}, nil
	}
	ns := pod.Namespace
	c.calls++
	rs, err := c.cs.AppsV1().ReplicaSets(ns).Get(ctx, rsName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	depName := ""
	for _, o := range rs.OwnerReferences {
		if o.Kind == "Deployment" {
			depName = o.Name
			break
		}
	}
	c.calls++
	dep, err := c.cs.AppsV1().Deployments(ns).Get(ctx, depName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	c.calls++
	list, err := c.cs.AppsV1().ReplicaSets(ns).List(ctx, metav1.ListOptions{LabelSelector: selector(dep)})
	if err != nil {
		return nil, err
	}
	sets := list.Items
	sort.SliceStable(sets, func(i, j int) bool { return revision(sets[i]) > revision(sets[j]) })

	st := dep.Status
	conds := []string{}
	for _, x := range st.Conditions {
		conds = append(conds, fmt.Sprintf("%s=%s (%s): %s", x.Type, x.Status, x.Reason, x.Message))
	}
	var revs []map[string]any
	for i, r := range sets {
		if i == 3 {
			break
		}
		var cause any
		if v, ok := r.Annotations["kubernetes.io/change-cause"]; ok {
			cause = v
		}
		revs = append(revs, map[string]any{"revision": revision(r), "name": r.Name,
			"created": r.CreationTimestamp.UTC().Format(time.RFC3339), "replicas": r.Status.Replicas, "change_cause": cause})
	}
	result := map[string]any{
		"deployment": depName,
		"replicas": map[string]int32{"replicas": st.Replicas, "readyReplicas": st.ReadyReplicas,
			"availableReplicas": st.AvailableReplicas, "unavailableReplicas": st.UnavailableReplicas,
			"updatedReplicas": st.UpdatedReplicas},
		"conditions": conds,
		"revisions":  revs,
	}
	if len(sets) >= 2 {
		result["template_diff_latest_vs_previous"] = DiffTemplates(templateView(&sets[0]), templateView(&sets[1]))
	}
	return result, nil
}

type containerView map[string]any

func templateView(rs *appsv1.ReplicaSet) map[string]containerView {
	out := map[string]containerView{}
	for _, c := range rs.Spec.Template.Spec.Containers {
		out[c.Name] = containerView{"image": c.Image, "command": nilIfEmpty(c.Command), "args": nilIfEmpty(c.Args),
			"env": nilIfEmpty(c.Env), "resources": c.Resources, "envFrom": nilIfEmpty(c.EnvFrom)}
	}
	return out
}

func nilIfEmpty[T any](s []T) any {
	if len(s) == 0 {
		return nil
	}
	return s
}

// Change is one field that differs between two ReplicaSet templates.
type Change struct {
	Container string `json:"container"`
	Field     string `json:"field"`
	Previous  any    `json:"previous"`
	Current   any    `json:"current"`
}

// DiffTemplates lists container fields whose value differs between the newer and the older template.
func DiffTemplates(newer, older map[string]containerView) []Change {
	changes := []Change{}
	names := make([]string, 0, len(newer))
	for n := range newer {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, name := range names {
		fields := make([]string, 0, len(newer[name]))
		for f := range newer[name] {
			fields = append(fields, f)
		}
		sort.Strings(fields)
		for _, f := range fields {
			cur, prev := newer[name][f], older[name][f]
			if jsonString(cur) != jsonString(prev) {
				changes = append(changes, Change{Container: name, Field: f, Previous: prev, Current: cur})
			}
		}
	}
	return changes
}

func jsonString(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

// endpoints is A8: which services select the pod and how many endpoints are ready.
func (c *collector) endpoints(ctx context.Context, pod *corev1.Pod) (any, error) {
	if pod == nil {
		return nil, fmt.Errorf("pod unavailable")
	}
	c.calls++
	svcs, err := c.cs.CoreV1().Services(pod.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	out := []map[string]any{}
	for _, svc := range svcs.Items {
		if len(svc.Spec.Selector) == 0 || !labels.SelectorFromSet(svc.Spec.Selector).Matches(labels.Set(pod.Labels)) {
			continue
		}
		c.calls++
		slices, err := c.cs.DiscoveryV1().EndpointSlices(pod.Namespace).List(ctx,
			metav1.ListOptions{LabelSelector: "kubernetes.io/service-name=" + svc.Name})
		if err != nil {
			return nil, err
		}
		total, ready := 0, 0
		for _, s := range slices.Items {
			for _, e := range s.Endpoints {
				total++
				if e.Conditions.Ready != nil && *e.Conditions.Ready {
					ready++
				}
			}
		}
		out = append(out, map[string]any{"service": svc.Name, "endpoints": total, "ready": ready})
	}
	return out, nil
}

// Markdown renders the evidence for humans.
func (e Evidence) Markdown() string {
	t := e.Target
	var b strings.Builder
	fmt.Fprintf(&b, "# Evidence: %s %s/%s\n\n- Read-only calls: **%d**, collection time: **%v s**\n\n",
		t.Alertname, t.Namespace, t.Pod, e.TotalCalls, e.TotalDurationS)
	for _, a := range e.Actions {
		body, _ := json.MarshalIndent(a.Output, "", "  ")
		fmt.Fprintf(&b, "## %s %s (%s; %d calls, %v s)\n\n```json\n%s\n```\n\n", a.ID, a.Name, a.Source, a.Calls, a.DurationS,
			truncate(string(body), 4000))
	}
	return b.String()
}

func round(v float64, digits int) float64 {
	p := math.Pow(10, float64(digits))
	return math.Round(v*p) / p
}

func lines(s string) []string {
	s = strings.TrimRight(s, "\n")
	if s == "" {
		return []string{}
	}
	return strings.Split(s, "\n")
}

func lastLine(s string) string {
	parts := strings.Split(strings.TrimSpace(s), "\n")
	return parts[len(parts)-1]
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n])
}
