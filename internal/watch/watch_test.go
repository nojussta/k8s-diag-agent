package watch

import (
	"testing"
	"time"

	"github.com/nojussta/k8s-diag-agent/internal/alertmsg"
)

func alert(status, name, pod string) alertmsg.Alert {
	return alertmsg.Alert{Status: status, Labels: map[string]string{"alertname": name, "pod": pod, "lab": "kursinis"}}
}

func TestPickInvestigatesEachFiringAlertOnce(t *testing.T) {
	d := &Dedup{}
	now := time.Now()
	n := func(as ...alertmsg.Alert) []alertmsg.Notification {
		return []alertmsg.Notification{{Payload: alertmsg.Payload{Alerts: as}}}
	}
	got := Pick(n(alert("firing", "LabPodCrashLooping", "a"), alert("resolved", "LabPodNotReady", "b"),
		alert("firing", "LabEnvSelfTest", "c")), d, now)
	if len(got) != 1 || got[0].Labels["pod"] != "a" {
		t.Fatalf("want only the firing lab alert, got %+v", got)
	}
	// Alertmanager repeats the notification: no second investigation.
	if got := Pick(n(alert("firing", "LabPodCrashLooping", "a")), d, now.Add(time.Minute)); len(got) != 0 {
		t.Fatalf("duplicate investigated: %+v", got)
	}
	// After the TTL the same alert is investigated again.
	if got := Pick(n(alert("firing", "LabPodCrashLooping", "a")), d, now.Add(DefaultTTL+time.Second)); len(got) != 1 {
		t.Fatalf("want a new investigation after the TTL, got %+v", got)
	}
}

func TestAlertsOutsideTheLabAreIgnored(t *testing.T) {
	a := alert("firing", "KubeCPUOvercommit", "x")
	delete(a.Labels, "lab")
	if got := Pick([]alertmsg.Notification{{Payload: alertmsg.Payload{Alerts: []alertmsg.Alert{a}}}}, &Dedup{}, time.Now()); len(got) != 0 {
		t.Fatalf("got %+v", got)
	}
}
