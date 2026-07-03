package store

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestSessionHashAssignedAndUnique(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	seen := map[string]bool{}
	for i := 0; i < 5; i++ {
		s, err := st.StartSession(ctx, StartParams{Topic: "x", Duration: time.Minute, Source: "cli"})
		if err != nil {
			t.Fatalf("start: %v", err)
		}
		if len(s.Hash) != 40 {
			t.Fatalf("hash len = %d, want 40: %q", len(s.Hash), s.Hash)
		}
		if seen[s.Hash] {
			t.Fatalf("duplicate hash %q", s.Hash)
		}
		seen[s.Hash] = true
		if err := st.FinishActiveSession(ctx, StatusDone); err != nil {
			t.Fatal(err)
		}
	}
}

func TestSessionByHashPrefix(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	s, err := st.StartSession(ctx, StartParams{Topic: "find me", Duration: time.Minute, Source: "cli"})
	if err != nil {
		t.Fatal(err)
	}

	got, err := st.SessionByHashPrefix(ctx, s.Hash[:7])
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	if got.ID != s.ID {
		t.Fatalf("got id %d, want %d", got.ID, s.ID)
	}

	if _, err := st.SessionByHashPrefix(ctx, "0000000000000000"); !errors.Is(err, ErrSessionNotFound) {
		t.Fatalf("expected ErrSessionNotFound, got %v", err)
	}
}

func TestSessionByHashPrefix_Ambiguous(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	// Force two sessions whose hashes share a prefix by writing them directly.
	for _, h := range []string{"abcd111111", "abcd222222"} {
		if _, err := st.db.ExecContext(ctx,
			`INSERT INTO sessions (hash, topic, start_time, stop_time, duration, status, source)
			 VALUES (?, 't', '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z', 0, 'done', 'cli')`, h); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := st.SessionByHashPrefix(ctx, "abcd"); !errors.Is(err, ErrAmbiguousHash) {
		t.Fatalf("expected ErrAmbiguousHash, got %v", err)
	}
}
