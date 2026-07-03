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

	if _, err := st.StartSession(ctx, StartParams{Topic: "first", Duration: 25 * time.Minute, Source: "cli"}); err != nil {
		t.Fatalf("first start: %v", err)
	}

	_, err := st.StartSession(ctx, StartParams{Topic: "second", Duration: 5 * time.Minute, Source: "cli"})
	if !errors.Is(err, ErrActiveExists) {
		t.Fatalf("expected ErrActiveExists, got %v", err)
	}
}

func TestSessionLifecycle(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	started, err := st.StartSession(ctx, StartParams{Topic: "topic", Duration: time.Minute, Source: "mcp"})
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
	if _, err := st.StartSession(ctx, StartParams{Topic: "next", Duration: time.Minute, Source: "cli"}); err != nil {
		t.Fatalf("start after finish: %v", err)
	}
}

func TestMoveSession(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	started, err := st.StartSession(ctx, StartParams{Topic: "move me", Duration: 25 * time.Minute, Source: "cli"})
	if err != nil {
		t.Fatalf("start: %v", err)
	}

	newStart := started.StartTime.Add(48 * time.Hour)
	moved, err := st.MoveSession(ctx, started.ID, newStart)
	if err != nil {
		t.Fatalf("move: %v", err)
	}
	if !moved.StartTime.Equal(newStart.UTC()) {
		t.Fatalf("start not updated: got %v want %v", moved.StartTime, newStart.UTC())
	}
	if moved.Duration != started.Duration {
		t.Fatalf("duration changed: got %v want %v", moved.Duration, started.Duration)
	}
	if want := newStart.UTC().Add(started.Duration); !moved.StopTime.Equal(want) {
		t.Fatalf("stop not recomputed: got %v want %v", moved.StopTime, want)
	}

	// Persisted?
	got, err := st.GetSession(ctx, started.ID)
	if err != nil || !got.StartTime.Equal(newStart.UTC()) {
		t.Fatalf("reload after move: %+v, err=%v", got, err)
	}
}

func TestMoveSession_NotFound(t *testing.T) {
	st := newTestStore(t)
	if _, err := st.MoveSession(context.Background(), 404, time.Now()); !errors.Is(err, ErrSessionNotFound) {
		t.Fatalf("expected ErrSessionNotFound, got %v", err)
	}
}

func TestFinishActiveSession_NoActive(t *testing.T) {
	st := newTestStore(t)
	if err := st.FinishActiveSession(context.Background(), StatusDone); !errors.Is(err, ErrNoActive) {
		t.Fatalf("expected ErrNoActive, got %v", err)
	}
}
