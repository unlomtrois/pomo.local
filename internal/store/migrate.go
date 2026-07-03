package store

import (
	"context"
	"fmt"
	"strings"
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
	id     INTEGER PRIMARY KEY AUTOINCREMENT,
	name   TEXT NOT NULL,
	color  TEXT NOT NULL DEFAULT '',
	-- ext_id is the stable id from a project's .pomo/config.json; NULL for
	-- projects not backed by a .pomo marker.
	ext_id TEXT
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
	-- completion-notification payload, carried so the daemon's timer can
	-- reproduce the notify-send + optional email the old scheduler flow did.
	-- stable, opaque per-entry id (like a git object name), assigned at creation.
	hash       TEXT,
	message    TEXT NOT NULL DEFAULT '',
	hint       TEXT NOT NULL DEFAULT '',
	email      INTEGER NOT NULL DEFAULT 0,
	-- open = 1 marks an open-ended stopwatch session (pomo start/end): no fixed
	-- stop time and no completion timer; stop_time/duration are filled on end.
	open       INTEGER NOT NULL DEFAULT 0,
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
	// Bring a pre-existing projects table (created before ext_id) up to date.
	// SQLite has no ADD COLUMN IF NOT EXISTS, so ignore the duplicate-column error.
	if _, err := s.db.ExecContext(ctx, `ALTER TABLE projects ADD COLUMN ext_id TEXT`); err != nil {
		if !strings.Contains(err.Error(), "duplicate column name") {
			return fmt.Errorf("migrate ext_id: %w", err)
		}
	}
	// Unique index enables ON CONFLICT(ext_id) upserts (multiple NULLs allowed).
	if _, err := s.db.ExecContext(ctx,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_projects_ext ON projects(ext_id)`); err != nil {
		return fmt.Errorf("migrate ext_id index: %w", err)
	}
	// Add the open-session flag to pre-existing sessions tables.
	if _, err := s.db.ExecContext(ctx, `ALTER TABLE sessions ADD COLUMN open INTEGER NOT NULL DEFAULT 0`); err != nil {
		if !strings.Contains(err.Error(), "duplicate column name") {
			return fmt.Errorf("migrate open: %w", err)
		}
	}
	// Add the per-entry hash to pre-existing sessions tables.
	if _, err := s.db.ExecContext(ctx, `ALTER TABLE sessions ADD COLUMN hash TEXT`); err != nil {
		if !strings.Contains(err.Error(), "duplicate column name") {
			return fmt.Errorf("migrate hash: %w", err)
		}
	}
	if err := s.backfillHashes(ctx); err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_sessions_hash ON sessions(hash)`); err != nil {
		return fmt.Errorf("migrate hash index: %w", err)
	}
	return nil
}

// backfillHashes assigns a hash to any pre-existing rows that lack one (must run
// before the unique index is created).
func (s *Store) backfillHashes(ctx context.Context) error {
	rows, err := s.db.QueryContext(ctx, `SELECT id FROM sessions WHERE hash IS NULL OR hash = ''`)
	if err != nil {
		return fmt.Errorf("backfill scan: %w", err)
	}
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			_ = rows.Close()
			return err
		}
		ids = append(ids, id)
	}
	_ = rows.Close()

	for _, id := range ids {
		if _, err := s.db.ExecContext(ctx,
			`UPDATE sessions SET hash = ? WHERE id = ?`, newHash(), id); err != nil {
			return fmt.Errorf("backfill hash: %w", err)
		}
	}
	return nil
}
