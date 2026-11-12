// Package scenario loads fault scenarios and runs one end to end:
//
//  1. baseline check - the demo app is healthy and no lab alert is firing;
//  2. inject         - run the scenario's inject commands (t_inject);
//  3. detect         - wait until the expected alert reaches the receiver (t_alert);
//  4. diagnose       - collect evidence with the fixed action set (A1-A11),
//     apply the rule-based baseline to it, and run every configured method
//     (agent loop, pipeline, ...) against the same live fault;
//  5. reset          - run the scenario's reset commands and wait for health.
//
// Results go to results/<scenario>/<timestamp>/: run.json, alert.json,
// alert-notification.json, evidence.json, evidence.md, and one directory per
// method (rules/, agent/, pipeline/, ...) with diagnosis.json, metrics.json,
// report.md, notification.json and the method's output.
package scenario

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"github.com/nojussta/k8s-diag-agent/internal/agent"
	"github.com/nojussta/k8s-diag-agent/internal/alertmsg"
	"github.com/nojussta/k8s-diag-agent/internal/evidence"
	"github.com/nojussta/k8s-diag-agent/internal/lab"
	"github.com/nojussta/k8s-diag-agent/internal/notify"
	"github.com/nojussta/k8s-diag-agent/internal/rules"
)

// Scenario is the content of scenarios/<id>/scenario.yaml.
type Scenario struct {
	ID            string        `json:"id"`
	Title         string        `json:"title"`
	Status        string        `json:"status"` // implemented | planned
	Story         string        `json:"story"`
	Inject        []string      `json:"inject"`
	Reset         []string      `json:"reset"`
	ExpectedAlert ExpectedAlert `json:"expected_alert"`
	GroundTruth   GroundTruth   `json:"ground_truth"`
}

type ExpectedAlert struct {
	Alertname      string            `json:"alertname"`
	Labels         map[string]string `json:"labels"`
	TimeoutSeconds int               `json:"timeout_seconds"`
}

type GroundTruth struct {
	Category          string   `json:"category"`
	Component         string   `json:"component"`
	RootCause         string   `json:"root_cause"`
	KeyEvidence       []string `json:"key_evidence"`
	Match             Match    `json:"match"`
	AcceptableActions []string `json:"acceptable_actions"`
}

// Match holds what the automatic judge looks for (see judge.go).
type Match struct {
	// CategoryAny lists the accepted root_cause.category values (default: the
	// ground-truth category). Compared exactly.
	CategoryAny []string `json:"category_any"`
	// ComponentAny: root_cause.component must contain one of these (the
	// workload that has to be fixed, not the one that raised the alert).
	ComponentAny []string `json:"component_any"`
	// ComponentNot: the component must not contain any of these (e.g. the
	// healthy workload that only raised the alert).
	ComponentNot []string `json:"component_not"`
	// ChangeAny: root_cause.explanation or summary must contain one of these
	// (the change that triggered the fault, e.g. the changed argument).
	ChangeAny []string `json:"change_any"`
}

// Load reads one scenario from <root>/scenarios.
func Load(root, id string) (*Scenario, error) {
	path := filepath.Join(root, "scenarios", id, "scenario.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("unknown scenario %q; available: %v", id, List(root))
	}
	var sc Scenario
	if err := yaml.Unmarshal(data, &sc); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return &sc, nil
}

// List returns the available scenario ids.
func List(root string) []string {
	entries, _ := os.ReadDir(filepath.Join(root, "scenarios"))
	var ids []string
	for _, e := range entries {
		if e.IsDir() {
			ids = append(ids, e.Name())
		}
	}
	sort.Strings(ids)
	return ids
}

// Matches reports whether an alert is the firing alert the scenario expects.
func Matches(a alertmsg.Alert, exp ExpectedAlert) bool {
	if a.Status != "firing" || a.Labels["alertname"] != exp.Alertname {
		return false
	}
	for k, v := range exp.Labels {
		if a.Labels[k] != v {
			return false
		}
	}
	return true
}

// Method is one diagnosis method run after the fault is detected. Cmd is a
// shell command template: {alert}, {out} (the method's directory),
// {evidence} (the fixed sequence's evidence.json) and {notify} (a local chat
// webhook stand-in that records what the method posts) are substituted.
type Method struct {
	Name string
	Cmd  string
}

// Options control a run.
type Options struct {
	Keep    bool // do not reset after the run
	Methods []Method
}

// Record is run.json.
type Record struct {
	Scenario           string             `json:"scenario"`
	Title              string             `json:"title"`
	Started            string             `json:"started"`
	Steps              map[string]string  `json:"steps"`
	TInject            float64            `json:"t_inject,omitempty"`
	TAlert             float64            `json:"t_alert,omitempty"`
	DetectionLatencyS  float64            `json:"detection_latency_s,omitempty"`
	AlertLabels        map[string]string  `json:"alert_labels,omitempty"`
	BaselineCollection *BaselineCollected `json:"baseline_collection,omitempty"`
	// Methods holds every diagnosis method of the run, keyed by name
	// ("rules" always; "agent", "pipeline", ... when configured).
	Methods map[string]*MethodRun `json:"methods,omitempty"`
	// Agent is the single-agent record of runs made before methods existed.
	Agent    *AgentRun `json:"agent,omitempty"`
	Finished string    `json:"finished"`
}

type BaselineCollected struct {
	Calls     int     `json:"calls"`
	DurationS float64 `json:"duration_s"`
}

type AgentRun struct {
	ExitCode  int     `json:"exit_code"`
	DurationS float64 `json:"duration_s"`
	// Filled when the agent wrote diagnosis.json: automatic comparison with the
	// scenario's ground truth (see docs/scenarios.md).
	Diagnosed bool   `json:"diagnosed"`
	Correct   bool   `json:"correct"`
	Verdict   string `json:"verdict,omitempty"`
	// Judgement is the three-part verdict (filled by labctl judge for old runs).
	Judgement *Verdict `json:"judgement,omitempty"`
}

// MethodRun is the outcome of one method in one run.
type MethodRun struct {
	Cmd       string          `json:"cmd,omitempty"`
	ExitCode  int             `json:"exit_code"`
	DurationS float64         `json:"duration_s"` // wall clock of the command
	Diagnosed bool            `json:"diagnosed"`
	Judgement *Verdict        `json:"judgement,omitempty"`
	Rule      string          `json:"rule,omitempty"` // rules: the rule that fired
	Metrics   json.RawMessage `json:"metrics,omitempty"`
	Notified  bool            `json:"notified"`
}

// Runner runs scenarios against the lab.
type Runner struct {
	Lab *lab.Lab
	Log func(format string, args ...any)
}

func (r *Runner) logf(format string, args ...any) {
	if r.Log != nil {
		r.Log(format, args...)
	}
}

func unix(t time.Time) float64 { return float64(t.UnixNano()) / 1e9 }

func round1(v float64) float64 { return math.Round(v*10) / 10 }

func (r *Runner) runCommands(ctx context.Context, cmds []string, step string) error {
	for _, cmd := range cmds {
		code, out := r.Lab.Sh(ctx, cmd, 5*time.Minute)
		head := strings.Fields(cmd)
		if len(head) > 4 {
			head = head[:4]
		}
		r.logf("  [%s] %v -> exit %d", step, head, code)
		if code != 0 {
			return fmt.Errorf("%s command failed (%d): %s\n%s", step, code, cmd, out)
		}
	}
	return nil
}

// BaselineOK reports whether the demo app is healthy and no lab alert fires.
func (r *Runner) BaselineOK(ctx context.Context) (bool, string, error) {
	for _, name := range []string{"backend", "frontend", "loadgen"} {
		d, err := r.Lab.Cs.AppsV1().Deployments("demo").Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return false, "", err
		}
		want := int32(1)
		if d.Spec.Replicas != nil {
			want = *d.Spec.Replicas
		}
		if d.Status.AvailableReplicas != want || d.Status.UpdatedReplicas != want {
			return false, fmt.Sprintf("deployment %s not fully available", name), nil
		}
	}
	firing, err := r.Lab.FiringLabAlerts(ctx)
	if err != nil {
		return false, "", err
	}
	if len(firing) > 0 {
		return false, fmt.Sprintf("lab alerts still firing: %v", firing), nil
	}
	return true, "healthy", nil
}

func (r *Runner) waitForBaseline(ctx context.Context, timeout time.Duration) (bool, string) {
	deadline := time.Now().Add(timeout)
	why := "timeout"
	for time.Now().Before(deadline) {
		ok, w, err := r.BaselineOK(ctx)
		if err != nil {
			w = err.Error()
		}
		if ok {
			return true, w
		}
		why = w
		if !sleep(ctx, 10*time.Second) {
			break
		}
	}
	return false, why
}

func (r *Runner) waitForAlert(ctx context.Context, exp ExpectedAlert, since time.Time) (float64, *alertmsg.Alert, *alertmsg.Payload) {
	timeout := exp.TimeoutSeconds
	if timeout == 0 {
		timeout = 420
	}
	deadline := since.Add(time.Duration(timeout) * time.Second)
	for time.Now().Before(deadline) {
		ns, err := r.Lab.Notifications(ctx, unix(since))
		if err == nil {
			for _, n := range ns {
				for i := range n.Payload.Alerts {
					if Matches(n.Payload.Alerts[i], exp) {
						p := n.Payload
						return n.ReceivedAt, &p.Alerts[i], &p
					}
				}
			}
		}
		if !sleep(ctx, 3*time.Second) {
			break
		}
	}
	return 0, nil, nil
}

func sleep(ctx context.Context, d time.Duration) bool {
	select {
	case <-time.After(d):
		return true
	case <-ctx.Done():
		return false
	}
}

func writeJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// RunOnce runs one scenario once and writes the results into outDir.
func (r *Runner) RunOnce(ctx context.Context, sc *Scenario, opt Options, outDir string) (rec *Record, err error) {
	rec = &Record{Scenario: sc.ID, Title: sc.Title, Started: lab.NowISO(time.Now()), Steps: map[string]string{}}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return rec, err
	}

	ok, why, err := r.BaselineOK(ctx)
	if err != nil {
		return rec, err
	}
	rec.Steps["baseline"] = why
	if !ok {
		r.logf("baseline not healthy (%s); waiting up to 5 min", why)
		if ok, why = r.waitForBaseline(ctx, 5*time.Minute); !ok {
			return rec, fmt.Errorf("baseline not healthy: %s", why)
		}
	}

	// Reset and the record are written even when the run fails or is
	// interrupted, so the cluster is never left broken by a stopped run.
	defer func() {
		cleanup := context.WithoutCancel(ctx)
		if opt.Keep {
			r.logf("--keep: leaving the fault in place (reset later with the scenario's reset commands)")
			rec.Steps["reset"] = "skipped (--keep)"
		} else {
			r.logf("reset")
			if rerr := r.runCommands(cleanup, sc.Reset, "reset"); rerr != nil {
				rec.Steps["reset"] = "failed: " + rerr.Error()
				if err == nil {
					err = rerr
				}
			} else if ok, why := r.waitForBaseline(cleanup, 5*time.Minute); ok {
				rec.Steps["reset"] = "ok"
			} else {
				rec.Steps["reset"] = "not healthy: " + why
			}
		}
		rec.Finished = lab.NowISO(time.Now())
		if werr := writeJSON(filepath.Join(outDir, "run.json"), rec); werr != nil && err == nil {
			err = werr
		}
	}()

	tInject := time.Now()
	r.logf("inject: %s", sc.Title)
	if err = r.runCommands(ctx, sc.Inject, "inject"); err != nil {
		return rec, err
	}
	rec.TInject = unix(tInject)

	r.logf("waiting for alert %s ...", sc.ExpectedAlert.Alertname)
	tAlert, alert, payload := r.waitForAlert(ctx, sc.ExpectedAlert, tInject)
	if alert == nil {
		rec.Steps["detect"] = "timeout"
		return rec, fmt.Errorf("expected alert did not arrive in time")
	}
	rec.TAlert = tAlert
	rec.DetectionLatencyS = round1(tAlert - rec.TInject)
	rec.AlertLabels = alert.Labels
	rec.Steps["detect"] = "ok"
	if err = writeJSON(filepath.Join(outDir, "alert.json"), alert); err != nil {
		return rec, err
	}
	if err = writeJSON(filepath.Join(outDir, "alert-notification.json"), payload); err != nil {
		return rec, err
	}
	r.logf("alert received after %v s: %s", rec.DetectionLatencyS, alert.Labels["pod"])

	r.logf("collecting evidence with the fixed action set (A1-A11)")
	ev := evidence.Collect(ctx, r.Lab.Cs, r.Lab, *alert)
	if err = writeJSON(filepath.Join(outDir, "evidence.json"), ev); err != nil {
		return rec, err
	}
	if err = os.WriteFile(filepath.Join(outDir, "evidence.md"), []byte(ev.Markdown()), 0o644); err != nil {
		return rec, err
	}
	rec.BaselineCollection = &BaselineCollected{Calls: ev.TotalCalls, DurationS: ev.TotalDurationS}
	rec.Steps["evidence"] = "ok"
	r.logf("evidence: %d read-only calls, %v s", ev.TotalCalls, ev.TotalDurationS)

	rec.Methods = map[string]*MethodRun{}
	evRaw, _ := json.Marshal(ev)
	rec.Methods["rules"] = r.runRules(sc, evRaw, filepath.Join(outDir, "rules"), alert.Labels["alertname"])
	r.logf("rules: %s (%s)", rec.Methods["rules"].Judgement.Text, rec.Methods["rules"].Rule)

	if len(opt.Methods) > 0 {
		sink, err := StartSink()
		if err != nil {
			return rec, err
		}
		defer sink.Close()
		for _, m := range opt.Methods {
			rec.Methods[m.Name] = r.runMethod(ctx, sc, m, outDir, sink)
			rec.Steps[m.Name] = "ok"
			if rec.Methods[m.Name].ExitCode != 0 {
				rec.Steps[m.Name] = "failed"
			}
		}
	}
	return rec, nil
}

// RerunMethod runs a method again on a recorded run (its alert.json and
// evidence.json), for methods that need no live cluster such as the pipeline,
// and returns the new record. The previous method directory is replaced.
func RerunMethod(ctx context.Context, l *lab.Lab, sc *Scenario, m Method, runDir string) (*MethodRun, error) {
	if err := os.RemoveAll(filepath.Join(runDir, m.Name)); err != nil {
		return nil, err
	}
	sink, err := StartSink()
	if err != nil {
		return nil, err
	}
	defer sink.Close()
	return (&Runner{Lab: l}).runMethod(ctx, sc, m, runDir, sink), nil
}

// ApplyRules (re)applies the rule-based baseline to a recorded run directory
// (used when the rules change) and returns the method record.
func ApplyRules(sc *Scenario, runDir string) (*MethodRun, error) {
	evRaw, err := os.ReadFile(filepath.Join(runDir, "evidence.json"))
	if err != nil {
		return nil, err
	}
	alertname := ""
	if b, err := os.ReadFile(filepath.Join(runDir, "alert.json")); err == nil {
		var a alertmsg.Alert
		if json.Unmarshal(b, &a) == nil {
			alertname = a.Labels["alertname"]
		}
	}
	return (&Runner{}).runRules(sc, evRaw, filepath.Join(runDir, "rules"), alertname), nil
}

// runRules applies the rule-based baseline to the fixed sequence's evidence.
func (r *Runner) runRules(sc *Scenario, evRaw []byte, dir, alertname string) *MethodRun {
	mr := &MethodRun{}
	_ = os.MkdirAll(dir, 0o755)
	t0 := time.Now()
	d, rule, err := rules.Diagnose(evRaw)
	if err != nil {
		mr.ExitCode = 1
		return mr
	}
	var ev struct {
		TotalCalls     int     `json:"total_calls"`
		TotalDurationS float64 `json:"total_duration_s"`
	}
	_ = json.Unmarshal(evRaw, &ev)
	met := agent.Metrics{Mode: "rules", Model: "none", ToolCalls: ev.TotalCalls, StopReason: "diagnosis",
		DurationS: ev.TotalDurationS + time.Since(t0).Seconds()}
	mr.DurationS = round1(met.DurationS)
	mr.Rule = rule
	mr.Metrics, _ = json.Marshal(met)
	_ = writeJSON(filepath.Join(dir, "diagnosis.json"), d)
	_ = writeJSON(filepath.Join(dir, "metrics.json"), met)
	_ = os.WriteFile(filepath.Join(dir, "report.md"), []byte(notify.Text(alertname, d, met)), 0o644)
	b, _ := json.Marshal(d)
	v := Judge(sc.GroundTruth, b)
	mr.Diagnosed, mr.Judgement = true, &v
	return mr
}

// runMethod runs one external method command against the live fault.
func (r *Runner) runMethod(ctx context.Context, sc *Scenario, m Method, outDir string, sink *Sink) *MethodRun {
	dir := filepath.Join(outDir, m.Name)
	_ = os.MkdirAll(dir, 0o755)
	cmd := strings.NewReplacer(
		"{alert}", shellQuote(filepath.Join(outDir, "alert.json")),
		"{out}", shellQuote(dir),
		"{evidence}", shellQuote(filepath.Join(outDir, "evidence.json")),
		"{notify}", shellQuote(sink.URL(m.Name)),
	).Replace(m.Cmd)
	r.logf("method %s: running", m.Name)
	t0 := time.Now()
	code, output := r.Lab.Sh(ctx, cmd, 10*time.Minute)
	mr := &MethodRun{Cmd: m.Cmd, ExitCode: code, DurationS: round1(time.Since(t0).Seconds())}
	_ = os.WriteFile(filepath.Join(dir, "output.txt"), []byte(output), 0o644)
	if b, err := os.ReadFile(filepath.Join(dir, "metrics.json")); err == nil {
		mr.Metrics = b
	}
	if msgs := sink.Take(m.Name); len(msgs) > 0 {
		mr.Notified = true
		_ = writeJSON(filepath.Join(dir, "notification.json"), msgs)
	}
	if b, err := os.ReadFile(filepath.Join(dir, "diagnosis.json")); err == nil {
		v := Judge(sc.GroundTruth, b)
		mr.Diagnosed, mr.Judgement = true, &v
		r.logf("method %s: %s (exit %d, %.1f s)", m.Name, v.Text, code, mr.DurationS)
	} else {
		r.logf("method %s: no diagnosis (exit %d, %.1f s)", m.Name, code, mr.DurationS)
	}
	return mr
}

func shellQuote(s string) string { return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'" }
