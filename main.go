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
	Title     string `csv:"title"`
	StartTime string `csv:"start_time"`
	StopTime  string `csv:"stop_time"`
	Duration  string `csv:"duration"`
}

func NewPomodoro(title string, startTime time.Time, stopTime time.Time, duration time.Duration) *Pomodoro {
	return &Pomodoro{
		Title:     title,
		StartTime: startTime.Format(time.DateTime),
		StopTime:  stopTime.Format(time.DateTime),
		Duration:  formatDuration(duration),
	}
}

func main() {
	newCmd := flag.NewFlagSet("new", flag.ExitOnError)
	restCmd := flag.NewFlagSet("rest", flag.ExitOnError)

	var duration int
	var title string
	var message string
	var noNotify bool

	newCmd.IntVar(&duration, "d", 25, "Duration in minutes (default: 25)")
	newCmd.StringVar(&title, "t", "Pomodoro Timer", "Notification title")
	newCmd.StringVar(&message, "m", "Pomodoro finished! Time for a break.", "Notification message")
	newCmd.BoolVar(&noNotify, "n", false, "Don't notify")

	restCmd.IntVar(&duration, "d", 5, "Duration in minutes (default: 5)")
	restCmd.StringVar(&title, "t", "Break Timer", "Notification title")
	restCmd.StringVar(&message, "m", "Break finished! Time for a pomodoro.", "Notification message")
	restCmd.BoolVar(&noNotify, "n", false, "Don't notify")

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
		newCmd.Parse(os.Args[2:])
	case "rest":
		restCmd.Parse(os.Args[2:])
	default:
		flag.Usage()
		os.Exit(1)
	}

	if duration <= 0 {
		fmt.Fprintln(os.Stderr, "Error: duration must be positive")
		os.Exit(1)
	}

	// Calculate the time when the timer should fire

	startTime := time.Now()
	stopTime := startTime.Add(time.Duration(duration) * time.Minute)

	// Build the notify-send command
	// Escape single quotes in title and message
	safeTitle := strings.ReplaceAll(title, "'", "'\"'\"'")
	safeMessage := strings.ReplaceAll(message, "'", "'\"'\"'")
	notifyCmd := fmt.Sprintf("DISPLAY=:0 notify-send -u critical '%s' '%s'", safeTitle, safeMessage)

	// pomodoro

	pomodoro := NewPomodoro(safeTitle, startTime, stopTime, time.Duration(duration)*time.Minute)

	file, err := os.OpenFile("pomodoro.csv", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating pomodoro.csv: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	writer := csv.NewWriter(file)

	marchalledPomodoro := []string{pomodoro.Title, pomodoro.StartTime, pomodoro.StopTime, pomodoro.Duration}
	if err := writer.Write(marchalledPomodoro); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to pomodoro.csv: %v\n", err)
		os.Exit(1)
	} else {
		fmt.Println("Pomodoro added to csv")
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		// Handle late-occurring write errors
		fmt.Fprintf(os.Stderr, "Error writing to pomodoro.csv: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("🍅 Pomodoro timer set for %d minutes\n", duration)

	if noNotify {
		os.Exit(0)
	}

	// Create the at command with "now + X minutes"
	atTime := fmt.Sprintf("now + %d minutes", duration)
	atCmd := exec.Command("at", atTime)

	// Pipe the notify-send command to at via stdin
	stdin, err := atCmd.StdinPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating stdin pipe: %v\n", err)
		os.Exit(1)
	}

	if err := atCmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting at command: %v\n", err)
		os.Exit(1)
	}

	_, err = stdin.Write([]byte(notifyCmd + "\n"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to stdin: %v\n", err)
		os.Exit(1)
	}
	stdin.Close()

	if err := atCmd.Wait(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\nMake sure 'at' daemon (atd) is running.\n", err)
		os.Exit(1)
	}

	fmt.Printf("   You'll be notified at %s\n", stopTime.Format("15:04:05"))
}
