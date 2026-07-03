package store

import (
	"context"
	"fmt"
)

// UpsertProject inserts (or updates the name of) a project identified by its
// stable external id (from .pomo/config.json) and returns its internal row id,
// suitable for sessions.project_id.
func (s *Store) UpsertProject(ctx context.Context, extID, name string) (int64, error) {
	var id int64
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO projects (ext_id, name) VALUES (?, ?)
		 ON CONFLICT(ext_id) DO UPDATE SET name = excluded.name
		 RETURNING id`, extID, name).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("upsert project: %w", err)
	}
	return id, nil
}
