package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// Status values for a session row.
const (
	StatusActive    = "active"
	StatusDone      = "done"
	StatusCancelled = "cancelled"
)

// ErrActiveExists is returned by StartSession when a session is already active.
// It replaces the old fs.ErrExist signal from the JSON-file check.
var ErrActiveExists = errors.New("an active pomodoro session already exists")

// ErrNoActive is returned when an operation needs an active session and none exists.
var ErrNoActive = errors.New("no active pomodoro session")

// ErrSessionNotFound is returned when a session with the given id does not exist.
var ErrSessionNotFound = errors.New("session not found")

// ErrAmbiguousHash is returned when a hash prefix matches more than one session.
var ErrAmbiguousHash = errors.New("ambiguous session hash prefix")

// Session is a stored pomodoro session. It is the daemon-era superset of
// pomo.Session, adding identity, lifecycle status, provenance, and the
// completion-notification payload.
type Session struct {
	ID        int64         `json:"id"`
	Hash      string        `json:"hash"`
	Topic     string        `json:"topic"`
	ProjectID *int64        `json:"project_id,omitempty"`
	StartTime time.Time     `json:"start_time"`
	StopTime  time.Time     `json:"stop_time"`
	Duration  time.Duration `json:"duration"`
	Status    string        `json:"status"`
	Source    string        `json:"source"`
	Message   string        `json:"message"`
	Hint      string        `json:"hint"`
	Email     bool          `json:"email"`
	Open      bool          `json:"open"`
}

// StartParams describes a session to start.
type StartParams struct {
	Topic     string
	Duration  time.Duration
	Source    string
	Message   string
	Hint      string
	Email     bool
	ProjectID *int64 // internal projects.id, or nil for no project
	Open      bool   // open-ended stopwatch: no fixed stop, no timer
}

// StartSession inserts a new active session. It fails with ErrActiveExists if
// one is already active, enforced atomically by the partial unique index.
func (s *Store) StartSession(ctx context.Context, p StartParams) (*Session, error) {
	start := time.Now().UTC()
	// Open sessions have no fixed duration; stop_time/duration are placeholders
	// until EndActiveSession fills them.
	if p.Open {
		p.Duration = 0
	}
	stop := start.Add(p.Duration)

	var projectID any
	if p.ProjectID != nil {
		projectID = *p.ProjectID
	}
	hash := newHash()
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO sessions (hash, topic, project_id, start_time, stop_time, duration, status, source, message, hint, email, open)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		hash, p.Topic, projectID, start.Format(time.RFC3339Nano), stop.Format(time.RFC3339Nano),
		int64(p.Duration), StatusActive, p.Source, p.Message, p.Hint, boolToInt(p.Email), boolToInt(p.Open),
	)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrActiveExists
		}
		return nil, fmt.Errorf("start session: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("start session id: %w", err)
	}

	return &Session{
		ID: id, Hash: hash, Topic: p.Topic, ProjectID: p.ProjectID, StartTime: start, StopTime: stop,
		Duration: p.Duration, Status: StatusActive, Source: p.Source,
		Message: p.Message, Hint: p.Hint, Email: p.Email, Open: p.Open,
	}, nil
}

// ActiveSession returns the currently active session, or ErrNoActive if none.
func (s *Store) ActiveSession(ctx context.Context) (*Session, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, hash, topic, project_id, start_time, stop_time, duration, status, source, message, hint, email, open
		 FROM sessions WHERE status = ? LIMIT 1`, StatusActive)

	sess, err := scanSession(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNoActive
	}
	return sess, err
}

// FinishActiveSession marks the active session with the given terminal status
// (StatusDone or StatusCancelled). Returns ErrNoActive if there is none.
func (s *Store) FinishActiveSession(ctx context.Context, status string) error {
	res, err := s.db.ExecContext(ctx,
		`UPDATE sessions SET status = ? WHERE status = ?`, status, StatusActive)
	if err != nil {
		return fmt.Errorf("finish session: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("finish session rows: %w", err)
	}
	if n == 0 {
		return ErrNoActive
	}
	return nil
}

// EndActiveSession closes the active session (open stopwatch or a doro stopped
// early): records the actual elapsed time, marks it done, clears the open flag,
// and optionally overrides the topic. Returns ErrNoActive if none is active.
func (s *Store) EndActiveSession(ctx context.Context, topic *string) (*Session, error) {
	sess, err := s.ActiveSession(ctx)
	if err != nil {
		return nil, err
	}

	stop := time.Now().UTC()
	dur := stop.Sub(sess.StartTime)
	if dur < 0 {
		dur = 0
	}
	newTopic := sess.Topic
	if topic != nil && *topic != "" {
		newTopic = *topic
	}

	if _, err := s.db.ExecContext(ctx,
		`UPDATE sessions SET stop_time = ?, duration = ?, status = ?, topic = ?, open = 0 WHERE id = ?`,
		stop.Format(time.RFC3339Nano), int64(dur), StatusDone, newTopic, sess.ID,
	); err != nil {
		return nil, fmt.Errorf("end session: %w", err)
	}

	sess.StopTime = stop
	sess.Duration = dur
	sess.Status = StatusDone
	sess.Topic = newTopic
	sess.Open = false
	return sess, nil
}

// GetSession returns a single session by id, or ErrSessionNotFound.
func (s *Store) GetSession(ctx context.Context, id int64) (*Session, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, hash, topic, project_id, start_time, stop_time, duration, status, source, message, hint, email, open
		 FROM sessions WHERE id = ?`, id)
	sess, err := scanSession(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrSessionNotFound
	}
	return sess, err
}

// MoveSession changes a session's start time to newStart, preserving its
// duration (stop_time is recomputed). Used by the calendar drag-and-drop.
// Returns the updated session, or ErrSessionNotFound.
func (s *Store) MoveSession(ctx context.Context, id int64, newStart time.Time) (*Session, error) {
	sess, err := s.GetSession(ctx, id)
	if err != nil {
		return nil, err
	}

	start := newStart.UTC()
	stop := start.Add(sess.Duration)
	if _, err := s.db.ExecContext(ctx,
		`UPDATE sessions SET start_time = ?, stop_time = ? WHERE id = ?`,
		start.Format(time.RFC3339Nano), stop.Format(time.RFC3339Nano), id,
	); err != nil {
		return nil, fmt.Errorf("move session: %w", err)
	}

	sess.StartTime = start
	sess.StopTime = stop
	return sess, nil
}

// SessionByHashPrefix resolves a session by a hash prefix, git-style: returns
// ErrSessionNotFound if nothing matches, ErrAmbiguousHash if more than one does.
func (s *Store) SessionByHashPrefix(ctx context.Context, prefix string) (*Session, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, hash, topic, project_id, start_time, stop_time, duration, status, source, message, hint, email, open
		 FROM sessions WHERE hash LIKE ? || '%' LIMIT 2`, prefix)
	if err != nil {
		return nil, fmt.Errorf("lookup hash: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var found []*Session
	for rows.Next() {
		sess, err := scanSession(rows)
		if err != nil {
			return nil, err
		}
		found = append(found, sess)
	}
	switch len(found) {
	case 0:
		return nil, ErrSessionNotFound
	case 1:
		return found[0], nil
	default:
		return nil, ErrAmbiguousHash
	}
}

// ListSessions returns the most recent sessions, newest first, capped at limit.
func (s *Store) ListSessions(ctx context.Context, limit int) ([]*Session, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, hash, topic, project_id, start_time, stop_time, duration, status, source, message, hint, email, open
		 FROM sessions ORDER BY start_time DESC LIMIT ?`, limit)
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []*Session
	for rows.Next() {
		sess, err := scanSession(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, sess)
	}
	return out, rows.Err()
}

// rowScanner is satisfied by both *sql.Row and *sql.Rows.
type rowScanner interface {
	Scan(dest ...any) error
}

func scanSession(sc rowScanner) (*Session, error) {
	var (
		sess      Session
		projectID sql.NullInt64
		start     string
		stop      string
		durNanos  int64
		email     int64
		open      int64
	)
	if err := sc.Scan(&sess.ID, &sess.Hash, &sess.Topic, &projectID, &start, &stop, &durNanos,
		&sess.Status, &sess.Source, &sess.Message, &sess.Hint, &email, &open); err != nil {
		return nil, err
	}
	if projectID.Valid {
		sess.ProjectID = &projectID.Int64
	}
	sess.Email = email != 0
	sess.Open = open != 0
	var err error
	if sess.StartTime, err = time.Parse(time.RFC3339Nano, start); err != nil {
		return nil, fmt.Errorf("parse start_time: %w", err)
	}
	if sess.StopTime, err = time.Parse(time.RFC3339Nano, stop); err != nil {
		return nil, fmt.Errorf("parse stop_time: %w", err)
	}
	sess.Duration = time.Duration(durNanos)
	return &sess, nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
