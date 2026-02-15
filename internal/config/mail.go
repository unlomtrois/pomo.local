package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type MailConfig struct {
	Host     string `json:"host"`     // e.g. "smtp.gmail.com"
	Port     int    `json:"port"`     // e.g., 587
	Sender   string `json:"sender"`   // your email
	Receiver string `json:"receiver"` // email for notifications
}

func (m *MailConfig) Save() error {
	data, err := json.MarshalIndent(m, "", "    ")
	if err != nil {
		return err
	}

	configDir, err := configDirFunc()
	if err != nil {
		return err
	}

	path := filepath.Join(configDir, "mail.json")
	return os.WriteFile(path, data, 0600)
}

func (m *MailConfig) Load() error {
	configDir, err := configDirFunc()
	if err != nil {
		return err
	}

	path := filepath.Join(configDir, "mail.json")
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return fmt.Errorf("No mail config found. You need to auth to fill it, call \"pomo auth --mail\"")
	}

	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

	return nil
}
