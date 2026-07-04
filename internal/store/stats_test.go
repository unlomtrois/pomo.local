package store

import (
	"context"
	"testing"
	"time"
)

func TestStats(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	pid, _ := st.UpsertProject(ctx, "ext-a", "A")

	// Two done sessions (25m + 35m) and one cancelled.
	mk := func(dur time.Duration, status string) {
		if _, err := st.StartSession(ctx, StartParams{Topic: "x", Duration: dur, Source: "cli", ProjectID: &pid}); err != nil {
			t.Fatal(err)
		}
		if err := st.FinishActiveSession(ctx, status); err != nil {
			t.Fatal(err)
		}
	}
	mk(25*time.Minute, StatusDone)
	mk(35*time.Minute, StatusDone)
	mk(10*time.Minute, StatusCancelled)

	got, err := st.Stats(ctx)
	if err != nil {
		t.Fatalf("stats: %v", err)
	}
	if got.TotalSessions != 3 {
		t.Fatalf("total = %d, want 3", got.TotalSessions)
	}
	if got.ByStatus[StatusDone] != 2 || got.ByStatus[StatusCancelled] != 1 {
		t.Fatalf("by-status wrong: %+v", got.ByStatus)
	}
	if got.Projects != 1 {
		t.Fatalf("projects = %d, want 1", got.Projects)
	}
	// tracked = 25m + 35m = 60m (cancelled excluded)
	if want := int64(60 * time.Minute); got.TrackedNanos != want {
		t.Fatalf("tracked = %d, want %d", got.TrackedNanos, want)
	}
}
