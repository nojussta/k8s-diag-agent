package watch

import (
	"fmt"
	"sync"
	"time"

	"github.com/nojussta/k8s-diag-agent/internal/alertmsg"
)

// Item is one alert the in-cluster agent has accepted or refused.
type Item struct {
	ID       string         `json:"id"`
	Alert    alertmsg.Alert `json:"alert"`
	Received time.Time      `json:"received"`
	// Status: queued -> running -> done | failed; or skipped (never investigated).
	Status   string     `json:"status"`
	Reason   string     `json:"reason,omitempty"` // why it was skipped or failed
	Started  *time.Time `json:"started,omitempty"`
	Finished *time.Time `json:"finished,omitempty"`
	Notified bool       `json:"notified"`
	// Result is whatever the investigation reports (metrics and diagnosis).
	Result any `json:"result,omitempty"`
}

// Queue is the bounded queue in front of the orchestrator (design component
// 2). It answers what happens when many alerts arrive at once:
//
//   - the same alert is not investigated twice within the dedup TTL (FR2);
//   - investigations run one at a time, in arrival order;
//   - at most Budget investigations are accepted per Window. An alert storm
//     therefore costs a bounded number of model calls; the alerts over the
//     budget are not dropped silently but reported in one summary message.
//
// The zero value is not usable; call NewQueue.
type Queue struct {
	Budget int           // investigations accepted per Window
	Window time.Duration // sliding window of the budget
	MaxLen int           // pending investigations (queued, not yet started)

	mu       sync.Mutex
	dedup    Dedup
	accepted []time.Time // acceptance times inside the window
	items    []*Item     // every item, oldest first (bounded by keep)
	pending  chan *Item
	seq      int
}

// Defaults: 5 investigations per 10 minutes, 20 pending.
const (
	DefaultBudget = 5
	DefaultWindow = 10 * time.Minute
	DefaultMaxLen = 20
	keep          = 200
	// IncidentTTL is how long one firing of an alert is remembered by the queue.
	IncidentTTL = 7 * 24 * time.Hour
)

// NewQueue returns a queue; zero arguments select the defaults.
func NewQueue(budget int, window time.Duration, maxLen int) *Queue {
	if budget <= 0 {
		budget = DefaultBudget
	}
	if window <= 0 {
		window = DefaultWindow
	}
	if maxLen <= 0 {
		maxLen = DefaultMaxLen
	}
	// One firing of an alert (fingerprint + startsAt) is investigated once, however
	// long it keeps firing and however often Alertmanager repeats the notification.
	return &Queue{Budget: budget, Window: window, MaxLen: maxLen, pending: make(chan *Item, maxLen),
		dedup: Dedup{TTL: IncidentTTL}}
}

// Wanted reports whether an alert of a notification is one to investigate:
// firing, and not a test alert (severity "none").
func Wanted(a alertmsg.Alert) bool {
	return a.Status == "firing" && a.Labels["severity"] != "none"
}

// Offer hands the alerts of one notification to the queue and returns the
// items it queued and the items it refused because of the budget or a full
// queue. Resolved alerts, test alerts and recent duplicates are ignored.
func (q *Queue) Offer(alerts []alertmsg.Alert, now time.Time) (queued, skipped []*Item) {
	q.mu.Lock()
	defer q.mu.Unlock()
	live := q.accepted[:0]
	for _, t := range q.accepted {
		if now.Sub(t) < q.Window {
			live = append(live, t)
		}
	}
	q.accepted = live
	for _, a := range alerts {
		if !Wanted(a) || !q.dedup.Fresh(incident(a), now) {
			continue
		}
		q.seq++
		it := &Item{ID: fmt.Sprintf("%s-%03d", now.UTC().Format("20060102-150405"), q.seq), Alert: a, Received: now}
		switch {
		case len(q.accepted) >= q.Budget:
			it.Status, it.Reason = "skipped", fmt.Sprintf("budget of %d investigations per %s is used up", q.Budget, q.Window)
			skipped = append(skipped, it)
		case len(q.pending) >= q.MaxLen:
			it.Status, it.Reason = "skipped", fmt.Sprintf("queue is full (%d pending)", q.MaxLen)
			skipped = append(skipped, it)
		default:
			it.Status = "queued"
			q.accepted = append(q.accepted, now)
			q.pending <- it
			queued = append(queued, it)
		}
		q.items = append(q.items, it)
	}
	if len(q.items) > keep {
		q.items = append([]*Item(nil), q.items[len(q.items)-keep:]...)
	}
	return queued, skipped
}

// incident identifies one firing of an alert: Alertmanager repeats a
// notification with the same startsAt while the alert keeps firing (a
// duplicate), and gives a new startsAt when it fires again after being
// resolved (a new incident, which is investigated).
func incident(a alertmsg.Alert) alertmsg.Alert {
	return alertmsg.Alert{Fingerprint: Key(a) + "@" + a.StartsAt}
}

// Pending is the channel the single worker reads from.
func (q *Queue) Pending() <-chan *Item { return q.pending }

// Update changes an item under the queue's lock.
func (q *Queue) Update(it *Item, f func(*Item)) {
	q.mu.Lock()
	defer q.mu.Unlock()
	f(it)
}

// Items returns a copy of every remembered item, oldest first.
func (q *Queue) Items() []Item {
	q.mu.Lock()
	defer q.mu.Unlock()
	out := make([]Item, len(q.items))
	for i, it := range q.items {
		out[i] = *it
	}
	return out
}

// Target names what an alert is about, for messages and logs.
func Target(a alertmsg.Alert) string {
	l := a.Labels
	for _, k := range []string{"pod", "service", "deployment", "node"} {
		if l[k] != "" {
			return l["alertname"] + " " + l["namespace"] + "/" + l[k]
		}
	}
	return l["alertname"] + " " + l["namespace"]
}

// StormText is the single message sent for the alerts a notification brought
// over the budget: the engineer learns that they exist and were not analysed.
func StormText(queued, skipped []*Item) string {
	s := fmt.Sprintf("Alert storm: %d alert(s) were NOT investigated automatically (%s).", len(skipped), skipped[0].Reason)
	if len(queued) > 0 {
		s += fmt.Sprintf(" %d alert(s) of the same notification are being investigated.", len(queued))
	}
	s += "\nNot investigated:\n"
	for i, it := range skipped {
		if i == 15 {
			s += fmt.Sprintf("- ... and %d more\n", len(skipped)-i)
			break
		}
		s += "- " + Target(it.Alert) + "\n"
	}
	return s
}
