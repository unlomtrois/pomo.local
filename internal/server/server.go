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

	"pomo.local/internal/mail"
	"pomo.local/internal/notifier"
	"pomo.local/internal/store"
	"pomo.local/internal/webui"
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
	mux.HandleFunc("GET /api/sessions/by-hash/{prefix}", s.handleByHash)
	mux.HandleFunc("DELETE /api/sessions/by-hash/{prefix}", s.handleDelete)
	mux.HandleFunc("PATCH /api/sessions/by-hash/{prefix}", s.handleEdit)
	mux.HandleFunc("PATCH /api/sessions/{id}", s.handleMove)
	mux.HandleFunc("GET /api/sessions/active", s.handleActive)
	mux.HandleFunc("POST /api/sessions/active/stop", s.handleStop)
	mux.HandleFunc("POST /api/sessions/active/end", s.handleEnd)

	// Everything else serves the embedded SPA (dashboard + calendar). The API
	// routes above are more specific, so they take precedence over "/".
	mux.Handle("/", webui.Handler())
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

	if sess.Open {
		slog.Info("reconcile: active session is an open stopwatch, no timer", "id", sess.ID)
		return nil
	}

	if time.Now().After(sess.StopTime) {
		slog.Info("reconcile: active session already elapsed, completing", "id", sess.ID)
		s.complete(sess)
		return nil
	}

	remaining := time.Until(sess.StopTime)
	slog.Info("reconcile: rescheduling active session", "id", sess.ID, "remaining", remaining)
	s.arm(sess, remaining)
	return nil
}

// arm schedules a completion timer for a session after d. The session is
// captured so the timer goroutine can reproduce its notification payload.
func (s *Server) arm(sess *store.Session, d time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if t, ok := s.timers[sess.ID]; ok {
		t.Stop()
	}
	s.timers[sess.ID] = time.AfterFunc(d, func() { s.complete(sess) })
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

// complete marks the session done and sends the completion notification (and
// email, if requested). It is invoked from the timer goroutine, so it must not
// assume a request context.
func (s *Server) complete(sess *store.Session) {
	s.mu.Lock()
	delete(s.timers, sess.ID)
	s.mu.Unlock()

	ctx := context.Background()
	if err := s.store.FinishActiveSession(ctx, store.StatusDone); err != nil {
		slog.Error("complete: finish session", "id", sess.ID, "err", err)
		return
	}

	body := sess.Message
	if body == "" {
		body = "Pomodoro session is ended!"
	}
	if err := s.notifier.Notify("Pomodoro", body, sess.Hint); err != nil {
		slog.Error("complete: notify", "id", sess.ID, "err", err)
	}

	if sess.Email {
		if err := mail.SendMail("Pomodoro", body); err != nil {
			slog.Error("complete: email", "id", sess.ID, "err", err)
		}
	}
}
