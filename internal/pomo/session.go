package pomo

import (
	"fmt"

	"strings"
	"time"
)

type Session struct { // This thing is saved to csv / database / toggl integration
	Topic     string        `json:"topic"`
	StartTime time.Time     `json:"start_time"`
	StopTime  time.Time     `json:"stop_time"`
	Duration  time.Duration `json:"duration"`
}

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

func (p *Session) Strings() []string {
	startTime := p.StartTime.Format(time.RFC3339) // in utc
	stopTime := p.StopTime.Format(time.RFC3339)   // in utc
	duration := formatDuration(p.Duration)
	return []string{p.Topic, startTime, stopTime, duration}
}
