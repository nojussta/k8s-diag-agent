// Package notify turns a diagnosis into a short message for the on-call
// engineer and delivers it to a chat webhook. The payload is the de facto
// incoming-webhook format {"text": "..."} accepted by Slack, Mattermost and
// Rocket.Chat; any other service can be bridged by a small relay. Delivery is
// best effort: a failed post never hides the diagnosis, which is always
// written to report.md as well.
package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/nojussta/k8s-diag-agent/internal/agent"
)

// Text renders the message: the conclusion first, then the evidence and the
// suggested steps, then how the result was obtained.
func Text(alertname string, d *agent.Diagnosis, m agent.Metrics) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Incident triage: %s\n", alertname)
	fmt.Fprintf(&b, "Probable cause: %s in %s (confidence %s)\n", d.RootCause.Category, d.RootCause.Component, d.Confidence)
	fmt.Fprintf(&b, "%s\n", d.RootCause.Explanation)
	if len(d.Evidence) > 0 {
		b.WriteString("\nEvidence:\n")
		for _, e := range d.Evidence {
			fmt.Fprintf(&b, "- [%s %s] %s\n", e.ToolCall, e.Source, e.Fact)
		}
	}
	if len(d.NextSteps) > 0 {
		b.WriteString("\nSuggested next steps (not executed; a person decides):\n")
		for _, s := range d.NextSteps {
			fmt.Fprintf(&b, "- %s", s.Action)
			if s.Command != nil && *s.Command != "" {
				fmt.Fprintf(&b, ": %s", *s.Command)
			}
			b.WriteString("\n")
		}
	}
	if len(d.Unknowns) > 0 {
		fmt.Fprintf(&b, "\nUnknown: %s\n", strings.Join(d.Unknowns, "; "))
	}
	fmt.Fprintf(&b, "\nMethod %s, model %s, %d read-only calls, %.1f s.\n", m.Mode, m.Model, m.ToolCalls, m.DurationS)
	return b.String()
}

// Post sends the text to an incoming webhook.
func Post(ctx context.Context, url, text string) error {
	body, _ := json.Marshal(map[string]string{"text": text})
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 300))
		return fmt.Errorf("webhook answered %s: %s", resp.Status, strings.TrimSpace(string(msg)))
	}
	return nil
}
