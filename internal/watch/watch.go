// Package watch is the queue and duplicate filter between the alert receiver
// and the agent (design component 2): it picks the firing lab alerts that have
// not been investigated recently, so each incident is diagnosed once and
// investigations run one at a time.
package watch

import (
	"sort"
	"strings"
	"time"

	"github.com/nojussta/k8s-diag-agent/internal/alertmsg"
)

// DefaultTTL is how long an investigated alert is not investigated again (FR2).
const DefaultTTL = 30 * time.Minute

// Dedup remembers investigated alerts.
type Dedup struct {
	TTL  time.Duration
	seen map[string]time.Time
}

// Key identifies an alert: Alertmanager's fingerprint, or its sorted labels.
func Key(a alertmsg.Alert) string {
	if a.Fingerprint != "" {
		return a.Fingerprint
	}
	keys := make([]string, 0, len(a.Labels))
	for k, v := range a.Labels {
		keys = append(keys, k+"="+v)
	}
	sort.Strings(keys)
	return strings.Join(keys, ",")
}

// Fresh reports whether the alert should be investigated now and, if so,
// remembers it.
func (d *Dedup) Fresh(a alertmsg.Alert, now time.Time) bool {
	if d.seen == nil {
		d.seen = map[string]time.Time{}
	}
	ttl := d.TTL
	if ttl == 0 {
		ttl = DefaultTTL
	}
	k := Key(a)
	if t, ok := d.seen[k]; ok && now.Sub(t) < ttl {
		return false
	}
	d.seen[k] = now
	return true
}

// Pick returns, in arrival order, the firing lab alerts of the notifications
// that are fresh. Resolved alerts and the environment self-test are skipped.
func Pick(ns []alertmsg.Notification, d *Dedup, now time.Time) []alertmsg.Alert {
	var out []alertmsg.Alert
	for _, n := range ns {
		for _, a := range n.Payload.Alerts {
			if a.Status != "firing" || a.Labels["lab"] != "kursinis" || a.Labels["alertname"] == "LabEnvSelfTest" {
				continue
			}
			if d.Fresh(a, now) {
				out = append(out, a)
			}
		}
	}
	return out
}
