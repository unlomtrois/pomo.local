// Package project implements git-style project markers: a .pomo directory whose
// location defines a project (like .git defines a repo). The CLI discovers the
// nearest .pomo by walking up from the working directory and tags sessions with
// its id and name.
package project

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// DirName is the marker directory created by `pomo init`.
const DirName = ".pomo"

// ErrExists is returned by Init when a .pomo already exists.
var ErrExists = errors.New("already a pomo project")

// ErrNotFound is returned by Find when no .pomo exists in the directory tree.
var ErrNotFound = errors.New("not inside a pomo project")

// Project is a discovered project marker.
type Project struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Root string `json:"-"` // directory containing .pomo
}

// config is the on-disk .pomo/config.json shape.
type config struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Init creates a .pomo marker in dir with a fresh id and a name defaulting to
// the directory's base name. It writes a self-ignoring .gitignore ("*") so the
// marker never shows up in the user's git status.
func Init(dir string) (*Project, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	pdir := filepath.Join(abs, DirName)
	if _, err := os.Stat(pdir); err == nil {
		return nil, ErrExists
	}
	if err := os.MkdirAll(pdir, 0755); err != nil {
		return nil, err
	}

	p := &Project{ID: newID(), Name: filepath.Base(abs), Root: abs}
	data, err := json.MarshalIndent(config{ID: p.ID, Name: p.Name}, "", "  ")
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(filepath.Join(pdir, "config.json"), data, 0644); err != nil {
		return nil, err
	}
	// "*" makes git treat .pomo as containing only ignored files, so it never
	// appears in git status and init doesn't touch the user's repo.
	if err := os.WriteFile(filepath.Join(pdir, ".gitignore"), []byte("*\n"), 0644); err != nil {
		return nil, err
	}
	return p, nil
}

// Find walks up from start looking for a .pomo/config.json, like git finds .git.
// Returns ErrNotFound if it reaches the filesystem root without one.
func Find(start string) (*Project, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return nil, err
	}
	for {
		cfgPath := filepath.Join(dir, DirName, "config.json")
		if data, err := os.ReadFile(cfgPath); err == nil {
			var c config
			if err := json.Unmarshal(data, &c); err != nil {
				return nil, fmt.Errorf("read %s: %w", cfgPath, err)
			}
			return &Project{ID: c.ID, Name: c.Name, Root: dir}, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return nil, ErrNotFound // reached filesystem root
		}
		dir = parent
	}
}

func newID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
