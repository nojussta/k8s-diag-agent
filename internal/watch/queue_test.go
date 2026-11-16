package watch

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/nojussta/k8s-diag-agent/internal/alertmsg"
)

func storm(n int) []alertmsg.Alert {
	var out []alertmsg.Alert
	for i := 0; i < n; i++ {
		out = append(out, alertmsg.Alert{Status: "firing", Fingerprint: fmt.Sprintf("fp%d", i),
			Labels: map[string]string{"alertname": "LabPodCrashLooping", "namespace": "demo", "pod": fmt.Sprintf("app-%d", i)}})
	}
	return out
}

// An alert storm: 20 alerts in one notification, budget 5 per 10 minutes.
func TestQueueBoundsAnAlertStorm(t *testing.T) {
	q := NewQueue(5, 10*time.Minute, 20)
	now := time.Now()
	queued, skipped := q.Offer(storm(20), now)
	if len(queued) != 5 || len(skipped) != 15 {
		t.Fatalf("queued %d skipped %d, want 5 and 15", len(queued), len(skipped))
	}
	if queued[0].Alert.Labels["pod"] != "app-0" || queued[4].Alert.Labels["pod"] != "app-4" {
		t.Fatal("alerts must be accepted in arrival order")
	}
	text := StormText(queued, skipped)
	if !strings.Contains(text, "15 alert(s) were NOT investigated") || !strings.Contains(text, "demo/app-19") {
		t.Fatalf("storm message: %s", text)
	}
	// The same alerts sent again (Alertmanager repeats) are duplicates: nothing new.
	if a, b := q.Offer(storm(20), now.Add(time.Minute)); len(a)+len(b) != 0 {
		t.Fatalf("repeat notification queued %d skipped %d, want 0 and 0", len(a), len(b))
	}
	if len(q.Items()) != 20 {
		t.Fatalf("items %d, want 20", len(q.Items()))
	}
}

func TestQueueBudgetRefillsAfterTheWindow(t *testing.T) {
	q := NewQueue(2, 10*time.Minute, 20)
	now := time.Now()
	if a, _ := q.Offer(storm(2), now); len(a) != 2 {
		t.Fatal("first two must be accepted")
	}
	other := []alertmsg.Alert{{Status: "firing", Fingerprint: "x", Labels: map[string]string{"alertname": "A"}}}
	if a, s := q.Offer(other, now.Add(5*time.Minute)); len(a) != 0 || len(s) != 1 {
		t.Fatal("inside the window the budget is used up")
	}
	later := []alertmsg.Alert{{Status: "firing", Fingerprint: "y", Labels: map[string]string{"alertname": "B"}}}
	if a, _ := q.Offer(later, now.Add(11*time.Minute)); len(a) != 1 {
		t.Fatal("after the window the budget is free again")
	}
}

func TestQueueIgnoresResolvedAndTestAlerts(t *testing.T) {
	q := NewQueue(0, 0, 0)
	alerts := []alertmsg.Alert{
		{Status: "resolved", Fingerprint: "a", Labels: map[string]string{"alertname": "A"}},
		{Status: "firing", Fingerprint: "b", Labels: map[string]string{"alertname": "LabEnvSelfTest", "severity": "none"}},
	}
	if a, s := q.Offer(alerts, time.Now()); len(a)+len(s) != 0 {
		t.Fatal("resolved and test alerts must be ignored")
	}
}

func TestQueueTreatsARefiredAlertAsANewIncident(t *testing.T) {
	q := NewQueue(0, 0, 0)
	now := time.Now()
	a := alertmsg.Alert{Status: "firing", Fingerprint: "fp", StartsAt: "2026-11-16T10:00:00Z", Labels: map[string]string{"alertname": "A"}}
	if x, _ := q.Offer([]alertmsg.Alert{a}, now); len(x) != 1 {
		t.Fatal("first firing must be queued")
	}
	if x, _ := q.Offer([]alertmsg.Alert{a}, now.Add(time.Minute)); len(x) != 0 {
		t.Fatal("a repeated notification of the same firing is a duplicate")
	}
	if x, _ := q.Offer([]alertmsg.Alert{a}, now.Add(3*time.Hour)); len(x) != 0 {
		t.Fatal("the same firing repeated hours later (repeat_interval) is still a duplicate")
	}
	a.StartsAt = "2026-11-16T10:20:00Z" // resolved in between, fired again
	if x, _ := q.Offer([]alertmsg.Alert{a}, now.Add(4*time.Hour)); len(x) != 1 {
		t.Fatal("a new firing of the same alert is a new incident")
	}
}
