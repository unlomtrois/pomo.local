package pomo

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type FileConfig struct {
	Pomodoro struct {
		DefaultDuration string `json:"default_duration,omitempty"` // "25m"
		DefaultMessage  string `json:"default_message,omitempty"`
	} `json:"pomodoro"`

	Rest struct {
		DefaultDuration string `json:"default_duration,omitempty"` // "5m"
		DefaultMessage  string `json:"default_message,omitempty"`
	} `json:"rest"`

	Toggl struct {
		Token       string `json:"token,omitempty"`
		WorkspaceID int    `json:"workspace_id,omitempty"`
		UserID      int    `json:"user_id,omitempty"`
	} `json:"toggl"`

	Notifications struct {
		Mute  bool   `json:"mute,omitempty"`
		Sound string `json:"sound,omitempty"`
	} `json:"notifications"`
}

func DefaultConfig() *FileConfig {
	cfg := &FileConfig{}
	cfg.Pomodoro.DefaultDuration = "25m"
	cfg.Rest.DefaultDuration = "5m"
	return cfg
}

func ConfigPath(configDir string) string {
	return filepath.Join(configDir, "config.json")
}

func LoadConfig(configDir string) (*FileConfig, error) {
	path := ConfigPath(configDir)

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		// Create default config
		cfg := DefaultConfig()
		if err := SaveConfig(configDir, cfg); err != nil {
			return nil, fmt.Errorf("creating default config: %w", err)
		}
		return cfg, nil
	}
	if err != nil {
		return nil, err
	}

	var cfg FileConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func SaveConfig(configDir string, cfg *FileConfig) error {
	// Ensure directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	path := ConfigPath(configDir)
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600) // 0600 for sensitive data
}
