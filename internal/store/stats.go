package store

import (
	"context"
	"fmt"
)

// Stats is an aggregate summary of the database.
type Stats struct {
	TotalSessions int            `json:"total_sessions"`
	ByStatus      map[string]int `json:"by_status"`      // active/done/cancelled → count
	Projects      int            `json:"projects"`       // number of known projects
	TrackedNanos  int64          `json:"tracked_nanos"`  // sum of done-session durations
}

// Stats returns aggregate counts across sessions and projects.
func (s *Store) Stats(ctx context.Context) (*Stats, error) {
	out := &Stats{ByStatus: map[string]int{}}

	rows, err := s.db.QueryContext(ctx,
		`SELECT status, COUNT(*), COALESCE(SUM(duration), 0) FROM sessions GROUP BY status`)
	if err != nil {
		return nil, fmt.Errorf("stats sessions: %w", err)
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var status string
		var count int
		var durNanos int64
		if err := rows.Scan(&status, &count, &durNanos); err != nil {
			return nil, err
		}
		out.ByStatus[status] = count
		out.TotalSessions += count
		if status == StatusDone {
			out.TrackedNanos += durNanos
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM projects`).Scan(&out.Projects); err != nil {
		return nil, fmt.Errorf("stats projects: %w", err)
	}
	return out, nil
}
