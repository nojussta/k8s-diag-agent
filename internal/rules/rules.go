// Package rules is the rule-based baseline: a fixed set of if-then rules over
// the evidence that the fixed action sequence (A1-A11, internal/evidence)
// collected. It plays the role of analyzers such as K8sGPT's: every rule
// recognises one known failure class from the pod status, the events and the
// node conditions. The newest Deployment change (A5) is attached to the
// explanation, so the baseline names the triggering change whenever a rule fires.
//
// It uses no language model, so it is deterministic, free and instant. It is
// compared with the agent to show what the language model adds: the rules
// cover only the classes they were written for and answer "Unknown" otherwise.
package rules

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/nojussta/k8s-diag-agent/internal/agent"
)

// Evidence is the part of evidence.json the rules read.
type Evidence struct {
	Target struct {
		Namespace  string `json:"namespace"`
		Pod        string `json:"pod"`
		Container  string `json:"container"`
		Deployment string `json:"deployment"`
		Service    string `json:"service"`
	} `json:"target"`
	Actions []struct {
		ID     string          `json:"id"`
		Name   string          `json:"name"`
		OK     bool            `json:"ok"`
		Output json.RawMessage `json:"output"`
	} `json:"actions"`
}

type containerState struct {
	State    string `json:"state"`
	Reason   string `json:"reason"`
	Message  string `json:"message"`
	ExitCode *int   `json:"exitCode"`
}

type podView struct {
	Phase      string            `json:"phase"`
	Conditions map[string]string `json:"conditions"`
	Containers []struct {
		Name         string          `json:"name"`
		Image        string          `json:"image"`
		RestartCount int             `json:"restartCount"`
		State        *containerState `json:"state"`
		LastState    *containerState `json:"lastState"`
	} `json:"containers"`
}

type rolloutView struct {
	Deployment string `json:"deployment"`
	Revisions  []struct {
		Revision    int     `json:"revision"`
		ChangeCause *string `json:"change_cause"`
	} `json:"revisions"`
	Diff []struct {
		Container string `json:"container"`
		Field     string `json:"field"`
		Previous  any    `json:"previous"`
		Current   any    `json:"current"`
	} `json:"template_diff_latest_vs_previous"`
}

type nodeView struct {
	Node       string            `json:"node"`
	Conditions map[string]string `json:"conditions"`
}

// Rule is one if-then rule: Match returns a cause sentence when it fires.
type Rule struct {
	ID       string
	Category string
	Match    func(f facts) (string, bool)
}

// facts are the parsed actions.
type facts struct {
	pod    podView
	events []string
	roll   rolloutView
	node   nodeView
}

var missingObject = regexp.MustCompile(`(configmap|secret) "[^"]+" not found`)

// Rules in priority order; the first that fires wins.
var Rules = []Rule{
	{"R1", "OOMKilled", func(f facts) (string, bool) {
		for _, c := range f.pod.Containers {
			for _, s := range []*containerState{c.LastState, c.State} {
				if s != nil && s.Reason == "OOMKilled" {
					return fmt.Sprintf("container %s was killed for exceeding its memory limit (OOMKilled%s)", c.Name, exit(s)), true
				}
			}
		}
		return "", false
	}},
	{"R2", "ImagePullError", func(f facts) (string, bool) {
		for _, c := range f.pod.Containers {
			if s := c.State; s != nil && (s.Reason == "ErrImagePull" || s.Reason == "ImagePullBackOff" || s.Reason == "InvalidImageName") {
				return fmt.Sprintf("image %s of container %s cannot be pulled (%s)", c.Image, c.Name, s.Reason), true
			}
		}
		return "", false
	}},
	{"R3", "ConfigurationError", func(f facts) (string, bool) {
		for _, c := range f.pod.Containers {
			if s := c.State; s != nil && (s.Reason == "CreateContainerConfigError" || s.Reason == "CreateContainerError") {
				return fmt.Sprintf("container %s cannot be created (%s: %s)", c.Name, s.Reason, s.Message), true
			}
		}
		for _, e := range f.events {
			if m := missingObject.FindString(e); m != "" {
				return "a referenced object is missing: " + m, true
			}
		}
		return "", false
	}},
	{"R4", "Scheduling", func(f facts) (string, bool) {
		for _, e := range f.events {
			if strings.Contains(e, "FailedScheduling") {
				return "the pod cannot be scheduled: " + after(e, ": "), true
			}
		}
		return "", false
	}},
	{"R5", "NodeProblem", func(f facts) (string, bool) {
		for k, v := range f.node.Conditions {
			if (k == "Ready" && v != "True") || (strings.HasSuffix(k, "Pressure") && v == "True") {
				return fmt.Sprintf("node %s reports %s=%s", f.node.Node, k, v), true
			}
		}
		return "", false
	}},
	{"R6", "ApplicationCrash", func(f facts) (string, bool) {
		for _, c := range f.pod.Containers {
			if s := c.LastState; s != nil && s.State == "terminated" && s.ExitCode != nil && *s.ExitCode != 0 && c.RestartCount > 0 {
				return fmt.Sprintf("container %s keeps exiting with an error (%s%s)", c.Name, s.Reason, exit(s)), true
			}
		}
		return "", false
	}},
	{"R7", "ProbeFailure", func(f facts) (string, bool) {
		// A probe failing once while a new pod starts is normal; the rule
		// fires only when the pod is not ready now.
		if f.pod.Conditions["Ready"] == "True" {
			return "", false
		}
		for _, e := range f.events {
			if strings.Contains(e, "Unhealthy") && strings.Contains(e, "probe failed") {
				return "health checks fail: " + after(e, ": "), true
			}
		}
		return "", false
	}},
}

// Diagnose applies the rules to an evidence.json document.
func Diagnose(evidenceJSON []byte) (*agent.Diagnosis, string, error) {
	var ev Evidence
	if err := json.Unmarshal(evidenceJSON, &ev); err != nil {
		return nil, "", fmt.Errorf("evidence.json: %w", err)
	}
	var f facts
	out := map[string]json.RawMessage{}
	for _, a := range ev.Actions {
		if a.OK {
			out[a.ID] = a.Output
		}
	}
	_ = json.Unmarshal(out["A2"], &f.pod)
	_ = json.Unmarshal(out["A3"], &f.events)
	_ = json.Unmarshal(out["A5"], &f.roll)
	_ = json.Unmarshal(out["A9"], &f.node)

	t := ev.Target
	component := fmt.Sprintf("%s/pod/%s", t.Namespace, t.Pod)
	if f.roll.Deployment != "" {
		component = fmt.Sprintf("%s/deployment/%s", t.Namespace, f.roll.Deployment)
	}
	if t.Container != "" {
		component += fmt.Sprintf(" (container %s)", t.Container)
	}
	change := latestChange(f.roll)

	d := &agent.Diagnosis{RootCause: agent.RootCause{Component: component}, Unknowns: []string{}}
	fired := ""
	for _, r := range Rules {
		if cause, ok := r.Match(f); ok {
			fired = r.ID
			d.RootCause.Category = r.Category
			d.RootCause.Explanation = cause + "."
			d.Confidence = "high"
			break
		}
	}
	if fired == "" {
		d.RootCause.Category = "Unknown"
		d.RootCause.Explanation = "No rule matched the pod status, events or node conditions."
		d.Confidence = "low"
		d.Unknowns = append(d.Unknowns, "the fault is not one of the known failure classes R1-R7")
	}
	if change != "" {
		if fired != "" {
			d.RootCause.Explanation += " It started with the newest revision: " + change
		} else {
			d.RootCause.Explanation += " The newest revision (not linked to a rule): " + change
		}
	}
	d.Summary = fmt.Sprintf("%s in %s: %s", d.RootCause.Category, component, d.RootCause.Explanation)

	if len(out["A2"]) > 0 {
		d.Evidence = append(d.Evidence, agent.EvidenceItem{ToolCall: "A2", Source: "pod status", Fact: podFact(f.pod)})
	}
	if n := len(f.events); n > 0 {
		d.Evidence = append(d.Evidence, agent.EvidenceItem{ToolCall: "A3", Source: "events", Fact: f.events[n-1]})
	}
	if change != "" {
		d.Evidence = append(d.Evidence, agent.EvidenceItem{ToolCall: "A5", Source: "rollout history", Fact: change})
	}
	if len(d.Evidence) == 0 {
		d.Evidence = append(d.Evidence, agent.EvidenceItem{ToolCall: "A1", Source: "alert", Fact: "target " + t.Namespace + "/" + t.Pod})
	}
	if f.roll.Deployment != "" && fired != "" && change != "" {
		cmd := fmt.Sprintf("kubectl -n %s rollout undo deployment/%s", t.Namespace, f.roll.Deployment)
		d.NextSteps = append(d.NextSteps, agent.NextStep{Action: "Roll back to the previous revision", Command: &cmd, RequiresHumanApproval: true})
	} else {
		d.NextSteps = append(d.NextSteps, agent.NextStep{Action: "Investigate manually: no rule covers this case", RequiresHumanApproval: true})
	}
	d.Impact = impact(f)
	if fired == "" {
		fired = "none"
	}
	return d, fired, nil
}

// latestChange describes the newest revision's template change, or "".
func latestChange(r rolloutView) string {
	if len(r.Diff) == 0 {
		return ""
	}
	var parts []string
	for _, x := range r.Diff {
		parts = append(parts, fmt.Sprintf("%s %s: %s -> %s", x.Container, x.Field, compact(x.Previous), compact(x.Current)))
	}
	s := ""
	if len(r.Revisions) > 0 {
		s = fmt.Sprintf("revision %d", r.Revisions[0].Revision)
		if c := r.Revisions[0].ChangeCause; c != nil && *c != "" {
			s += fmt.Sprintf(" (change-cause: %s)", *c)
		}
		s += " changed "
	}
	return s + strings.Join(parts, "; ") + "."
}

func podFact(p podView) string {
	var parts []string
	for _, c := range p.Containers {
		s := fmt.Sprintf("%s restarts=%d", c.Name, c.RestartCount)
		if c.State != nil && c.State.Reason != "" {
			s += " state=" + c.State.Reason
		}
		if c.LastState != nil && c.LastState.Reason != "" {
			s += " lastState=" + c.LastState.Reason + exit(c.LastState)
		}
		parts = append(parts, s)
	}
	return "phase=" + p.Phase + "; " + strings.Join(parts, "; ")
}

func impact(f facts) string {
	if f.pod.Phase == "Running" && len(f.pod.Containers) > 0 && f.pod.Conditions["Ready"] == "True" {
		return "The inspected pod is running and ready."
	}
	return "The inspected pod is not ready (phase " + f.pod.Phase + ")."
}

func exit(s *containerState) string {
	if s.ExitCode == nil {
		return ""
	}
	return fmt.Sprintf(", exit code %d", *s.ExitCode)
}

func after(s, sep string) string {
	if i := strings.Index(s, sep); i >= 0 {
		return s[i+len(sep):]
	}
	return s
}

func compact(v any) string {
	b, _ := json.Marshal(v)
	s := string(b)
	if len(s) > 300 {
		s = s[:300] + "..."
	}
	return s
}
