package scheduler

import (
	"fmt"
	"log/slog"
	"os/exec"
	"slices"
	"strings"
	"time"
)

type SystemdScheduler struct {
	verbose bool
}

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

	taskArgs := []string{task.Binary}
	taskArgs = append(taskArgs, task.Args...)
	slog.Debug("Task:", "cmd", strings.Join(taskArgs, " "))

	args = append(args, taskArgs...)
	cmd := exec.Command("systemd-run", args...)
	slog.Debug("Scheduling command:", "cmd", cmd.String())

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Failed to schedule notification: %v: %s", err, out)
	}
	slog.Debug("systemd-run's output:", "out", out)

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
