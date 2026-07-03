// Package store provides the SQLite-backed persistence layer owned by the pomo daemon.
//
// The daemon is the sole writer to this database; the CLI, web UI, and MCP
// server all reach it through the daemon's HTTP API rather than opening the
// database directly. This keeps SQLite's single-writer model simple and avoids
// the concurrent-writer races the file-based state (JSON/CSV) was prone to.
package store

import (
	"context"
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite" // pure-Go SQLite driver (no cgo)
)

// Store wraps the SQLite database connection and exposes session operations.
type Store struct {
	db *sql.DB
}

// Open opens (creating if needed) the SQLite database at path, applies
// pragmatic connection settings, and runs migrations.
//
// WAL mode plus a single writer lets readers (web/CLI/MCP GETs) proceed without
// blocking the daemon's writes. busy_timeout guards the brief windows where a
// checkpoint or another connection holds the write lock.
func Open(path string) (*Store, error) {
	dsn := fmt.Sprintf("file:%s?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(ON)", path)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	s := &Store{db: db}
	if err := s.migrate(context.Background()); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

// Close closes the underlying database.
func (s *Store) Close() error {
	return s.db.Close()
}
