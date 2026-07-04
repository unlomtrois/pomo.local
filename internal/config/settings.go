package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/adrg/xdg"
	"pomo.local/internal/utils"
)

// Duration is a time.Duration that JSON-encodes as a human string ("25m") so
// settings.json stays hand-editable.
type Duration time.Duration

// MarshalJSON writes the duration as a compact Go duration string ("25m").
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(utils.ShortDuration(time.Duration(d)))
}

// UnmarshalJSON parses a Go duration string ("25m", "1h30m").
func (d *Duration) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	v, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*d = Duration(v)
	return nil
}

// Std returns the standard library duration.
func (d Duration) Std() time.Duration { return time.Duration(d) }

// Settings holds user-configurable durations (settings.json).
type Settings struct {
	DoroDuration Duration `json:"doro_duration"`
	LongDuration Duration `json:"long_duration"`
	RestDuration Duration `json:"rest_duration"`
}

// DefaultSettings returns the built-in defaults.
func DefaultSettings() *Settings {
	return &Settings{
		DoroDuration: Duration(25 * time.Minute),
		LongDuration: Duration(50 * time.Minute),
		RestDuration: Duration(5 * time.Minute),
	}
}

// SettingsPath is the settings.json location (does not create anything).
func SettingsPath() string {
	return filepath.Join(xdg.ConfigHome, "pomo", "settings.json")
}

// SettingsExist reports whether a settings file is present.
func SettingsExist() bool {
	_, err := os.Stat(SettingsPath())
	return err == nil
}

// LoadSettings reads settings.json, falling back to defaults for a missing file
// or any missing/invalid field. Best-effort: it never returns an error, so
// callers (including flag setup in init) can use it unconditionally.
func LoadSettings() *Settings {
	s := DefaultSettings()
	data, err := os.ReadFile(SettingsPath())
	if err != nil {
		return s // missing or unreadable → defaults
	}
	_ = json.Unmarshal(data, s) // partial/invalid JSON leaves defaults in place
	// Guard against present-but-nonpositive values.
	def := DefaultSettings()
	if s.DoroDuration <= 0 {
		s.DoroDuration = def.DoroDuration
	}
	if s.LongDuration <= 0 {
		s.LongDuration = def.LongDuration
	}
	if s.RestDuration <= 0 {
		s.RestDuration = def.RestDuration
	}
	return s
}

// Save writes the settings to settings.json (creating the directory).
func (s *Settings) Save() error {
	path := SettingsPath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
