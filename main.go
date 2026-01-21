package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

func main() {
	duration := flag.Int("d", 25, "Duration in minutes (default: 25)")
	message := flag.String("m", "Pomodoro finished! Time for a break.", "Notification message")
	title := flag.String("t", "Pomodoro Timer", "Notification title")
	flag.Parse()

	if *duration <= 0 {
		fmt.Fprintln(os.Stderr, "Error: duration must be positive")
		os.Exit(1)
	}

	// Calculate the time when the timer should fire
	endTime := time.Now().Add(time.Duration(*duration) * time.Minute)

	// Build the notify-send command
	// Escape single quotes in title and message
	safeTitle := strings.ReplaceAll(*title, "'", "'\"'\"'")
	safeMessage := strings.ReplaceAll(*message, "'", "'\"'\"'")
	notifyCmd := fmt.Sprintf("DISPLAY=:0 notify-send -u critical '%s' '%s'", safeTitle, safeMessage)

	// Create the at command with "now + X minutes"
	atTime := fmt.Sprintf("now + %d minutes", *duration)

	cmd := exec.Command("at", atTime)

	// Pipe the notify-send command to at via stdin
	stdin, err := cmd.StdinPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating stdin pipe: %v\n", err)
		os.Exit(1)
	}

	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting at command: %v\n", err)
		os.Exit(1)
	}

	_, err = stdin.Write([]byte(notifyCmd + "\n"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to stdin: %v\n", err)
		os.Exit(1)
	}
	stdin.Close()

	if err := cmd.Wait(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\nMake sure 'at' daemon (atd) is running.\n", err)
		os.Exit(1)
	}

	fmt.Printf("🍅 Pomodoro timer set for %d minutes\n", *duration)
	fmt.Printf("   You'll be notified at %s\n", endTime.Format("15:04:05"))
}
