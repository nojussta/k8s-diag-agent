package rules

import (
	"os"
	"strings"
	"testing"
)

// The fixtures are evidence.json files recorded by the fixed action sequence
// in live runs of S1-S3 (results/...).
func TestRulesRecogniseTheKnownFailureClasses(t *testing.T) {
	cases := []struct {
		file, rule, category, component, change string
	}{
		{"s1.json", "R1", "OOMKilled", "demo/deployment/backend", "--stress-memory=256"},
		{"s2.json", "R2", "ImagePullError", "demo/deployment/frontend", "6.15.99"},
		{"s3.json", "R3", "ConfigurationError", "demo/deployment/frontend", "frontend-config"},
	}
	for _, c := range cases {
		raw, err := os.ReadFile("testdata/" + c.file)
		if err != nil {
			t.Fatal(err)
		}
		d, rule, err := Diagnose(raw)
		if err != nil {
			t.Fatal(err)
		}
		if rule != c.rule || d.RootCause.Category != c.category || !strings.HasPrefix(d.RootCause.Component, c.component) {
			t.Errorf("%s: rule %s, %s, %s", c.file, rule, d.RootCause.Category, d.RootCause.Component)
		}
		if !strings.Contains(d.RootCause.Explanation, c.change) {
			t.Errorf("%s: explanation lacks the change %q: %s", c.file, c.change, d.RootCause.Explanation)
		}
		for _, s := range d.NextSteps {
			if s.Command != nil && !s.RequiresHumanApproval {
				t.Errorf("%s: command without approval", c.file)
			}
		}
	}
}

// A healthy pod with a recent change (as in S4/S5) matches no rule.
func TestRulesSayUnknownWhenNoKnownClassMatches(t *testing.T) {
	ev := `{"target":{"namespace":"demo","pod":"frontend-1"},"actions":[
	 {"id":"A2","ok":true,"output":{"phase":"Running","conditions":{"Ready":"True"},"containers":[{"name":"frontend","restartCount":0,"state":{"state":"running"}}]}},
	 {"id":"A3","ok":true,"output":["Normal Scheduled x1: assigned"]},
	 {"id":"A5","ok":true,"output":{"deployment":"frontend","revisions":[{"revision":3,"change_cause":"release 2.1"}],
	   "template_diff_latest_vs_previous":[{"container":"frontend","field":"args","previous":["--backend-url=http://backend:9898/echo"],"current":["--backend-url=http://backend:9899/echo"]}]}},
	 {"id":"A9","ok":true,"output":{"node":"n1","conditions":{"Ready":"True","MemoryPressure":"False"}}}]}`
	d, rule, err := Diagnose([]byte(ev))
	if err != nil {
		t.Fatal(err)
	}
	if rule != "none" || d.RootCause.Category != "Unknown" || d.Confidence != "low" {
		t.Fatalf("got rule %s, %+v", rule, d.RootCause)
	}
	if !strings.Contains(d.RootCause.Explanation, "not linked to a rule") {
		t.Errorf("the change should be reported but not claimed as the cause: %s", d.RootCause.Explanation)
	}
}
