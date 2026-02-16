package scheduler

import (
	"fmt"
	"os/exec"
	"slices"
	"strings"
	"time"
)

type SystemdScheduler struct{}

func (sd *SystemdScheduler) Schedule(task Task) error {
	delay := max(time.Until(task.ExecuteAt), 0)
	delayStr := fmt.Sprintf("%.3fs", delay.Seconds())
	unitName := fmt.Sprintf("pomo-notify-%s", task.ID)

	args := []string{
		"--user", "--collect",
		"--unit=" + unitName,
		"--on-active=" + delayStr,
		"--timer-property=AccuracySec=100ms",
	}
	args = append(args, task.Binary)
	args = append(args, task.Args...)

	cmd := exec.Command("systemd-run", args...)

	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("Failed to schedule notification: %v: %s", err, out)
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "You'll be notified at %s (via systemd timer %s", task.ExecuteAt.Format("15:04:05"), unitName)
	if slices.Contains(args, "--email") {
		sb.WriteString(", and via email")
	}
	sb.WriteString(")\n")
	fmt.Print(sb.String())

	return nil
}

func (sd *SystemdScheduler) Cancel(taskID string) error {
	panic("todo")
}
