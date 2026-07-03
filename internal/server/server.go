// Package server implements the pomo daemon: an HTTP API over the SQLite store
// plus an in-process timer engine that fires notifications when sessions end.
//
// With the daemon alive, the running instance no longer needs systemd-run/at to
// fire the completion notification — a time.AfterFunc holds the timer. The OS
// scheduler remains useful only as a crash-survival fallback (not yet wired).
package server

import (
	"context"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"pomo.local/internal/notifier"
	"pomo.local/internal/store"
)

// Notifier is the desktop-notification dependency (satisfied by
// *notifier.LibnotifyNotifier), narrowed to what the server uses.
type Notifier interface {
	Notify(summary, body, hint string) error
}

// Server owns the store, the notifier, and the live per-session timers.
type Server struct {
	store    *store.Store
	notifier Notifier

	mu     sync.Mutex
	timers map[int64]*time.Timer // session id -> pending completion timer
}

// New constructs a Server. Passing a nil notifier falls back to libnotify.
func New(st *store.Store, n Notifier) *Server {
	if n == nil {
		n = &notifier.LibnotifyNotifier{}
	}
	return &Server{
		store:    st,
		notifier: n,
		timers:   make(map[int64]*time.Timer),
	}
}

// Handler returns the HTTP routes for the daemon.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.handleHealth)
	mux.HandleFunc("POST /api/sessions", s.handleStart)
	mux.HandleFunc("GET /api/sessions", s.handleList)
	mux.HandleFunc("GET /api/sessions/active", s.handleActive)
	mux.HandleFunc("POST /api/sessions/active/stop", s.handleStop)
	return mux
}

// Reconcile brings timer state in line with the database on startup. A session
// that should already have finished (daemon was down when it elapsed) is
// completed immediately; one still running gets a timer for its remaining time.
func (s *Server) Reconcile(ctx context.Context) error {
	sess, err := s.store.ActiveSession(ctx)
	if err == store.ErrNoActive {
		return nil
	}
	if err != nil {
		return err
	}

	if time.Now().After(sess.StopTime) {
		slog.Info("reconcile: active session already elapsed, completing", "id", sess.ID)
		s.complete(sess.ID, sess.Topic)
		return nil
	}

	remaining := time.Until(sess.StopTime)
	slog.Info("reconcile: rescheduling active session", "id", sess.ID, "remaining", remaining)
	s.arm(sess.ID, sess.Topic, remaining)
	return nil
}

// arm schedules a completion timer for a session after d.
func (s *Server) arm(id int64, topic string, d time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if t, ok := s.timers[id]; ok {
		t.Stop()
	}
	s.timers[id] = time.AfterFunc(d, func() { s.complete(id, topic) })
}

// disarm cancels a pending timer without firing it (used when a session is
// stopped early).
func (s *Server) disarm(id int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if t, ok := s.timers[id]; ok {
		t.Stop()
		delete(s.timers, id)
	}
}

// complete marks the session done and sends the completion notification. It is
// invoked from the timer goroutine, so it must not assume a request context.
func (s *Server) complete(id int64, topic string) {
	s.mu.Lock()
	delete(s.timers, id)
	s.mu.Unlock()

	ctx := context.Background()
	if err := s.store.FinishActiveSession(ctx, store.StatusDone); err != nil {
		slog.Error("complete: finish session", "id", id, "err", err)
		return
	}
	body := "Pomodoro session is ended!"
	if topic != "" {
		body = topic + " — session ended!"
	}
	if err := s.notifier.Notify("Pomodoro", body, ""); err != nil {
		slog.Error("complete: notify", "id", id, "err", err)
	}
}
