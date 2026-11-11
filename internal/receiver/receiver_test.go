package receiver

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nojussta/k8s-diag-agent/internal/alertmsg"
)

const sample = `{"version":"4","status":"firing","alerts":[{"status":"firing","labels":{"alertname":"LabPodCrashLooping","namespace":"demo","pod":"backend-1"}}]}`

func TestPostThenGet(t *testing.T) {
	var logs bytes.Buffer
	srv := httptest.NewServer(NewHandler(NewStore(10), &logs))
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/alerts", "application/json", strings.NewReader(sample))
	if err != nil || resp.StatusCode != 200 {
		t.Fatalf("post: %v %v", err, resp)
	}
	resp, _ = http.Get(srv.URL + "/alerts")
	var got []alertmsg.Notification
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil || len(got) != 1 {
		t.Fatalf("get: %v %v", err, got)
	}
	if got[0].ReceivedAt <= 0 || got[0].Payload.Alerts[0].Labels["pod"] != "backend-1" {
		t.Fatalf("unexpected notification: %+v", got[0])
	}
	if !strings.Contains(logs.String(), "firing LabPodCrashLooping demo/backend-1") {
		t.Fatalf("missing log line: %q", logs.String())
	}
}

func TestSinceFilterAndBound(t *testing.T) {
	s := NewStore(2)
	for _, ts := range []float64{1, 2, 3} {
		s.Add(alertmsg.Notification{ReceivedAt: ts})
	}
	if got := s.Since(0); len(got) != 2 || got[0].ReceivedAt != 2 {
		t.Fatalf("store not bounded to newest 2: %+v", got)
	}
	if got := s.Since(3); len(got) != 1 {
		t.Fatalf("since filter: %+v", got)
	}
}

func TestBadRequests(t *testing.T) {
	srv := httptest.NewServer(NewHandler(NewStore(1), &bytes.Buffer{}))
	defer srv.Close()
	resp, _ := http.Post(srv.URL+"/alerts", "application/json", strings.NewReader("{nope"))
	if resp.StatusCode != 400 {
		t.Fatalf("invalid json: got %d", resp.StatusCode)
	}
	resp, _ = http.Get(srv.URL + "/alerts?since=abc")
	if resp.StatusCode != 400 {
		t.Fatalf("invalid since: got %d", resp.StatusCode)
	}
	resp, _ = http.Get(srv.URL + "/healthz")
	if resp.StatusCode != 200 {
		t.Fatalf("healthz: got %d", resp.StatusCode)
	}
	resp, _ = http.Get(srv.URL + "/nope")
	if resp.StatusCode != 404 {
		t.Fatalf("unknown path: got %d", resp.StatusCode)
	}
}
