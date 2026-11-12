package notify

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nojussta/k8s-diag-agent/internal/agent"
)

func sample() *agent.Diagnosis {
	cmd := "kubectl -n demo rollout undo deployment/backend"
	return &agent.Diagnosis{
		Summary:    "s",
		RootCause:  agent.RootCause{Category: "OOMKilled", Component: "demo/deployment/backend", Explanation: "revision 2 added --stress-memory=256"},
		Confidence: "high",
		Evidence:   []agent.EvidenceItem{{ToolCall: "t1", Source: "get_pod_status", Fact: "OOMKilled"}},
		NextSteps:  []agent.NextStep{{Action: "Roll back", Command: &cmd, RequiresHumanApproval: true}},
	}
}

func TestTextHasCauseEvidenceAndSteps(t *testing.T) {
	txt := Text("LabPodCrashLooping", sample(), agent.Metrics{Mode: "agent", Model: "m", ToolCalls: 4, DurationS: 12.3})
	for _, want := range []string{"OOMKilled in demo/deployment/backend", "--stress-memory=256", "[t1 get_pod_status]", "not executed", "rollout undo", "4 read-only calls"} {
		if !strings.Contains(txt, want) {
			t.Errorf("message lacks %q:\n%s", want, txt)
		}
	}
}

func TestPostSendsSlackCompatiblePayload(t *testing.T) {
	var got map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &got)
	}))
	defer srv.Close()
	if err := Post(context.Background(), srv.URL, "hello"); err != nil {
		t.Fatal(err)
	}
	if got["text"] != "hello" {
		t.Fatalf("payload = %v", got)
	}
}

func TestPostReportsHTTPErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "no_service", http.StatusNotFound)
	}))
	defer srv.Close()
	if err := Post(context.Background(), srv.URL, "x"); err == nil || !strings.Contains(err.Error(), "404") {
		t.Fatalf("want a 404 error, got %v", err)
	}
}
