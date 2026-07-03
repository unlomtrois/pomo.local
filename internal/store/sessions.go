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

// Session is a stored pomodoro session. It is the daemon-era superset of
// pomo.Session, adding identity, lifecycle status, provenance, and the
// completion-notification payload.
type Session struct {
	ID        int64         `json:"id"`
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
}

// StartParams describes a session to start.
type StartParams struct {
	Topic    string
	Duration time.Duration
	Source   string
	Message  string
	Hint     string
	Email    bool
}

// StartSession inserts a new active session. It fails with ErrActiveExists if
// one is already active, enforced atomically by the partial unique index.
func (s *Store) StartSession(ctx context.Context, p StartParams) (*Session, error) {
	start := time.Now().UTC()
	stop := start.Add(p.Duration)

	res, err := s.db.ExecContext(ctx,
		`INSERT INTO sessions (topic, start_time, stop_time, duration, status, source, message, hint, email)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.Topic, start.Format(time.RFC3339Nano), stop.Format(time.RFC3339Nano),
		int64(p.Duration), StatusActive, p.Source, p.Message, p.Hint, boolToInt(p.Email),
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
		ID: id, Topic: p.Topic, StartTime: start, StopTime: stop,
		Duration: p.Duration, Status: StatusActive, Source: p.Source,
		Message: p.Message, Hint: p.Hint, Email: p.Email,
	}, nil
}

// ActiveSession returns the currently active session, or ErrNoActive if none.
func (s *Store) ActiveSession(ctx context.Context) (*Session, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, topic, project_id, start_time, stop_time, duration, status, source, message, hint, email
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

// ListSessions returns the most recent sessions, newest first, capped at limit.
func (s *Store) ListSessions(ctx context.Context, limit int) ([]*Session, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, topic, project_id, start_time, stop_time, duration, status, source, message, hint, email
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
	)
	if err := sc.Scan(&sess.ID, &sess.Topic, &projectID, &start, &stop, &durNanos,
		&sess.Status, &sess.Source, &sess.Message, &sess.Hint, &email); err != nil {
		return nil, err
	}
	if projectID.Valid {
		sess.ProjectID = &projectID.Int64
	}
	sess.Email = email != 0
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
