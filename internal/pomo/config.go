package pomo

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type FileConfig struct {
	DefaultDuration string `json:"default_duration"`
	Toggl           struct {
		Token       string `json:"token"`
		WorkspaceID int    `json:"workspace_id"`
		UserID      int    `json:"user_id"`
	} `json:"toggl"`
}

func LoadConfig(configDir string) (*FileConfig, error) {
	path := filepath.Join(configDir, "config.json")

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &FileConfig{}, nil // return defaults
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
