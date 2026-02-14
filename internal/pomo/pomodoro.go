package pomo

import (
	"encoding/csv"
	"fmt"
	"os"
	"os/exec"

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

func (p *Pomodoro) notifyViaAt(hint string) error {
	atTime := fmt.Sprintf("now + %d minutes", int(p.Duration.Minutes()))
	cmd := exec.Command("at", atTime)

	// Pipe the notify-send command to at via stdin
	notifyCmd := fmt.Sprintf("pomo notify -t %q -m %q --hint=%q", p.Title, p.Message, hint)
	sleepAndNotify := fmt.Sprintf("sleep %d && %s", p.StopTime.Second(), notifyCmd)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("Error creating stdin pipe: %v\n", err)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("Error starting at command: %v\n", err)
	}
	if _, err = stdin.Write([]byte(sleepAndNotify + "\n")); err != nil {
		return fmt.Errorf("Error writing to stdin: %v\n", err)
	}
	if err := stdin.Close(); err != nil {
		return fmt.Errorf("Error closing stdin: %v\n", err)
	}
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("Error: %v\nMake sure 'at' daemon (atd) is running.\n", err)
	}

	fmt.Printf("You'll be notified at %s\n", p.StopTime.Format("15:04:05"))
	return nil
}

func (p *Pomodoro) notifyViaSystemd(hint string) error {
	delay := max(time.Until(p.StopTime), 0)
	delayStr := fmt.Sprintf("%.3fs", delay.Seconds())
	unitName := fmt.Sprintf("pomo-notify-%d", time.Now().UnixNano())

	pomoPath, err := exec.LookPath("pomo")
	if err != nil {
		return fmt.Errorf("could not find pomo executable: %v", err)
	}

	cmd := exec.Command("systemd-run",
		"--user", "--collect", "--unit="+unitName, "--on-active="+delayStr, "--timer-property=AccuracySec=100ms",
		pomoPath, "notify", "-t="+p.Title, "-m="+p.Message, "--hint="+hint,
	)

	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("Failed to schedule notification: %v: %s", err, out)
	}

	fmt.Printf("You'll be notified at %s (via systemd timer %s)\n", p.StopTime.Format("15:04:05"), unitName)
	return nil
}

func (p *Pomodoro) NotifyLater(hint string) error {
	if utils.HasSystemd() {
		return p.notifyViaSystemd(hint)
	}

	if utils.HasAt() {
		if p.Duration < 1*time.Minute {
			fmt.Println("Warning: \"at\" does not support sub-minute precision")
		}
		return p.notifyViaAt(hint)
	}

	return fmt.Errorf("neither 'systemd-run' nor 'at' found. Please install one of them for background notifications")
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
