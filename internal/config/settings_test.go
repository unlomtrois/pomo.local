package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/adrg/xdg"
)

func TestLoadSettingsDefaults(t *testing.T) {
	xdg.ConfigHome = t.TempDir() // no file present

	s := LoadSettings()
	if s.DoroDuration.Std() != 25*time.Minute ||
		s.LongDuration.Std() != 50*time.Minute ||
		s.RestDuration.Std() != 5*time.Minute {
		t.Fatalf("defaults wrong: %+v", s)
	}
}

func TestSettingsRoundTripAndPartial(t *testing.T) {
	xdg.ConfigHome = t.TempDir()

	// Save custom, reload.
	orig := &Settings{
		DoroDuration: Duration(30 * time.Minute),
		LongDuration: Duration(90 * time.Minute),
		RestDuration: Duration(10 * time.Minute),
	}
	if err := orig.Save(); err != nil {
		t.Fatalf("save: %v", err)
	}
	got := LoadSettings()
	if got.DoroDuration.Std() != 30*time.Minute || got.LongDuration.Std() != 90*time.Minute {
		t.Fatalf("round-trip wrong: %+v", got)
	}

	// A partial file keeps defaults for the missing field.
	if err := os.MkdirAll(filepath.Dir(SettingsPath()), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(SettingsPath(), []byte(`{"doro_duration":"40m"}`), 0644); err != nil {
		t.Fatal(err)
	}
	partial := LoadSettings()
	if partial.DoroDuration.Std() != 40*time.Minute {
		t.Fatalf("doro not read: %+v", partial)
	}
	if partial.LongDuration.Std() != 50*time.Minute { // default preserved
		t.Fatalf("missing long should default, got %v", partial.LongDuration.Std())
	}
}
