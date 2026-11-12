package scenario

import (
	"path/filepath"
	"testing"

	"github.com/nojussta/k8s-diag-agent/internal/alertmsg"
)

func TestMatches(t *testing.T) {
	exp := ExpectedAlert{Alertname: "LabPodCrashLooping", Labels: map[string]string{"namespace": "demo", "container": "backend"}}
	good := alertmsg.Alert{Status: "firing", Labels: map[string]string{"alertname": "LabPodCrashLooping", "namespace": "demo", "container": "backend", "pod": "x"}}
	if !Matches(good, exp) {
		t.Fatal("should match")
	}
	resolved := good
	resolved.Status = "resolved"
	other := alertmsg.Alert{Status: "firing", Labels: map[string]string{"alertname": "LabPodNotReady", "namespace": "demo", "container": "backend"}}
	wrongLabel := alertmsg.Alert{Status: "firing", Labels: map[string]string{"alertname": "LabPodCrashLooping", "namespace": "demo", "container": "frontend"}}
	for name, a := range map[string]alertmsg.Alert{"resolved": resolved, "other alertname": other, "label mismatch": wrongLabel} {
		if Matches(a, exp) {
			t.Errorf("%s must not match", name)
		}
	}
}

// The shipped scenario files must parse into the typed model.
func TestShippedScenariosLoad(t *testing.T) {
	root, _ := filepath.Abs("../..")
	ids := List(root)
	if len(ids) < 3 {
		t.Fatalf("expected at least 3 scenarios, got %v", ids)
	}
	for _, id := range ids {
		sc, err := Load(root, id)
		if err != nil {
			t.Fatalf("%s: %v", id, err)
		}
		if sc.ID != id || len(sc.Inject) == 0 || len(sc.Reset) == 0 || sc.ExpectedAlert.Alertname == "" || sc.GroundTruth.RootCause == "" {
			t.Errorf("%s incomplete: %+v", id, sc)
		}
	}
	sc, _ := Load(root, "s1-oomkilled")
	if sc.Status != "implemented" || sc.ExpectedAlert.TimeoutSeconds != 420 || sc.GroundTruth.Category != "OOMKilled" {
		t.Errorf("s1 parsed wrongly: %+v", sc)
	}
}

func TestLoadUnknown(t *testing.T) {
	if _, err := Load(t.TempDir(), "nope"); err == nil {
		t.Fatal("want error")
	}
}

func TestShellQuote(t *testing.T) {
	if got := shellQuote("a'b c"); got != `'a'\''b c'` {
		t.Fatal(got)
	}
}

func TestJudgeNeedsCategoryComponentAndChange(t *testing.T) {
	gt := GroundTruth{Category: "OOMKilled", Match: Match{
		ComponentAny: []string{"backend"},
		ChangeAny:    []string{"stress-memory", "release 2.0"},
	}}
	diag := func(cat, comp, expl string) []byte {
		return []byte(`{"summary":"s","root_cause":{"category":"` + cat + `","component":"` + comp + `","explanation":"` + expl + `"}}`)
	}
	cases := []struct {
		name              string
		in                []byte
		cat, comp, change bool
		correct           bool
	}{
		{"all three", diag("OOMKilled", "demo/deployment/backend", "revision 3 added --stress-memory=256"), true, true, true, true},
		// The old, lenient judge accepted this: right class, but no root cause.
		{"symptom only", diag("OOMKilled", "demo/backend-abc", "insufficient memory"), true, true, false, false},
		// A keyword in the text no longer rescues a wrong category.
		{"keyword is not a category", diag("NodeProblem", "demo/deployment/backend", "OOMKilled after release 2.0"), false, true, true, false},
		{"wrong component", diag("OOMKilled", "demo/deployment/frontend", "--stress-memory=256"), true, false, true, false},
		{"unreadable", []byte("not json"), false, false, false, false},
	}
	for _, c := range cases {
		v := Judge(gt, c.in)
		if v.Category != c.cat || v.Component != c.comp || v.Change != c.change || v.Correct != c.correct {
			t.Errorf("%s: got %+v", c.name, v)
		}
	}
	if s := Judge(gt, diag("OOMKilled", "backend", "x")).Score(); s < 0.66 || s > 0.67 {
		t.Errorf("two of three parts should score 2/3, got %v", s)
	}
}

func TestJudgeAcceptsAlternativeCategories(t *testing.T) {
	gt := GroundTruth{Category: "ConfigurationError", Match: Match{
		CategoryAny: []string{"ConfigurationError", "DependencyFailure"}, ComponentAny: []string{"frontend"}, ChangeAny: []string{"9899"}}}
	d := []byte(`{"summary":"","root_cause":{"category":"DependencyFailure","component":"demo/deployment/frontend","explanation":"backend-url port 9899"}}`)
	if v := Judge(gt, d); !v.Correct {
		t.Fatalf("got %+v", v)
	}
}

func TestShippedScenariosDeclareWhatTheJudgeNeeds(t *testing.T) {
	for _, id := range List("../..") {
		sc, err := Load("../..", id)
		if err != nil {
			t.Fatal(err)
		}
		m := sc.GroundTruth.Match
		if len(m.ComponentAny) == 0 || len(m.ChangeAny) == 0 {
			t.Errorf("%s: match needs component_any and change_any", id)
		}
	}
}
