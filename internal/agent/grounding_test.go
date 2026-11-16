package agent

import "testing"

func TestCheckGrounding(t *testing.T) {
	d := &Diagnosis{
		Summary:   "frontend calls the backend on port 9899",
		RootCause: RootCause{Explanation: "revision 7 changed --backend-url to port 9899; p95 is 2000 ms"},
		Evidence: []EvidenceItem{
			{ToolCall: "t1", Fact: "backend call failed, url http://backend:9899/echo"},
			{ToolCall: "t2", Fact: "exitCode=137, limit 128Mi"},
			{ToolCall: "t2", Fact: "the pod is Running"}, // no number: not checked
		},
	}
	results := map[string]string{
		"t1": `{"ok":true,"data":["backend call failed url http://backend:9899/echo"]}`,
		"t2": `{"ok":true,"data":{"exitCode":137,"limits":{"memory":"64Mi"}}}`,
	}
	g := CheckGrounding(d, results, `{"labels":{"alertname":"LabHighLatency"}}`)
	if g.Facts != 2 || g.FactsGrounded != 1 {
		t.Fatalf("facts %d grounded %d, want 2 and 1", g.Facts, g.FactsGrounded)
	}
	// evidence: 9899 ok, 137 ok, 128 missing; explanation: 9899 ok, 2000 missing
	if g.Numbers != 5 || g.NumbersGrounded != 3 {
		t.Fatalf("numbers %d grounded %d, want 5 and 3", g.Numbers, g.NumbersGrounded)
	}
	if len(g.Ungrounded) != 2 || g.Ungrounded[0] != "evidence[1] (t2): 128" || g.Ungrounded[1] != "explanation: 2000" {
		t.Fatalf("ungrounded = %v", g.Ungrounded)
	}
}

func TestCheckGroundingUsesTheCitedResultOnly(t *testing.T) {
	d := &Diagnosis{Evidence: []EvidenceItem{{ToolCall: "t1", Fact: "exit code 137"}}}
	g := CheckGrounding(d, map[string]string{"t1": "no numbers here", "t2": "exitCode 137"}, "")
	if g.FactsGrounded != 0 || len(g.Ungrounded) != 1 {
		t.Fatalf("a number found only in another result must not ground the fact: %+v", g)
	}
}

func TestPromptIDIsStable(t *testing.T) {
	if PromptID(SystemPrompt) == PromptID(PipelineSystemPrompt) || len(PromptID(SystemPrompt)) != 8 {
		t.Fatal("PromptID must be 8 hex digits and differ between prompts")
	}
}
