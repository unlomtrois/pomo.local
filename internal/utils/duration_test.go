package utils

import (
	"testing"
	"time"
)

func TestShortDuration(t *testing.T) {
	tests := []struct {
		name string
		d    time.Duration
		want string
	}{
		{"Seconds only", 30 * time.Second, "30s"},
		{"Minutes only", 2 * time.Minute, "2m"},
		{"Hours only", 1 * time.Hour, "1h"},
		{"Mixed units", 1*time.Hour + 30*time.Minute, "1h30m"},
		{"With seconds", 1*time.Minute + 5*time.Second, "1m5s"},
		{"Zero duration", 0, "0s"},
		{"Negative duration", -1 * time.Hour, "-1h"},
		{"Sub-second", 500 * time.Millisecond, "500ms"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ShortDuration(tt.d); got != tt.want {
				t.Errorf("ShortDuration(%v) = %v, want %v", tt.d, got, tt.want)
			}
		})
	}
}
