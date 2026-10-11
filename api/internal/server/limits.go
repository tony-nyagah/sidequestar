package server

import (
	"context"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// window counts events per key over a sliding window. It lives in memory, which is
// fine for a single server: a restart forgets recent attempts, nothing more.
type window struct {
	mu     sync.Mutex
	max    int
	period time.Duration
	events map[string][]time.Time
}

func newWindow(max int, period time.Duration) *window {
	return &window{max: max, period: period, events: map[string][]time.Time{}}
}

func (w *window) prune(key string, now time.Time) []time.Time {
	kept := w.events[key][:0]
	for _, t := range w.events[key] {
		if now.Sub(t) < w.period {
			kept = append(kept, t)
		}
	}
	if len(kept) == 0 {
		delete(w.events, key)
		return nil
	}
	w.events[key] = kept
	return kept
}

// full reports whether key has used up its allowance.
func (w *window) full(key string) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return len(w.prune(key, time.Now())) >= w.max
}

func (w *window) add(key string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	now := time.Now()
	w.events[key] = append(w.prune(key, now), now)
}

// take records an event if there's room, and reports whether there was.
func (w *window) take(key string) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	now := time.Now()
	if len(w.prune(key, now)) >= w.max {
		return false
	}
	w.events[key] = append(w.events[key], now)
	return true
}

func (w *window) reset(key string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	delete(w.events, key)
}

type limits struct {
	loginByUser *window // failed sign-ins per username
	loginByIP   *window // failed sign-ins per address
	register    *window // new accounts per address
	quests      *window // generated quests per user
	model       chan struct{}
}

func newLimits() *limits {
	return &limits{
		loginByUser: newWindow(5, 15*time.Minute),
		loginByIP:   newWindow(20, 15*time.Minute),
		register:    newWindow(5, time.Hour),
		quests:      newWindow(30, time.Hour),
		// One model call at a time: on a CPU, two at once are each slower than one after the other.
		model: make(chan struct{}, 1),
	}
}

// useModel waits for the model to be free, or gives up when ctx ends.
func (l *limits) useModel(ctx context.Context) (release func(), err error) {
	select {
	case l.model <- struct{}{}:
		return func() { <-l.model }, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// clientIP is the visitor's address as passed on by the frontend. The API only
// listens on localhost by default, so the frontend is the only thing that can set it.
func clientIP(r *http.Request) string {
	if ip := strings.TrimSpace(r.Header.Get("X-Client-IP")); ip != "" {
		return ip
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func tooMany(w http.ResponseWriter, msg string) {
	w.Header().Set("Retry-After", "900")
	writeError(w, http.StatusTooManyRequests, msg)
}
