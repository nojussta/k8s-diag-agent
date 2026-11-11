// Package lab holds the helpers shared by the lab tooling: an isolated
// connection to the kind cluster and typed access to Prometheus, Loki,
// Alertmanager and the alert receiver through the Kubernetes API server
// service proxy (so no port-forwarding is needed and the lab works the same on
// any machine).
package lab

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/nojussta/k8s-diag-agent/internal/alertmsg"
)

// KindContext is the only kube context the lab ever talks to.
const KindContext = "kind-kursinis"

// Svc addresses a Service port reachable through the API server proxy.
type Svc struct{ Namespace, Name, Port string }

func (s Svc) proxyName() string { return s.Name + ":" + s.Port }

var (
	Prometheus   = Svc{"monitoring", "kps-prometheus", "9090"}
	Alertmanager = Svc{"monitoring", "kps-alertmanager", "9093"}
	Loki         = Svc{"logging", "loki", "3100"}
	Receiver     = Svc{"diag-agent", "alert-receiver", "8080"}
)

// Lab is an isolated connection to the lab cluster.
type Lab struct {
	Root       string
	Kubeconfig string
	Cs         kubernetes.Interface
}

// FindRoot walks up from the working directory to the directory holding go.mod
// (the k8s-diag-agent project root). LAB_ROOT overrides it.
func FindRoot() (string, error) {
	if r := os.Getenv("LAB_ROOT"); r != "" {
		return filepath.Abs(r)
	}
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found above %s; run from the k8s-diag-agent directory or set LAB_ROOT", dir)
		}
		dir = parent
	}
}

// New connects using the isolated kubeconfig: $KUBECONFIG when set (the
// Makefile exports it), otherwise <root>/.kube/config. The user's default
// ~/.kube/config is never read, and the context is forced to kind-kursinis, so
// a work cluster can never be reached by accident.
func New() (*Lab, error) {
	root, err := FindRoot()
	if err != nil {
		return nil, err
	}
	path := os.Getenv("KUBECONFIG")
	if path == "" || strings.Contains(path, string(os.PathListSeparator)) {
		path = filepath.Join(root, ".kube", "config")
	}
	cfg, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: path},
		&clientcmd.ConfigOverrides{CurrentContext: KindContext},
	).ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("kubeconfig %s: %w (run `make up` first)", path, err)
	}
	cfg.QPS, cfg.Burst = 50, 100
	cfg.Timeout = 60 * time.Second
	cs, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	return &Lab{Root: root, Kubeconfig: path, Cs: cs}, nil
}

// ProxyGet performs GET <svc>/<path>?<params> via the API server proxy.
func (l *Lab) ProxyGet(ctx context.Context, svc Svc, path string, params url.Values) ([]byte, error) {
	req := l.Cs.CoreV1().RESTClient().Get().
		Namespace(svc.Namespace).Resource("services").Name(svc.proxyName()).SubResource("proxy").
		Suffix(path)
	for k, vs := range params {
		for _, v := range vs {
			req = req.Param(k, v)
		}
	}
	return req.DoRaw(ctx)
}

// ProxyPostJSON performs POST <svc>/<path> with a JSON body via the proxy.
func (l *Lab) ProxyPostJSON(ctx context.Context, svc Svc, path string, body any) ([]byte, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	return l.Cs.CoreV1().RESTClient().Post().
		Namespace(svc.Namespace).Resource("services").Name(svc.proxyName()).SubResource("proxy").
		Suffix(path).SetHeader("Content-Type", "application/json").Body(data).DoRaw(ctx)
}

// Sample is one instant-query result.
type Sample struct {
	Metric map[string]string `json:"metric"`
	Value  float64           `json:"value"`
}

// PromQuery runs an instant PromQL query.
func (l *Lab) PromQuery(ctx context.Context, expr string) ([]Sample, error) {
	raw, err := l.ProxyGet(ctx, Prometheus, "api/v1/query", url.Values{"query": {expr}})
	if err != nil {
		return nil, err
	}
	var res struct {
		Data struct {
			Result []struct {
				Metric map[string]string `json:"metric"`
				Value  [2]any            `json:"value"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &res); err != nil {
		return nil, err
	}
	out := make([]Sample, 0, len(res.Data.Result))
	for _, r := range res.Data.Result {
		s, _ := r.Value[1].(string)
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return nil, fmt.Errorf("bad sample value %q: %w", s, err)
		}
		out = append(out, Sample{Metric: r.Metric, Value: v})
	}
	return out, nil
}

// LogLine is one Loki entry.
type LogLine struct {
	TimestampNs int64
	Labels      map[string]string
	Line        string
}

// LokiQuery runs a LogQL range query and returns lines oldest first.
func (l *Lab) LokiQuery(ctx context.Context, logql string, since time.Duration, limit int) ([]LogLine, error) {
	end := time.Now()
	raw, err := l.ProxyGet(ctx, Loki, "loki/api/v1/query_range", url.Values{
		"query":     {logql},
		"start":     {strconv.FormatInt(end.Add(-since).UnixNano(), 10)},
		"end":       {strconv.FormatInt(end.UnixNano(), 10)},
		"limit":     {strconv.Itoa(limit)},
		"direction": {"backward"},
	})
	if err != nil {
		return nil, err
	}
	var res struct {
		Data struct {
			Result []struct {
				Stream map[string]string `json:"stream"`
				Values [][2]string       `json:"values"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &res); err != nil {
		return nil, err
	}
	var rows []LogLine
	for _, st := range res.Data.Result {
		for _, v := range st.Values {
			ts, _ := strconv.ParseInt(v[0], 10, 64)
			rows = append(rows, LogLine{ts, st.Stream, v[1]})
		}
	}
	sortLogLines(rows)
	return rows, nil
}

// Notifications returns what the receiver got since the given Unix time.
func (l *Lab) Notifications(ctx context.Context, since float64) ([]alertmsg.Notification, error) {
	raw, err := l.ProxyGet(ctx, Receiver, "alerts", url.Values{"since": {strconv.FormatFloat(since, 'f', 3, 64)}})
	if err != nil {
		return nil, err
	}
	var out []alertmsg.Notification
	return out, json.Unmarshal(raw, &out)
}

// Sh runs a scenario shell command from the project root with the lab
// kubeconfig. It returns the exit code and combined output.
func (l *Lab) Sh(ctx context.Context, command string, timeout time.Duration) (int, string) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	cmd.Dir = l.Root
	cmd.Env = append(os.Environ(), "KUBECONFIG="+l.Kubeconfig)
	out, err := cmd.CombinedOutput()
	code := 0
	if err != nil {
		code = -1
		if ee, ok := err.(*exec.ExitError); ok {
			code = ee.ExitCode()
		}
	}
	return code, strings.TrimSpace(string(out))
}

// NowISO formats a time as 2006-01-02T15:04:05-0700.
func NowISO(t time.Time) string { return t.Format("2006-01-02T15:04:05-0700") }

// FiringLabAlerts lists the names of lab alerts (label lab="kursinis") that
// Prometheus currently reports as firing.
func (l *Lab) FiringLabAlerts(ctx context.Context) ([]string, error) {
	raw, err := l.ProxyGet(ctx, Prometheus, "api/v1/alerts", nil)
	if err != nil {
		return nil, err
	}
	var res struct {
		Data struct {
			Alerts []struct {
				Labels map[string]string `json:"labels"`
				State  string            `json:"state"`
			} `json:"alerts"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &res); err != nil {
		return nil, err
	}
	var firing []string
	for _, a := range res.Data.Alerts {
		if a.Labels["lab"] == "kursinis" && a.State == "firing" {
			firing = append(firing, a.Labels["alertname"])
		}
	}
	return firing, nil
}

// RuleNames lists the alert rule names of one Prometheus rule group.
func (l *Lab) RuleNames(ctx context.Context, group string) ([]string, error) {
	raw, err := l.ProxyGet(ctx, Prometheus, "api/v1/rules", nil)
	if err != nil {
		return nil, err
	}
	var res struct {
		Data struct {
			Groups []struct {
				Name  string `json:"name"`
				Rules []struct {
					Name string `json:"name"`
				} `json:"rules"`
			} `json:"groups"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &res); err != nil {
		return nil, err
	}
	var names []string
	for _, g := range res.Data.Groups {
		if g.Name == group {
			for _, r := range g.Rules {
				names = append(names, r.Name)
			}
		}
	}
	return names, nil
}

// AlertmanagerConfig returns the running Alertmanager configuration text.
func (l *Lab) AlertmanagerConfig(ctx context.Context) (string, error) {
	raw, err := l.ProxyGet(ctx, Alertmanager, "api/v2/status", nil)
	if err != nil {
		return "", err
	}
	var res struct {
		Config struct {
			Original string `json:"original"`
		} `json:"config"`
	}
	if err := json.Unmarshal(raw, &res); err != nil {
		return "", err
	}
	return res.Config.Original, nil
}
