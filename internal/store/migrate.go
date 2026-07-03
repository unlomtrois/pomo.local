package store

import (
	"context"
	"fmt"
)

// schema is applied idempotently on every Open. Keep statements additive and
// guarded with IF NOT EXISTS so startup is safe against an existing database.
//
// The sessions table mirrors pomo.Session plus daemon-era fields:
//   - status  : "active" | "done" | "cancelled" — replaces the old
//     "active_session.json exists?" check with a queryable state.
//   - source  : "cli" | "web" | "mcp" — free provenance analytics.
//   - project_id: optional FK, the seed of the Toggl-style project grouping.
const schema = `
CREATE TABLE IF NOT EXISTS projects (
	id    INTEGER PRIMARY KEY AUTOINCREMENT,
	name  TEXT NOT NULL UNIQUE,
	color TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS sessions (
	id         INTEGER PRIMARY KEY AUTOINCREMENT,
	topic      TEXT NOT NULL,
	project_id INTEGER REFERENCES projects(id) ON DELETE SET NULL,
	start_time TEXT NOT NULL,
	stop_time  TEXT NOT NULL,
	duration   INTEGER NOT NULL, -- nanoseconds, matches time.Duration
	status     TEXT NOT NULL DEFAULT 'active',
	source     TEXT NOT NULL DEFAULT 'cli',
	created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_sessions_status ON sessions(status);
CREATE INDEX IF NOT EXISTS idx_sessions_start ON sessions(start_time);

-- At most one active session at a time: a partial unique index enforces the
-- invariant that the JSON-file check used to guard, but atomically.
CREATE UNIQUE INDEX IF NOT EXISTS idx_sessions_single_active
	ON sessions(status) WHERE status = 'active';
`

func (s *Store) migrate(ctx context.Context) error {
	if _, err := s.db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	return nil
}
