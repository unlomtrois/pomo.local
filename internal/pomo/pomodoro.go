package pomo

import (
	"encoding/csv"
	"fmt"
	"os"
	"os/exec"

	"strings"
	"time"

	"pomo.local/internal/toggl"
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

var DefaultHint = "string:sound-name:complete" // XDG Sound Theme spec

func (p *Pomodoro) Notify(muteNotifySound bool) error {
	var Hint = DefaultHint
	if muteNotifySound {
		Hint = "boolean:suppress-sound:true"
	}

	notifyCmd := fmt.Sprintf("DISPLAY=:0 notify-send -u critical '%s' '%s' --hint=\"%s\"", p.Title, p.Message, Hint)
	atTime := fmt.Sprintf("now + %d minutes", int(p.Duration.Minutes())) // todo: perhaps this blocks me from setting for 10 seconds, and runs automatically
	atCmd := exec.Command("at", atTime)

	// Pipe the notify-send command to at via stdin
	stdin, err := atCmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("Error creating stdin pipe: %v\n", err)
	}

	if err := atCmd.Start(); err != nil {
		return fmt.Errorf("Error starting at command: %v\n", err)
	}

	_, err = stdin.Write([]byte(notifyCmd + "\n"))
	if err != nil {
		return fmt.Errorf("Error writing to stdin: %v\n", err)
	}
	if err := stdin.Close(); err != nil {
		return fmt.Errorf("Error closing stdin: %v\n", err)
	}

	if err := atCmd.Wait(); err != nil {
		return fmt.Errorf("Error: %v\nMake sure 'at' daemon (atd) is running.\n", err)
	}

	fmt.Printf("   You'll be notified at %s\n", p.StopTime.Format("15:04:05"))
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
