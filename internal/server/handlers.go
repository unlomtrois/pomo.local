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
	Topic       string `json:"topic"`
	Duration    string `json:"duration"`
	Source      string `json:"source"`
	Message     string `json:"message"`
	Hint        string `json:"hint"`
	Email       bool   `json:"email"`
	Project     string `json:"project"`      // stable ext id from .pomo/config.json
	ProjectName string `json:"project_name"` // display name from .pomo
	Open        bool   `json:"open"`         // open-ended stopwatch (no timer)
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

	var projectID *int64
	if req.Project != "" {
		id, err := s.store.UpsertProject(r.Context(), req.Project, req.ProjectName)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		projectID = &id
	}

	sess, err := s.store.StartSession(r.Context(), store.StartParams{
		Topic:     req.Topic,
		Duration:  duration,
		Source:    source,
		Message:   req.Message,
		Hint:      req.Hint,
		Email:     req.Email,
		ProjectID: projectID,
		Open:      req.Open,
	})
	if errors.Is(err, store.ErrActiveExists) {
		writeError(w, http.StatusConflict, "you can only have 1 active pomodoro session at once")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Open stopwatches have no completion timer; only doros are armed.
	if !sess.Open {
		s.arm(sess, time.Until(sess.StopTime))
	}
	writeJSON(w, http.StatusCreated, sess)
}

// moveRequest is the PATCH /api/sessions/{id} body: the new start time as
// RFC3339. Duration is preserved, so stop_time is recomputed server-side.
type moveRequest struct {
	StartTime time.Time `json:"start_time"`
}

func (s *Server) handleMove(w http.ResponseWriter, r *http.Request) {
	id, err := parsePositiveInt64(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid session id")
		return
	}

	var req moveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.StartTime.IsZero() {
		writeError(w, http.StatusBadRequest, "start_time is required (RFC3339)")
		return
	}

	sess, err := s.store.MoveSession(r.Context(), id, req.StartTime)
	if errors.Is(err, store.ErrSessionNotFound) {
		writeError(w, http.StatusNotFound, "session not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// If the moved session is still the active one, re-arm its timer for the
	// new stop time so the completion notification fires at the right moment.
	if sess.Status == store.StatusActive {
		s.arm(sess, time.Until(sess.StopTime))
	}
	writeJSON(w, http.StatusOK, sess)
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

// endRequest is the optional POST /api/sessions/active/end body. A nil Topic
// keeps the session's existing topic; a non-empty one overrides it.
type endRequest struct {
	Topic *string `json:"topic"`
}

func (s *Server) handleEnd(w http.ResponseWriter, r *http.Request) {
	var req endRequest
	// Body is optional (`pomo end` with no topic sends none).
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}

	// Disarm any timer (a doro stopped early) before closing the session.
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

	ended, err := s.store.EndActiveSession(r.Context(), req.Topic)
	if errors.Is(err, store.ErrNoActive) {
		writeError(w, http.StatusNotFound, "no active pomodoro session")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, ended)
}

func (s *Server) handleByHash(w http.ResponseWriter, r *http.Request) {
	prefix := r.PathValue("prefix")
	if len(prefix) < 4 {
		writeError(w, http.StatusBadRequest, "hash prefix must be at least 4 characters")
		return
	}
	sess, err := s.store.SessionByHashPrefix(r.Context(), prefix)
	if errors.Is(err, store.ErrSessionNotFound) {
		writeError(w, http.StatusNotFound, "no session matching "+prefix)
		return
	}
	if errors.Is(err, store.ErrAmbiguousHash) {
		writeError(w, http.StatusConflict, "ambiguous hash prefix "+prefix)
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, sess)
}

// editRequest is the PATCH /api/sessions/by-hash/{prefix} body. All fields are
// optional; only present ones change (partial update). Duration is a Go string.
type editRequest struct {
	Topic     *string    `json:"topic"`
	StartTime *time.Time `json:"start_time"`
	Duration  *string    `json:"duration"`
}

func (s *Server) handleEdit(w http.ResponseWriter, r *http.Request) {
	prefix := r.PathValue("prefix")
	if len(prefix) < 4 {
		writeError(w, http.StatusBadRequest, "hash prefix must be at least 4 characters")
		return
	}

	var req editRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	params := store.EditParams{Topic: req.Topic, Start: req.StartTime}
	if req.Duration != nil {
		d, err := time.ParseDuration(*req.Duration)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid duration: "+err.Error())
			return
		}
		params.Duration = &d
	}

	sess, err := s.store.SessionByHashPrefix(r.Context(), prefix)
	if errors.Is(err, store.ErrSessionNotFound) {
		writeError(w, http.StatusNotFound, "no session matching "+prefix)
		return
	}
	if errors.Is(err, store.ErrAmbiguousHash) {
		writeError(w, http.StatusConflict, "ambiguous hash prefix "+prefix)
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	updated, err := s.store.EditSession(r.Context(), sess.ID, params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Re-arm the completion timer if we changed the timing of the active doro.
	if updated.Status == store.StatusActive && !updated.Open {
		s.arm(updated, time.Until(updated.StopTime))
	}
	writeJSON(w, http.StatusOK, updated)
}

func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request) {
	prefix := r.PathValue("prefix")
	if len(prefix) < 4 {
		writeError(w, http.StatusBadRequest, "hash prefix must be at least 4 characters")
		return
	}
	sess, err := s.store.SessionByHashPrefix(r.Context(), prefix)
	if errors.Is(err, store.ErrSessionNotFound) {
		writeError(w, http.StatusNotFound, "no session matching "+prefix)
		return
	}
	if errors.Is(err, store.ErrAmbiguousHash) {
		writeError(w, http.StatusConflict, "ambiguous hash prefix "+prefix)
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// If the deleted session is the active one, cancel any pending timer so it
	// doesn't fire against a now-nonexistent row.
	s.disarm(sess.ID)

	if err := s.store.DeleteSession(r.Context(), sess.ID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, sess) // return the deleted session for confirmation
}

func (s *Server) handleList(w http.ResponseWriter, r *http.Request) {
	limit := 50
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := parsePositiveInt(v); err == nil {
			limit = n
		}
	}

	// ?project=<ext_id> scopes the list to a single project (used by `pomo log`).
	var sessions []*store.Session
	var err error
	if proj := r.URL.Query().Get("project"); proj != "" {
		sessions, err = s.store.ListSessionsByProject(r.Context(), proj, limit)
	} else {
		sessions, err = s.store.ListSessions(r.Context(), limit)
	}
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

func parsePositiveInt64(s string) (int64, error) {
	n, err := parsePositiveInt(s)
	return int64(n), err
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
