package scenario

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nojussta/k8s-diag-agent/internal/alertmsg"
	"github.com/nojussta/k8s-diag-agent/internal/lab"
	"github.com/nojussta/k8s-diag-agent/internal/watch"
)

// InClusterMethod is the method name of the agent that runs in the cluster.
const InClusterMethod = "agent-incluster"

// Investigation is one entry of the in-cluster agent's GET /investigations.
type Investigation struct {
	ID       string         `json:"id"`
	Alert    alertmsg.Alert `json:"alert"`
	Received time.Time      `json:"received"`
	Status   string         `json:"status"`
	Reason   string         `json:"reason"`
	Started  *time.Time     `json:"started"`
	Finished *time.Time     `json:"finished"`
	Notified bool           `json:"notified"`
}

// Investigations reads the in-cluster agent's queue.
func Investigations(ctx context.Context, l *lab.Lab) ([]Investigation, error) {
	raw, err := l.ProxyGet(ctx, lab.Receiver, "investigations", nil)
	if err != nil {
		return nil, err
	}
	var out []Investigation
	return out, json.Unmarshal(raw, &out)
}

// Messages reads what the in-cluster agent posted to its built-in chat stand-in.
func Messages(ctx context.Context, l *lab.Lab) ([]Message, error) {
	raw, err := l.ProxyGet(ctx, lab.Receiver, "messages", nil)
	if err != nil {
		return nil, err
	}
	var out []Message
	return out, json.Unmarshal(raw, &out)
}

// waitInCluster waits until the in-cluster agent has finished the alert that
// labctl only observed, then copies its files next to the other methods and
// judges the diagnosis. Nothing is started from the workstation.
func (r *Runner) waitInCluster(ctx context.Context, sc *Scenario, alert *alertmsg.Alert, since time.Time, outDir string) *MethodRun {
	dir := filepath.Join(outDir, InClusterMethod)
	_ = os.MkdirAll(dir, 0o755)
	mr := &MethodRun{Cmd: "in-cluster agent (Deployment diag-agent/diag-agent, started by the alert)", ExitCode: 1}
	note := func(format string, args ...any) {
		msg := fmt.Sprintf(format, args...)
		r.logf("method %s: %s", InClusterMethod, msg)
		_ = os.WriteFile(filepath.Join(dir, "output.txt"), []byte(msg+"\n"), 0o644)
	}

	var item *Investigation
	deadline := time.Now().Add(8 * time.Minute)
	for time.Now().Before(deadline) {
		list, err := Investigations(ctx, r.Lab)
		if err == nil {
			for i := range list {
				it := &list[i]
				if watch.Key(it.Alert) == watch.Key(*alert) && it.Received.After(since.Add(-5*time.Second)) {
					item = it
				}
			}
			if item != nil && (item.Status == "done" || item.Status == "failed" || item.Status == "skipped") {
				break
			}
		}
		if !sleep(ctx, 3*time.Second) {
			break
		}
	}
	switch {
	case item == nil:
		note("the in-cluster agent did not pick the alert up (is it deployed with AUTO_INVESTIGATE=true? run `make agent-on`)")
		return mr
	case item.Status == "skipped":
		note("the in-cluster agent skipped the alert: %s", item.Reason)
		return mr
	case item.Finished == nil:
		note("the in-cluster investigation %s did not finish in time (status %s)", item.ID, item.Status)
		return mr
	}
	for _, f := range []string{"trace.jsonl", "metrics.json", "diagnosis.json", "report.md"} {
		if b, err := r.Lab.ProxyGet(ctx, lab.Receiver, "investigations/"+item.ID+"/"+f, nil); err == nil {
			_ = os.WriteFile(filepath.Join(dir, f), b, 0o644)
		}
	}
	_ = writeJSON(filepath.Join(dir, "investigation.json"), item)
	if item.Started != nil {
		mr.DurationS = round1(item.Finished.Sub(*item.Started).Seconds())
	}
	if b, err := os.ReadFile(filepath.Join(dir, "metrics.json")); err == nil {
		mr.Metrics = b
	}
	// The message counts as delivered only if the stand-in channel really holds this report.
	if report, err := os.ReadFile(filepath.Join(dir, "report.md")); err == nil && item.Notified {
		if msgs, err := Messages(ctx, r.Lab); err == nil {
			for _, m := range msgs {
				if strings.TrimSpace(m.Text) == strings.TrimSpace(string(report)) {
					mr.Notified = true
					_ = writeJSON(filepath.Join(dir, "notification.json"), []Message{m})
					break
				}
			}
		}
	}
	if b, err := os.ReadFile(filepath.Join(dir, "diagnosis.json")); err == nil {
		v := Judge(sc.GroundTruth, b)
		mr.Diagnosed, mr.Judgement, mr.ExitCode = true, &v, 0
		note("investigation %s: %s (%.1f s, started by the alert %.1f s after it arrived, message delivered: %v)",
			item.ID, v.Text, mr.DurationS, item.Started.Sub(item.Received).Seconds(), mr.Notified)
	} else {
		note("investigation %s ended without a diagnosis: %s", item.ID, item.Reason)
	}
	return mr
}
