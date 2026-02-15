package config

import (
	"path/filepath"
	"testing"

	"github.com/adrg/xdg"
)

func TestMailConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configDirFunc = func(relPath string) (string, error) { return filepath.Join(tmpDir, relPath), nil }
	defer func() { configDirFunc = xdg.ConfigFile }()

	cfg := &MailConfig{
		Host: "smtp.gmail.com",
		Port: 587,
	}

	if err := cfg.Load(); err == nil {
		t.Error("Expected error when loading non-existent config, got nil.")
	}

	if err := cfg.Save(); err != nil {
		t.Fatalf("Could not save mail config: %v", err)
	}

	newCfg := &MailConfig{}
	if err := newCfg.Load(); err != nil {
		t.Fatalf("Could not load mail config: %v", err)
	}

	if newCfg.Host != cfg.Host || newCfg.Port != cfg.Port {
		t.Fatalf("Data mismatch! Expected %v, got %v", cfg, newCfg)
	}
}
