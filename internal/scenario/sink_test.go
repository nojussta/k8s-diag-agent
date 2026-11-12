package scenario

import (
	"net/http"
	"strings"
	"testing"
)

func TestSinkRecordsMessagesPerMethod(t *testing.T) {
	s, err := StartSink()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	resp, err := http.Post(s.URL("agent"), "application/json", strings.NewReader(`{"text":"OOMKilled in backend"}`))
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("post: %v %v", resp, err)
	}
	bad, _ := http.Post(s.URL("agent"), "application/json", strings.NewReader(`{"blocks":[]}`))
	if bad.StatusCode != http.StatusBadRequest {
		t.Errorf("a payload without text must be rejected, got %d", bad.StatusCode)
	}
	if got := s.Take("agent"); len(got) != 1 || got[0].Text != "OOMKilled in backend" {
		t.Fatalf("got %+v", got)
	}
	if got := s.Take("agent"); len(got) != 0 {
		t.Fatalf("Take must forget the messages, got %+v", got)
	}
	if got := s.Take("pipeline"); len(got) != 0 {
		t.Fatalf("other methods have their own channel, got %+v", got)
	}
}
