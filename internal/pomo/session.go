// Package pomo defines the core session type for the pomodoro timer.
package pomo

import (
	"fmt"

	"strings"
	"time"
)

// Session represents a single pomodoro work session saved to CSV, database, and Toggl.
type Session struct {
	Topic     string        `json:"topic"`
	StartTime time.Time     `json:"start_time"`
	StopTime  time.Time     `json:"stop_time"`
	Duration  time.Duration `json:"duration"`
}

// NewSession creates a new Session starting now with the given topic and duration.
func NewSession(topic string, duration time.Duration) *Session {
	startTime := time.Now()
	stopTime := startTime.Add(duration)
	safeTopic := strings.ReplaceAll(topic, "'", "'\"'\"'")

	return &Session{
		Topic:     safeTopic,
		StartTime: startTime,
		StopTime:  stopTime,
		Duration:  duration,
	}
}

func formatDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	return fmt.Sprintf("%02d:%02d", h, m)
}

// Strings returns the session fields as a slice for CSV serialization.
func (p *Session) Strings() []string {
	startTime := p.StartTime.Format(time.RFC3339) // in utc
	stopTime := p.StopTime.Format(time.RFC3339)   // in utc
	duration := formatDuration(p.Duration)
	return []string{p.Topic, startTime, stopTime, duration}
}
