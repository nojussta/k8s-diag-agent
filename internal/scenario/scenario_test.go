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

func TestJudgeComparesWithGroundTruth(t *testing.T) {
	gt := GroundTruth{Category: "OOMKilled", Match: Match{CauseAny: []string{"OOM", "memory limit"}, ComponentAny: []string{"backend"}}}
	mk := func(cat, comp, expl string) []byte {
		return []byte(`{"summary":"s","root_cause":{"category":"` + cat + `","component":"` + comp + `","explanation":"` + expl + `"}}`)
	}
	cases := []struct {
		name string
		in   []byte
		want bool
	}{
		{"right category and component", mk("OOMKilled", "demo/deployment/backend", "x"), true},
		{"cause only in text", mk("Unknown", "demo/backend", "exceeded the memory limit"), true},
		{"wrong component", mk("OOMKilled", "demo/deployment/frontend", "x"), false},
		{"wrong cause", mk("ImagePullError", "demo/deployment/backend", "bad tag"), false},
		{"unreadable", []byte("nope"), false},
	}
	for _, c := range cases {
		if got, why := Judge(gt, c.in); got != c.want {
			t.Errorf("%s: got %v (%s)", c.name, got, why)
		}
	}
}
