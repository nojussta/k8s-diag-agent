package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// Diagnosis is the agent's output (docs/schemas/diagnosis.schema.json). The
// rule-based baseline (internal/rules) produces the same structure, so every
// method is judged by the same code.
type Diagnosis struct {
	Summary    string         `json:"summary"`
	RootCause  RootCause      `json:"root_cause"`
	Confidence string         `json:"confidence"`
	Evidence   []EvidenceItem `json:"evidence"`
	Impact     string         `json:"impact"`
	NextSteps  []NextStep     `json:"next_steps"`
	Unknowns   []string       `json:"unknowns"`
}

// RootCause is the cause class, the affected component and the explanation
// (which should name the triggering change when there is one).
type RootCause struct {
	Category    string `json:"category"`
	Component   string `json:"component"`
	Explanation string `json:"explanation"`
}

// EvidenceItem cites the tool call (tN) or fixed-sequence action (AN) whose
// result shows the fact.
type EvidenceItem struct {
	ToolCall string `json:"tool_call"`
	Source   string `json:"source"`
	Fact     string `json:"fact"`
}

// NextStep is a suggestion; a command is never run by the agent.
type NextStep struct {
	Action                string  `json:"action"`
	Command               *string `json:"command"`
	RequiresHumanApproval bool    `json:"requires_human_approval"`
}

var categories = map[string]bool{
	"OOMKilled": true, "ApplicationCrash": true, "ImagePullError": true, "ConfigurationError": true,
	"ProbeFailure": true, "Scheduling": true, "NodeProblem": true, "DependencyFailure": true, "Unknown": true,
}

// ParseDiagnosis decodes and validates the model's final answer. executed is
// the set of tool-call ids that really ran; every evidence item must cite one
// of them and every suggested command needs human approval (security
// level 4: the agent only proposes, a person decides).
func ParseDiagnosis(raw []byte, executed map[string]bool) (*Diagnosis, error) {
	raw = bytes.TrimSpace(raw)
	raw = bytes.TrimPrefix(raw, []byte("```json"))
	raw = bytes.TrimPrefix(raw, []byte("```"))
	raw = bytes.TrimSuffix(raw, []byte("```"))
	dec := json.NewDecoder(bytes.NewReader(bytes.TrimSpace(raw)))
	dec.DisallowUnknownFields()
	var d Diagnosis
	if err := dec.Decode(&d); err != nil {
		return nil, fmt.Errorf("not valid JSON for the schema: %v", err)
	}
	var problems []string
	need := func(ok bool, msg string) {
		if !ok {
			problems = append(problems, msg)
		}
	}
	need(strings.TrimSpace(d.Summary) != "", "summary is empty")
	need(categories[d.RootCause.Category], fmt.Sprintf("root_cause.category %q is not in the enum", d.RootCause.Category))
	need(d.RootCause.Component != "" && d.RootCause.Explanation != "", "root_cause.component and explanation are required")
	need(d.Confidence == "high" || d.Confidence == "medium" || d.Confidence == "low", "confidence must be high, medium or low")
	need(len(d.Evidence) > 0, "evidence is empty")
	for i, e := range d.Evidence {
		need(executed[e.ToolCall], fmt.Sprintf("evidence[%d] cites %q, which is not a tool call that ran", i, e.ToolCall))
		need(strings.TrimSpace(e.Fact) != "", fmt.Sprintf("evidence[%d].fact is empty", i))
	}
	for i, s := range d.NextSteps {
		need(s.Command == nil || s.RequiresHumanApproval,
			fmt.Sprintf("next_steps[%d] has a command but requires_human_approval is false", i))
	}
	if len(problems) > 0 {
		return nil, fmt.Errorf("%s", strings.Join(problems, "; "))
	}
	return &d, nil
}
