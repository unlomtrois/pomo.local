package scheduler

import (
	"fmt"
	"os/exec"
	"slices"
	"strings"
	"time"
)

type AtScheduler struct {
	verbose bool
}

func (_ *AtScheduler) Schedule(task Task) error {
	if time.Until(task.ExecuteAt) < 1*time.Minute {
		fmt.Println("Warning: \"at\" does not support sub-minute precision")
	}

	atTime := fmt.Sprintf("now + %d minutes", int(time.Until(task.ExecuteAt).Minutes()))
	cmd := exec.Command("at", atTime)

	// Pipe the notify-send command to at via stdin
	args := []string{task.Binary}
	args = append(args, task.Args...)
	notifyCmd := strings.Join(args, " ")

	sleepAndNotify := fmt.Sprintf("sleep %d && %s", task.ExecuteAt.Second(), notifyCmd)
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

	var sb strings.Builder
	fmt.Fprintf(&sb, "You'll be notified at %s (via at", task.ExecuteAt.Format("15:04:05"))
	if slices.Contains(args, "--email") {
		sb.WriteString(", and via email")
	}
	sb.WriteString(")\n")
	fmt.Print(sb.String())
	return nil
}

func (_ *AtScheduler) Cancel(taskID string) error {
	panic("todo")
}
