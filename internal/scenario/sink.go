package scenario

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Sink is a local stand-in for a chat webhook (Slack-style {"text": ...}).
// During a run every method gets its own URL, so the run records whether the
// report really reached the "channel"; in production the same agent flag
// points to a real incoming webhook.
type Sink struct {
	srv  *http.Server
	base string
	mu   sync.Mutex
	got  map[string][]Message
}

// Message is one received webhook post.
type Message struct {
	Received string `json:"received"`
	Text     string `json:"text"`
}

func StartSink() (*Sink, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	s := &Sink{base: "http://" + ln.Addr().String(), got: map[string][]Message{}}
	s.srv = &http.Server{Handler: http.HandlerFunc(s.handle), ReadHeaderTimeout: 5 * time.Second}
	go func() { _ = s.srv.Serve(ln) }()
	return s, nil
}

func (s *Sink) handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	var p struct {
		Text string `json:"text"`
	}
	b, _ := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if json.Unmarshal(b, &p) != nil || p.Text == "" {
		http.Error(w, `expected {"text": "..."}`, http.StatusBadRequest)
		return
	}
	name := strings.Trim(r.URL.Path, "/")
	s.mu.Lock()
	s.got[name] = append(s.got[name], Message{Received: time.Now().UTC().Format(time.RFC3339), Text: p.Text})
	s.mu.Unlock()
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

// URL is the webhook address for one method.
func (s *Sink) URL(method string) string { return s.base + "/" + method }

// Take returns and forgets the messages a method posted.
func (s *Sink) Take(method string) []Message {
	s.mu.Lock()
	defer s.mu.Unlock()
	m := s.got[method]
	delete(s.got, method)
	return m
}

func (s *Sink) Close() { _ = s.srv.Close() }
