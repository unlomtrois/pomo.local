package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"os/exec"
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

func NewPomodoro(title string, message string, startTime time.Time, stopTime time.Time, duration time.Duration) *Pomodoro {
	return &Pomodoro{
		Title:     title,
		Message:   message,
		StartTime: startTime,
		StopTime:  stopTime,
		Duration:  duration,
	}
}

func (p *Pomodoro) Strings() []string {
	startTime := p.StartTime.Format(time.DateTime)
	stopTime := p.StopTime.Format(time.DateTime)
	duration := formatDuration(p.Duration)
	return []string{p.Title, startTime, stopTime, duration}
}

func (p *Pomodoro) Notify() error {
	notifyCmd := fmt.Sprintf("DISPLAY=:0 notify-send -u critical '%s' '%s'", p.Title, p.Message)
	atTime := fmt.Sprintf("now + %d minutes", int(p.Duration.Minutes()))
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

func initCsv(filename string) error {
	if _, err := os.Stat(filename); os.IsExist(err) {
		return nil
	}

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("Error creating pomodoro.csv: %v\n", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	if err := writer.Write([]string{"title", "start_time", "stop_time", "duration"}); err != nil {
		return fmt.Errorf("Error writing to pomodoro.csv: %v\n", err)
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return fmt.Errorf("Error writing to pomodoro.csv: %v\n", err)
	}

	fmt.Println("Pomodoro csv initialized")

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

func main() {
	newCmd := flag.NewFlagSet("new", flag.ExitOnError)
	restCmd := flag.NewFlagSet("rest", flag.ExitOnError)

	var duration int
	var title string
	var message string
	var noNotify bool

	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: pomodoro <command> [options]")
		fmt.Fprintln(os.Stderr, "Commands:")
		fmt.Fprintln(os.Stderr, "  new - Set a new pomodoro timer")
		fmt.Fprintln(os.Stderr, "  rest - Set a rest timer")
		fmt.Fprintln(os.Stderr, "Options:")
		fmt.Fprintln(os.Stderr, "  -d duration - Duration in minutes (default: 25 for new, 5 for rest)")
		fmt.Fprintln(os.Stderr, "  -m message - Notification message (default: Pomodoro finished! Time for a break for new, Break finished! Time for a pomodoro for rest)")
		fmt.Fprintln(os.Stderr, "  -t title - Notification title (default: Pomodoro Timer for new, Break Timer for rest)")
		fmt.Fprintln(os.Stderr, "  -n - Don't notify")
	}

	if len(os.Args) < 2 {
		flag.Usage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "new":
		newCmd.IntVar(&duration, "d", 25, "Duration in minutes (default: 25)")
		newCmd.StringVar(&title, "t", "Pomodoro Timer", "Notification title")
		newCmd.StringVar(&message, "m", "Pomodoro finished! Time for a break.", "Notification message")
		newCmd.BoolVar(&noNotify, "n", false, "Don't notify")
		newCmd.Parse(os.Args[2:])
	case "rest":
		restCmd.IntVar(&duration, "d", 5, "Duration in minutes (default: 5)")
		restCmd.StringVar(&title, "t", "Break Timer", "Notification title")
		restCmd.StringVar(&message, "m", "Break finished! Time for a pomodoro.", "Notification message")
		restCmd.BoolVar(&noNotify, "n", false, "Don't notify")
		restCmd.Parse(os.Args[2:])
	default:
		flag.Usage()
		os.Exit(1)
	}

	if duration <= 0 {
		fmt.Fprintln(os.Stderr, "Error: duration must be positive")
		os.Exit(1)
	}

	// Construct pomodoro entry

	startTime := time.Now()
	stopTime := startTime.Add(time.Duration(duration) * time.Minute)
	safeTitle := strings.ReplaceAll(title, "'", "'\"'\"'")
	safeMessage := strings.ReplaceAll(message, "'", "'\"'\"'")

	pomodoro := NewPomodoro(safeTitle, safeMessage, startTime, stopTime, time.Duration(duration)*time.Minute)

	if err := initCsv("pomodoro.csv"); err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing pomodoro.csv: %v\n", err)
		os.Exit(1)
	}

	if err := pomodoro.Save("pomodoro.csv"); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving pomodoro: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("🍅 Pomodoro timer set for %d minutes\n", duration)

	if noNotify {
		os.Exit(0)
	}

	if err := pomodoro.Notify(); err != nil {
		fmt.Fprintf(os.Stderr, "Error notifying: %v\n", err)
		os.Exit(1)
	}
}
