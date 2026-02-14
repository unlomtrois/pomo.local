package pomo

import (
	"encoding/csv"
	"fmt"
	"os"

	"strings"
	"time"

	"pomo.local/internal/toggl"
	"pomo.local/internal/utils"
)

var (
	HintMute    = "boolean:suppress-sound:true"
	HintDefault = "string:sound-name:complete"
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

func (p *Pomodoro) NotifyLater(hint string) error {
	var notifier Notifier

	if utils.HasSystemd() {
		notifier = &SystemdNotifier{}
	} else if utils.HasAt() {
		notifier = &AtNotifier{}
	} else {
		return fmt.Errorf("neither 'systemd-run' nor 'at' found. Please install one of them for background notifications")
	}

	if err := notifier.NotifyLater(p, hint); err != nil {
		return err
	}

	return nil
}

func (p *Pomodoro) Save(filename string) error {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("Error opening pomodoro.csv: %v\n", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	if err := writer.Write(p.Strings()); err != nil {
		return fmt.Errorf("Error writing to pomodoro.csv: %v\n", err)
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return fmt.Errorf("Error writing to pomodoro.csv: %v\n", err)
	}

	fmt.Println("Pomodoro added to csv")

	return nil
}

func (p *Pomodoro) SaveInToggl(token string, workspaceId int, userId int) error {
	entry := toggl.NewTogglEntry(p.Title, p.StartTime, p.StopTime, userId, workspaceId)
	if err := entry.Save(token, workspaceId); err != nil {
		return fmt.Errorf("Error saving entry: %v", err)
	}
	return nil
}
