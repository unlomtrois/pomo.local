package main

import (
	"flag"
	"fmt"
	"os"
	"pomodoro/internal/pomo"

	"strings"
	"time"
)

func main() {
	newCmd := flag.NewFlagSet("new", flag.ExitOnError)
	restCmd := flag.NewFlagSet("rest", flag.ExitOnError)

	var duration int
	var title string
	var message string
	var noNotify bool

	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: pomo <command> [options]")
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

	if err := pomo.InitCsv("pomodoro.csv"); err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing pomodoro.csv: %v\n", err)
		os.Exit(1)
	}

	startTime := time.Now()
	stopTime := startTime.Add(time.Duration(duration) * time.Minute)
	safeTitle := strings.ReplaceAll(title, "'", "'\"'\"'")
	safeMessage := strings.ReplaceAll(message, "'", "'\"'\"'")
	pomodoro := pomo.NewPomodoro(safeTitle, safeMessage, startTime, stopTime, time.Duration(duration)*time.Minute)

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
