package pomo

import (
	"fmt"

	"strings"
	"time"
)

func formatDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	return fmt.Sprintf("%02d:%02d", h, m)
}

type Pomodoro struct { // This thing is saved to csv / database / toggl integration
	Title     string        `csv:"title"`
	Message   string        `csv:"message"`
	StartTime time.Time     `csv:"start_time"`
	StopTime  time.Time     `csv:"stop_time"`
	Duration  time.Duration `csv:"duration"`
}

func NewPomodoro(title string, message string, duration time.Duration) *Pomodoro {
	startTime := time.Now()
	stopTime := startTime.Add(duration)
	safeTitle := strings.ReplaceAll(title, "'", "'\"'\"'")
	safeMessage := strings.ReplaceAll(message, "'", "'\"'\"'")

	return &Pomodoro{
		Title:     safeTitle,
		Message:   safeMessage,
		StartTime: startTime,
		StopTime:  stopTime,
		Duration:  duration,
	}
}

func (p *Pomodoro) Strings() []string {
	startTime := p.StartTime.Format(time.RFC3339) // in utc
	stopTime := p.StopTime.Format(time.RFC3339)   // in utc
	duration := formatDuration(p.Duration)
	return []string{p.Title, startTime, stopTime, duration}
}
