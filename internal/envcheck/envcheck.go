// Package envcheck verifies that the experimental environment is ready. Each
// check reports PASS/FAIL; the last one sends a synthetic alert to Alertmanager
// and waits for it at the receiver, which proves the whole notification path
// (Alertmanager route -> webhook -> receiver) works.
package envcheck

import (
	"context"
	"fmt"
	"strings"
	"time"

	authv1 "k8s.io/api/authorization/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/nojussta/k8s-diag-agent/internal/lab"
)

var namespaces = []string{"monitoring", "logging", "demo", "diag-agent"}

// Result is the outcome of one check.
type Result struct {
	Name   string
	OK     bool
	Detail string
}

type check struct {
	name string
	fn   func(ctx context.Context, l *lab.Lab) (string, error)
}

// Run executes all checks in order, reporting each through report as it finishes.
func Run(ctx context.Context, l *lab.Lab, report func(Result)) []Result {
	var results []Result
	for _, c := range checks {
		detail, err := c.fn(ctx, l)
		r := Result{Name: c.name, OK: err == nil, Detail: detail}
		if err != nil {
			r.Detail = firstLine(err.Error(), 200)
		}
		results = append(results, r)
		if report != nil {
			report(r)
		}
	}
	return results
}

var checks = []check{
	{"Nodes are Ready", func(ctx context.Context, l *lab.Lab) (string, error) {
		nodes, err := l.Cs.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
		if err != nil {
			return "", err
		}
		ready := 0
		for _, n := range nodes.Items {
			for _, c := range n.Status.Conditions {
				if c.Type == corev1.NodeReady && c.Status == corev1.ConditionTrue {
					ready++
				}
			}
		}
		if ready != len(nodes.Items) || ready == 0 {
			return "", fmt.Errorf("%d/%d ready", ready, len(nodes.Items))
		}
		return fmt.Sprintf("%d nodes", ready), nil
	}},
	{"Lab pods are Running and Ready", func(ctx context.Context, l *lab.Lab) (string, error) {
		var bad []string
		total := 0
		for _, ns := range namespaces {
			pods, err := l.Cs.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{})
			if err != nil {
				return "", err
			}
			for _, p := range pods.Items {
				if p.Status.Phase == corev1.PodSucceeded {
					continue
				}
				total++
				ready := p.Status.Phase == corev1.PodRunning
				for _, s := range p.Status.ContainerStatuses {
					ready = ready && s.Ready
				}
				if !ready {
					bad = append(bad, ns+"/"+p.Name)
				}
			}
		}
		if len(bad) > 0 {
			return "", fmt.Errorf("not ready: %s", strings.Join(bad, ", "))
		}
		return fmt.Sprintf("%d pods", total), nil
	}},
	{"Prometheus scrapes the demo app", func(ctx context.Context, l *lab.Lab) (string, error) {
		up, err := l.PromQuery(ctx, `up{namespace="demo"}`)
		if err != nil {
			return "", err
		}
		if len(up) == 0 {
			return "", fmt.Errorf("no demo targets")
		}
		var down []string
		for _, r := range up {
			if r.Value != 1 {
				down = append(down, r.Metric["pod"])
			}
		}
		if len(down) > 0 {
			return "", fmt.Errorf("targets down: %v", down)
		}
		return fmt.Sprintf("%d targets up", len(up)), nil
	}},
	{"Prometheus has kube-state-metrics data", func(ctx context.Context, l *lab.Lab) (string, error) {
		rows, err := l.PromQuery(ctx, `kube_pod_container_status_restarts_total{namespace="demo"}`)
		if err != nil {
			return "", err
		}
		if len(rows) == 0 {
			return "", fmt.Errorf("no kube-state-metrics series for demo")
		}
		return fmt.Sprintf("%d container series", len(rows)), nil
	}},
	{"Lab alert rules are loaded", func(ctx context.Context, l *lab.Lab) (string, error) {
		names, err := l.RuleNames(ctx, "lab.kursinis.workloads")
		if err != nil {
			return "", err
		}
		if len(names) == 0 {
			return "", fmt.Errorf("group lab.kursinis.workloads not found")
		}
		return strings.Join(names, ", "), nil
	}},
	{"No lab alerts firing in the baseline", func(ctx context.Context, l *lab.Lab) (string, error) {
		firing, err := l.FiringLabAlerts(ctx)
		if err != nil {
			return "", err
		}
		if len(firing) > 0 {
			return "", fmt.Errorf("firing: %v", firing)
		}
		return "", nil
	}},
	{"Loki receives demo app logs", func(ctx context.Context, l *lab.Lab) (string, error) {
		rows, err := l.LokiQuery(ctx, `{namespace="demo"}`, 5*time.Minute, 50)
		if err != nil {
			return "", err
		}
		if len(rows) == 0 {
			return "", fmt.Errorf("no log lines for namespace=demo in the last 5 min")
		}
		return fmt.Sprintf("%d lines (sample)", len(rows)), nil
	}},
	{"Loki receives Kubernetes events", func(ctx context.Context, l *lab.Lab) (string, error) {
		rows, err := l.LokiQuery(ctx, `{job="integrations/kubernetes/eventhandler"}`, time.Hour, 20)
		if err != nil {
			return "", err
		}
		if len(rows) == 0 {
			return "", fmt.Errorf("no events stream")
		}
		return fmt.Sprintf("%d events (sample)", len(rows)), nil
	}},
	{"Agent ServiceAccount is read-only (RBAC)", rbacCheck},
	{"Alertmanager routes lab alerts to the receiver", func(ctx context.Context, l *lab.Lab) (string, error) {
		// The webhook URL itself is masked as a secret in /api/v2/status, so
		// check the receiver and the route instead.
		cfg, err := l.AlertmanagerConfig(ctx)
		if err != nil {
			return "", err
		}
		if !strings.Contains(cfg, "name: diag-agent") || !strings.Contains(cfg, "webhook_configs") {
			return "", fmt.Errorf("diag-agent webhook receiver missing")
		}
		if !strings.Contains(cfg, `lab="kursinis"`) {
			return "", fmt.Errorf(`route for lab="kursinis" missing`)
		}
		return "", nil
	}},
	{"End-to-end: synthetic alert reaches the receiver", endToEnd},
}

type perm struct {
	verb, group, resource, sub string
	allowed                    bool
}

func (p perm) String() string {
	r := p.resource
	if p.sub != "" {
		r += "/" + p.sub
	}
	return p.verb + " " + r
}

var expectations = []perm{
	{"get", "", "pods", "", true}, {"get", "", "pods", "log", true}, {"list", "", "events", "", true},
	{"list", "apps", "replicasets", "", true}, {"get", "", "nodes", "", true},
	{"delete", "", "pods", "", false}, {"patch", "apps", "deployments", "", false},
	{"create", "", "pods", "exec", false}, {"create", "", "pods", "portforward", false},
	{"get", "", "secrets", "", false}, {"update", "", "configmaps", "", false},
}

func rbacCheck(ctx context.Context, l *lab.Lab) (string, error) {
	const sa = "system:serviceaccount:diag-agent:diag-agent"
	var wrong []string
	for _, p := range expectations {
		review := &authv1.SubjectAccessReview{Spec: authv1.SubjectAccessReviewSpec{
			User:   sa,
			Groups: []string{"system:serviceaccounts", "system:serviceaccounts:diag-agent", "system:authenticated"},
			ResourceAttributes: &authv1.ResourceAttributes{
				Namespace: "demo", Verb: p.verb, Group: p.group, Resource: p.resource, Subresource: p.sub},
		}}
		res, err := l.Cs.AuthorizationV1().SubjectAccessReviews().Create(ctx, review, metav1.CreateOptions{})
		if err != nil {
			return "", err
		}
		if res.Status.Allowed != p.allowed {
			wrong = append(wrong, p.String())
		}
	}
	if len(wrong) > 0 {
		return "", fmt.Errorf("unexpected permissions: %s", strings.Join(wrong, ", "))
	}
	return fmt.Sprintf("%d permissions as expected (no write, exec or secrets)", len(expectations)), nil
}

func endToEnd(ctx context.Context, l *lab.Lab) (string, error) {
	start := time.Now()
	_, err := l.ProxyPostJSON(ctx, lab.Alertmanager, "api/v2/alerts", []map[string]any{{
		"labels":      map[string]string{"alertname": "LabEnvSelfTest", "lab": "kursinis", "severity": "none", "namespace": "diag-agent"},
		"annotations": map[string]string{"summary": "Synthetic alert sent by labctl check"},
		"endsAt":      start.Add(60 * time.Second).UTC().Format(time.RFC3339),
	}})
	if err != nil {
		return "", err
	}
	since := float64(start.UnixNano()) / 1e9
	deadline := start.Add(90 * time.Second)
	for time.Now().Before(deadline) {
		ns, err := l.Notifications(ctx, since)
		if err == nil {
			for _, n := range ns {
				for _, a := range n.Payload.Alerts {
					if a.Labels["alertname"] == "LabEnvSelfTest" {
						return fmt.Sprintf("delivered in %.1fs", n.ReceivedAt-since), nil
					}
				}
			}
		}
		select {
		case <-time.After(2 * time.Second):
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
	return "", fmt.Errorf("synthetic alert not received within 90s")
}

func firstLine(s string, n int) string {
	s = strings.SplitN(strings.TrimSpace(s), "\n", 2)[0]
	if len(s) > n {
		s = s[:n]
	}
	return s
}
