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
