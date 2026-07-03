package store

import (
	"context"
	"testing"
	"time"
)

func TestUpsertProjectIdempotent(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	id1, err := st.UpsertProject(ctx, "ext-abc", "myproj")
	if err != nil {
		t.Fatalf("upsert: %v", err)
	}
	// Same ext_id returns the same row and updates the name.
	id2, err := st.UpsertProject(ctx, "ext-abc", "renamed")
	if err != nil {
		t.Fatalf("upsert 2: %v", err)
	}
	if id1 != id2 {
		t.Fatalf("expected same id, got %d and %d", id1, id2)
	}

	// A different ext_id is a distinct project.
	id3, err := st.UpsertProject(ctx, "ext-xyz", "other")
	if err != nil {
		t.Fatalf("upsert 3: %v", err)
	}
	if id3 == id1 {
		t.Fatalf("expected distinct id for different ext_id")
	}
}

func TestListSessionsByProject(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	pa, _ := st.UpsertProject(ctx, "ext-a", "A")
	pb, _ := st.UpsertProject(ctx, "ext-b", "B")

	mk := func(topic string, pid int64) {
		if _, err := st.StartSession(ctx, StartParams{Topic: topic, Duration: time.Minute, Source: "cli", ProjectID: &pid}); err != nil {
			t.Fatal(err)
		}
		if err := st.FinishActiveSession(ctx, StatusDone); err != nil {
			t.Fatal(err)
		}
	}
	mk("a1", pa)
	mk("a2", pa)
	mk("b1", pb)

	a, err := st.ListSessionsByProject(ctx, "ext-a", 50)
	if err != nil {
		t.Fatalf("list a: %v", err)
	}
	if len(a) != 2 {
		t.Fatalf("project A: got %d sessions, want 2", len(a))
	}

	// Unknown ext id → no rows (not an error).
	none, err := st.ListSessionsByProject(ctx, "ext-missing", 50)
	if err != nil || len(none) != 0 {
		t.Fatalf("unknown project: got %d sessions, err %v", len(none), err)
	}
}

func TestStartSessionWithProject(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	pid, err := st.UpsertProject(ctx, "ext-abc", "myproj")
	if err != nil {
		t.Fatalf("upsert: %v", err)
	}

	sess, err := st.StartSession(ctx, StartParams{
		Topic: "tagged", Duration: 25 * time.Minute, Source: "cli", ProjectID: &pid,
	})
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	if sess.ProjectID == nil || *sess.ProjectID != pid {
		t.Fatalf("session project_id = %v, want %d", sess.ProjectID, pid)
	}

	got, err := st.GetSession(ctx, sess.ID)
	if err != nil || got.ProjectID == nil || *got.ProjectID != pid {
		t.Fatalf("reloaded project_id = %v, err %v", got.ProjectID, err)
	}
}
