package store

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	st, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	return st
}

func TestStartSession_SingleActiveInvariant(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	if _, err := st.StartSession(ctx, "first", 25*time.Minute, "cli"); err != nil {
		t.Fatalf("first start: %v", err)
	}

	_, err := st.StartSession(ctx, "second", 5*time.Minute, "cli")
	if !errors.Is(err, ErrActiveExists) {
		t.Fatalf("expected ErrActiveExists, got %v", err)
	}
}

func TestSessionLifecycle(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	started, err := st.StartSession(ctx, "topic", time.Minute, "mcp")
	if err != nil {
		t.Fatalf("start: %v", err)
	}

	active, err := st.ActiveSession(ctx)
	if err != nil {
		t.Fatalf("active: %v", err)
	}
	if active.ID != started.ID || active.Topic != "topic" || active.Source != "mcp" {
		t.Fatalf("active mismatch: %+v", active)
	}

	if err := st.FinishActiveSession(ctx, StatusDone); err != nil {
		t.Fatalf("finish: %v", err)
	}

	if _, err := st.ActiveSession(ctx); !errors.Is(err, ErrNoActive) {
		t.Fatalf("expected ErrNoActive after finish, got %v", err)
	}

	// A completed session frees the slot for a new one.
	if _, err := st.StartSession(ctx, "next", time.Minute, "cli"); err != nil {
		t.Fatalf("start after finish: %v", err)
	}
}

func TestFinishActiveSession_NoActive(t *testing.T) {
	st := newTestStore(t)
	if err := st.FinishActiveSession(context.Background(), StatusDone); !errors.Is(err, ErrNoActive) {
		t.Fatalf("expected ErrNoActive, got %v", err)
	}
}
