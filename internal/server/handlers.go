package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"pomo.local/internal/store"
)

// startRequest is the POST /api/sessions body. Duration is a Go duration string
// (e.g. "25m", "1h") so the wire format matches the CLI's -d flag.
type startRequest struct {
	Topic    string `json:"topic"`
	Duration string `json:"duration"`
	Source   string `json:"source"`
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleStart(w http.ResponseWriter, r *http.Request) {
	var req startRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	duration := 25 * time.Minute
	if req.Duration != "" {
		d, err := time.ParseDuration(req.Duration)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid duration: "+err.Error())
			return
		}
		duration = d
	}
	source := req.Source
	if source == "" {
		source = "cli"
	}

	sess, err := s.store.StartSession(r.Context(), req.Topic, duration, source)
	if errors.Is(err, store.ErrActiveExists) {
		writeError(w, http.StatusConflict, "you can only have 1 active pomodoro session at once")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.arm(sess.ID, sess.Topic, time.Until(sess.StopTime))
	writeJSON(w, http.StatusCreated, sess)
}

func (s *Server) handleActive(w http.ResponseWriter, r *http.Request) {
	sess, err := s.store.ActiveSession(r.Context())
	if errors.Is(err, store.ErrNoActive) {
		writeError(w, http.StatusNotFound, "no active pomodoro session")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, sess)
}

func (s *Server) handleStop(w http.ResponseWriter, r *http.Request) {
	sess, err := s.store.ActiveSession(r.Context())
	if errors.Is(err, store.ErrNoActive) {
		writeError(w, http.StatusNotFound, "no active pomodoro session")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.disarm(sess.ID)
	if err := s.store.FinishActiveSession(r.Context(), store.StatusCancelled); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	sess.Status = store.StatusCancelled
	writeJSON(w, http.StatusOK, sess)
}

func (s *Server) handleList(w http.ResponseWriter, r *http.Request) {
	limit := 50
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := parsePositiveInt(v); err == nil {
			limit = n
		}
	}
	sessions, err := s.store.ListSessions(r.Context(), limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if sessions == nil {
		sessions = []*store.Session{}
	}
	writeJSON(w, http.StatusOK, sessions)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func parsePositiveInt(s string) (int, error) {
	var n int
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, errors.New("not a number")
		}
		n = n*10 + int(c-'0')
	}
	if n == 0 {
		return 0, errors.New("zero")
	}
	return n, nil
}
