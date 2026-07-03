package project

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestInitAndFind(t *testing.T) {
	root := t.TempDir()

	p, err := Init(root)
	if err != nil {
		t.Fatalf("init: %v", err)
	}
	if p.ID == "" {
		t.Fatal("expected non-empty id")
	}
	if p.Name != filepath.Base(root) {
		t.Fatalf("name = %q, want %q", p.Name, filepath.Base(root))
	}

	// Marker files exist.
	if _, err := os.Stat(filepath.Join(root, DirName, "config.json")); err != nil {
		t.Fatalf("config.json missing: %v", err)
	}
	gi, err := os.ReadFile(filepath.Join(root, DirName, ".gitignore"))
	if err != nil || string(gi) != "*\n" {
		t.Fatalf(".gitignore = %q, err %v", gi, err)
	}

	// Find from a nested subdirectory walks up to the same project.
	deep := filepath.Join(root, "a", "b", "c")
	if err := os.MkdirAll(deep, 0755); err != nil {
		t.Fatal(err)
	}
	found, err := Find(deep)
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if found.ID != p.ID || found.Name != p.Name || found.Root != root {
		t.Fatalf("found = %+v, want id %s name %s root %s", found, p.ID, p.Name, root)
	}
}

func TestInitTwice(t *testing.T) {
	root := t.TempDir()
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	if _, err := Init(root); !errors.Is(err, ErrExists) {
		t.Fatalf("expected ErrExists, got %v", err)
	}
}

func TestFindNotFound(t *testing.T) {
	if _, err := Find(t.TempDir()); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
