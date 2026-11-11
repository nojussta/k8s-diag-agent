// Package tools is the agent's tool layer: ten read-only functions that map
// onto the diagnostic action set A1-A11 (table "MCP server
// tools"). The package has no write path: it only holds a
// kubernetes.Interface and an Observability handle, and it only calls get,
// list and query operations. Every result is size-limited before it reaches
// the model.
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/nojussta/k8s-diag-agent/internal/evidence"
)

// Limits of the tool layer.
const (
	MaxBytes   = 8 * 1024
	MaxPods    = 50
	MaxEvents  = 20
	MaxLogs    = 50
	MaxLokiMin = 60
	MaxRows    = 20
)

// Toolset holds the read-only handles.
type Toolset struct {
	CS  kubernetes.Interface
	Obs evidence.Observability
}

// Result is the uniform envelope every tool returns.
type Result struct {
	OK        bool   `json:"ok"`
	Source    string `json:"source"`
	Truncated bool   `json:"truncated"`
	Data      any    `json:"data,omitempty"`
	Error     string `json:"error,omitempty"`
}

// Name of every tool, in the order of the action set.
var Names = []string{
	"get_pod_status", "list_pods", "get_events", "get_container_logs", "get_rollout_history",
	"get_resource_usage", "get_service_endpoints", "get_node_status", "query_logs", "query_metrics",
}

var k8sName = regexp.MustCompile(`^[a-z0-9]([-a-z0-9.]*[a-z0-9])?$`)

func validName(field, v string) error {
	if len(v) > 253 || !k8sName.MatchString(v) {
		return fmt.Errorf("invalid %s %q", field, v)
	}
	return nil
}

func ok(source string, data any) Result { return finish(Result{OK: true, Source: source, Data: data}) }

func fail(source string, err error) Result {
	return Result{OK: false, Source: source, Error: truncateStr(err.Error(), 300)}
}

// finish enforces the 8 KB cap. Lists lose their oldest items first; anything
// else is replaced by a short marker so the model never gets broken JSON.
func finish(r Result) Result {
	for size(r) > MaxBytes {
		r.Truncated = true
		switch d := r.Data.(type) {
		case []string:
			if len(d) <= 1 {
				r.Data = []string{truncateStr(strings.Join(d, "\n"), MaxBytes/2)}
			} else {
				r.Data = d[len(d)/2:]
			}
		case []map[string]any:
			if len(d) <= 1 {
				r.Data = "result too large, narrow the query"
			} else {
				r.Data = d[:len(d)/2]
			}
		default:
			r.Data = "result too large, narrow the query"
		}
	}
	return r
}

func size(r Result) int { b, _ := json.Marshal(r); return len(b) }

func truncateStr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// GetPodStatus is A2.
func (t *Toolset) GetPodStatus(ctx context.Context, ns, pod string) Result {
	const src = "K8s API: pods"
	if err := firstErr(validName("namespace", ns), validName("pod", pod)); err != nil {
		return fail(src, err)
	}
	p, err := t.CS.CoreV1().Pods(ns).Get(ctx, pod, metav1.GetOptions{})
	if err != nil {
		return fail(src, err)
	}
	return ok(src, evidence.PodStatus(p))
}

// ListPods is A1 (finding the right pod).
func (t *Toolset) ListPods(ctx context.Context, ns, labelSelector string) Result {
	const src = "K8s API: pods"
	if err := validName("namespace", ns); err != nil {
		return fail(src, err)
	}
	list, err := t.CS.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{LabelSelector: labelSelector, Limit: MaxPods})
	if err != nil {
		return fail(src, err)
	}
	out := []map[string]any{}
	for _, p := range list.Items {
		if len(out) == MaxPods {
			break
		}
		restarts := int32(0)
		for _, s := range p.Status.ContainerStatuses {
			restarts += s.RestartCount
		}
		out = append(out, map[string]any{"name": p.Name, "phase": p.Status.Phase, "restarts": restarts, "node": p.Spec.NodeName})
	}
	return ok(src, out)
}

// GetEvents is A3.
func (t *Toolset) GetEvents(ctx context.Context, ns, object string) Result {
	const src = "K8s API: events"
	if err := firstErr(validName("namespace", ns), validName("object", object)); err != nil {
		return fail(src, err)
	}
	list, err := t.CS.CoreV1().Events(ns).List(ctx, metav1.ListOptions{FieldSelector: "involvedObject.name=" + object})
	if err != nil {
		return fail(src, err)
	}
	items := list.Items
	sort.SliceStable(items, func(i, j int) bool { return evTime(items[i]) < evTime(items[j]) })
	if len(items) > MaxEvents {
		items = items[len(items)-MaxEvents:]
	}
	out := make([]string, 0, len(items))
	for _, e := range items {
		n := max(e.Count, 1)
		out = append(out, fmt.Sprintf("%s %s %s x%d: %s", evTime(e), e.Type, e.Reason, n, e.Message))
	}
	return ok(src, out)
}

func evTime(e corev1.Event) string {
	for _, t := range []time.Time{e.LastTimestamp.Time, e.EventTime.Time, e.FirstTimestamp.Time, e.CreationTimestamp.Time} {
		if !t.IsZero() {
			return t.UTC().Format(time.RFC3339)
		}
	}
	return ""
}

// GetContainerLogs is A4.
func (t *Toolset) GetContainerLogs(ctx context.Context, ns, pod, container string, previous bool, tail int) Result {
	const src = "K8s API: pods/log"
	if err := firstErr(validName("namespace", ns), validName("pod", pod)); err != nil {
		return fail(src, err)
	}
	if container != "" {
		if err := validName("container", container); err != nil {
			return fail(src, err)
		}
	}
	if tail <= 0 || tail > MaxLogs {
		tail = MaxLogs
	}
	n := int64(tail)
	raw, err := t.CS.CoreV1().Pods(ns).GetLogs(pod, &corev1.PodLogOptions{Container: container, Previous: previous, TailLines: &n}).DoRaw(ctx)
	if err != nil {
		return fail(src, err)
	}
	return ok(src, splitLines(string(raw)))
}

// GetRolloutHistory is A5.
func (t *Toolset) GetRolloutHistory(ctx context.Context, ns, pod string) Result {
	const src = "K8s API: replicasets, deployments"
	if err := firstErr(validName("namespace", ns), validName("pod", pod)); err != nil {
		return fail(src, err)
	}
	p, err := t.CS.CoreV1().Pods(ns).Get(ctx, pod, metav1.GetOptions{})
	if err != nil {
		return fail(src, err)
	}
	d, err := evidence.Rollout(ctx, t.CS, p)
	if err != nil {
		return fail(src, err)
	}
	return ok(src, d)
}

// GetResourceUsage is A6 and A7 with fixed PromQL (the model supplies only names).
func (t *Toolset) GetResourceUsage(ctx context.Context, ns, pod string) Result {
	const src = "Prometheus"
	if err := firstErr(validName("namespace", ns), validName("pod", pod)); err != nil {
		return fail(src, err)
	}
	sel := fmt.Sprintf(`namespace=%q, pod=%q`, ns, pod)
	queries := map[string]string{
		"restarts_15m":                     fmt.Sprintf("increase(kube_pod_container_status_restarts_total{%s}[15m])", sel),
		"last_terminated_reason":           fmt.Sprintf("kube_pod_container_status_last_terminated_reason{%s} == 1", sel),
		"memory_working_set_max_15m_bytes": fmt.Sprintf(`max_over_time(container_memory_working_set_bytes{%s, container!=""}[15m])`, sel),
		"memory_limit_bytes":               fmt.Sprintf(`kube_pod_container_resource_limits{%s, resource="memory"}`, sel),
		"cpu_usage_cores_5m":               fmt.Sprintf(`rate(container_cpu_usage_seconds_total{%s, container!=""}[5m])`, sel),
	}
	out := map[string]any{}
	for k, q := range queries {
		rows, err := t.prom(ctx, q)
		if err != nil {
			return fail(src, err)
		}
		out[k] = rows
	}
	return ok(src, out)
}

// GetServiceEndpoints is A8.
func (t *Toolset) GetServiceEndpoints(ctx context.Context, ns, pod string) Result {
	const src = "K8s API: services, endpointslices"
	if err := firstErr(validName("namespace", ns), validName("pod", pod)); err != nil {
		return fail(src, err)
	}
	p, err := t.CS.CoreV1().Pods(ns).Get(ctx, pod, metav1.GetOptions{})
	if err != nil {
		return fail(src, err)
	}
	d, err := evidence.Endpoints(ctx, t.CS, p)
	if err != nil {
		return fail(src, err)
	}
	return ok(src, d)
}

// GetNodeStatus is A9.
func (t *Toolset) GetNodeStatus(ctx context.Context, node string) Result {
	const src = "K8s API: nodes"
	if err := validName("node", node); err != nil {
		return fail(src, err)
	}
	n, err := t.CS.CoreV1().Nodes().Get(ctx, node, metav1.GetOptions{})
	if err != nil {
		return fail(src, err)
	}
	conds := map[string]string{}
	for _, c := range n.Status.Conditions {
		conds[string(c.Type)] = string(c.Status)
	}
	capacity := map[string]string{}
	for k, v := range n.Status.Allocatable {
		capacity[string(k)] = v.String()
	}
	return ok(src, map[string]any{"node": n.Name, "conditions": conds, "allocatable": capacity})
}

// QueryLogs is A10: a LogQL query over at most 60 minutes and 50 lines.
func (t *Toolset) QueryLogs(ctx context.Context, logql string, minutes, limit int) Result {
	const src = "Loki"
	if strings.TrimSpace(logql) == "" || len(logql) > 500 {
		return fail(src, fmt.Errorf("logql must be 1-500 characters"))
	}
	if minutes <= 0 || minutes > MaxLokiMin {
		minutes = 15
	}
	if limit <= 0 || limit > MaxLogs {
		limit = MaxLogs
	}
	rows, err := t.Obs.LokiQuery(ctx, logql, time.Duration(minutes)*time.Minute, limit)
	if err != nil {
		return fail(src, err)
	}
	out := make([]string, 0, len(rows))
	for _, r := range rows {
		out = append(out, r.Line)
	}
	return ok(src, out)
}

// QueryMetrics is A11 and free-form PromQL (instant query, at most 20 rows).
func (t *Toolset) QueryMetrics(ctx context.Context, promql string) Result {
	const src = "Prometheus"
	if strings.TrimSpace(promql) == "" || len(promql) > 500 {
		return fail(src, fmt.Errorf("promql must be 1-500 characters"))
	}
	rows, err := t.prom(ctx, promql)
	if err != nil {
		return fail(src, err)
	}
	return ok(src, rows)
}

func (t *Toolset) prom(ctx context.Context, expr string) ([]map[string]any, error) {
	rows, err := t.Obs.PromQuery(ctx, expr)
	if err != nil {
		return nil, err
	}
	out := []map[string]any{}
	for _, r := range rows {
		if len(out) == MaxRows {
			break
		}
		m := map[string]any{"value": r.Value}
		for k, v := range r.Metric {
			m[k] = v
		}
		out = append(out, m)
	}
	return out, nil
}

func splitLines(s string) []string {
	s = strings.TrimRight(s, "\n")
	if s == "" {
		return []string{}
	}
	return strings.Split(s, "\n")
}

func firstErr(errs ...error) error {
	for _, e := range errs {
		if e != nil {
			return e
		}
	}
	return nil
}
