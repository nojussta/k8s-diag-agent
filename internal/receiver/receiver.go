// Package receiver implements the Alertmanager webhook handler and the bounded
// in-memory notification store behind the alert-receiver command.
package receiver

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/nojussta/k8s-diag-agent/internal/alertmsg"
)

// Store keeps the newest N notifications.
type Store struct {
	mu    sync.Mutex
	max   int
	items []alertmsg.Notification
}

func NewStore(max int) *Store { return &Store{max: max} }

func (s *Store) Add(n alertmsg.Notification) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = append(s.items, n)
	if len(s.items) > s.max {
		s.items = append([]alertmsg.Notification(nil), s.items[len(s.items)-s.max:]...)
	}
}

// Since returns notifications received at or after the given Unix time.
func (s *Store) Since(ts float64) []alertmsg.Notification {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]alertmsg.Notification, 0, len(s.items))
	for _, n := range s.items {
		if n.ReceivedAt >= ts {
			out = append(out, n)
		}
	}
	return out
}

// NewHandler wires the three endpoints. One JSON line per notification is
// written to logs (picked up by Alloy -> Loki as well).
func NewHandler(store *Store, logs io.Writer) http.Handler {
	mux := http.NewServeMux()
	logger := json.NewEncoder(logs)
	var logMu sync.Mutex

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /alerts", func(w http.ResponseWriter, r *http.Request) {
		since := 0.0
		if v := r.URL.Query().Get("since"); v != "" {
			f, err := strconv.ParseFloat(v, 64)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid since"})
				return
			}
			since = f
		}
		writeJSON(w, http.StatusOK, store.Since(since))
	})
	mux.HandleFunc("POST /alerts", func(w http.ResponseWriter, r *http.Request) {
		var payload alertmsg.Payload
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<20)).Decode(&payload); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
			return
		}
		store.Add(alertmsg.Notification{ReceivedAt: float64(time.Now().UnixNano()) / 1e9, Payload: payload})
		logMu.Lock()
		_ = logger.Encode(map[string]any{"event": "alert_received", "alerts": alertmsg.Summarize(payload)})
		logMu.Unlock()
		writeJSON(w, http.StatusOK, map[string]string{"status": "received"})
	})
	return mux
}

func writeJSON(w http.ResponseWriter, code int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(body)
}
