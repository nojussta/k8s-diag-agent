// Package scenario loads fault scenarios and runs one end to end:
//
//  1. baseline check - the demo app is healthy and no lab alert is firing;
//  2. inject         - run the scenario's inject commands (t_inject);
//  3. detect         - wait until the expected alert reaches the receiver (t_alert);
//  4. diagnose       - collect evidence with the fixed action set (A1-A11) and,
//     when an agent command is given, run the agent (t_diagnosis);
//  5. reset          - run the scenario's reset commands and wait for health.
//
// Results go to results/<scenario>/<timestamp>/: run.json, alert.json,
// alert-notification.json, evidence.json, evidence.md and (with an agent
// command) agent-output.txt.
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

	"github.com/nojussta/k8s-diag-agent/internal/alertmsg"
	"github.com/nojussta/k8s-diag-agent/internal/evidence"
	"github.com/nojussta/k8s-diag-agent/internal/lab"
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

type Match struct {
	CauseAny     []string `json:"cause_any"`
	ComponentAny []string `json:"component_any"`
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

// Options control a run.
type Options struct {
	Keep     bool   // do not reset after the run
	AgentCmd string // command template; {alert} and {out} are substituted
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
	Agent              *AgentRun          `json:"agent,omitempty"`
	Finished           string             `json:"finished"`
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
}

// Judge compares the agent's diagnosis with the ground truth. The answer is
// correct when the category equals the expected one or the cause text contains
// an expected keyword, and the component names an expected component.
func Judge(gt GroundTruth, diagnosis []byte) (bool, string) {
	var d struct {
		Summary   string `json:"summary"`
		RootCause struct {
			Category    string `json:"category"`
			Component   string `json:"component"`
			Explanation string `json:"explanation"`
		} `json:"root_cause"`
	}
	if err := json.Unmarshal(diagnosis, &d); err != nil {
		return false, "diagnosis.json unreadable"
	}
	anyOf := func(text string, words []string) bool {
		t := strings.ToLower(text)
		for _, w := range words {
			if strings.Contains(t, strings.ToLower(w)) {
				return true
			}
		}
		return false
	}
	cause := d.RootCause.Category == gt.Category ||
		anyOf(d.RootCause.Category+" "+d.RootCause.Explanation+" "+d.Summary, gt.Match.CauseAny)
	comp := anyOf(d.RootCause.Component, gt.Match.ComponentAny)
	switch {
	case cause && comp:
		return true, "cause and component match"
	case !cause && !comp:
		return false, "neither cause nor component match"
	case !cause:
		return false, "cause does not match"
	}
	return false, "component does not match"
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

	if opt.AgentCmd != "" {
		cmd := strings.NewReplacer(
			"{alert}", shellQuote(filepath.Join(outDir, "alert.json")),
			"{out}", shellQuote(outDir),
		).Replace(opt.AgentCmd)
		t0 := time.Now()
		code, output := r.Lab.Sh(ctx, cmd, 10*time.Minute)
		rec.Agent = &AgentRun{ExitCode: code, DurationS: round1(time.Since(t0).Seconds())}
		_ = os.WriteFile(filepath.Join(outDir, "agent-output.txt"), []byte(output), 0o644)
		rec.Steps["agent"] = "failed"
		if code == 0 {
			rec.Steps["agent"] = "ok"
		}
		if b, err := os.ReadFile(filepath.Join(outDir, "diagnosis.json")); err == nil {
			rec.Agent.Diagnosed = true
			rec.Agent.Correct, rec.Agent.Verdict = Judge(sc.GroundTruth, b)
			r.logf("agent diagnosis: %s", rec.Agent.Verdict)
		}
	} else {
		rec.Steps["agent"] = "skipped (prototype phase)"
	}
	return rec, nil
}

func shellQuote(s string) string { return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'" }
