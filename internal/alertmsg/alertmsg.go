// Package alertmsg describes the Alertmanager webhook payload (version "4") and
// the notification record the receiver stores. It is shared by the receiver and
// the lab tooling so both sides agree on one wire format.
package alertmsg

import (
	"fmt"
	"strings"
)

// Alert is one alert inside a webhook payload.
type Alert struct {
	Status       string            `json:"status"`
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations,omitempty"`
	StartsAt     string            `json:"startsAt,omitempty"`
	EndsAt       string            `json:"endsAt,omitempty"`
	GeneratorURL string            `json:"generatorURL,omitempty"`
	Fingerprint  string            `json:"fingerprint,omitempty"`
}

// Payload is the Alertmanager webhook body.
type Payload struct {
	Version           string            `json:"version,omitempty"`
	GroupKey          string            `json:"groupKey,omitempty"`
	TruncatedAlerts   int               `json:"truncatedAlerts,omitempty"`
	Status            string            `json:"status,omitempty"`
	Receiver          string            `json:"receiver,omitempty"`
	GroupLabels       map[string]string `json:"groupLabels,omitempty"`
	CommonLabels      map[string]string `json:"commonLabels,omitempty"`
	CommonAnnotations map[string]string `json:"commonAnnotations,omitempty"`
	ExternalURL       string            `json:"externalURL,omitempty"`
	Alerts            []Alert           `json:"alerts"`
}

// Notification is a payload together with the time the receiver got it
// (Unix seconds with a fractional part).
type Notification struct {
	ReceivedAt float64 `json:"received_at"`
	Payload    Payload `json:"payload"`
}

// Summarize returns one short line per alert, for logs and quick inspection.
func Summarize(p Payload) []string {
	lines := make([]string, 0, len(p.Alerts))
	for _, a := range p.Alerts {
		var target []string
		if ns := a.Labels["namespace"]; ns != "" {
			target = append(target, ns)
		}
		if w := firstNonEmpty(a.Labels["pod"], a.Labels["deployment"]); w != "" {
			target = append(target, w)
		}
		lines = append(lines, fmt.Sprintf("%s %s %s", orQ(a.Status), orQ(a.Labels["alertname"]), strings.Join(target, "/")))
	}
	return lines
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func orQ(s string) string {
	if s == "" {
		return "?"
	}
	return s
}
