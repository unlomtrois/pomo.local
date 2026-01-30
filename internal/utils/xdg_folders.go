package utils

import (
	"os"
	"path/filepath"
)

func GetConfigDir() (string, error) {
	// Respect XDG_CONFIG_HOME if set
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return filepath.Join(dir, "pomo"), nil
	}
	// Default to ~/.config/pomo
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "pomo"), nil
}

func GetDataDir() (string, error) {
	// Respect XDG_DATA_HOME if set
	if dir := os.Getenv("XDG_DATA_HOME"); dir != "" {
		return filepath.Join(dir, "pomo"), nil
	}
	// Default to ~/.local/share/pomo
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share", "pomo"), nil
}
