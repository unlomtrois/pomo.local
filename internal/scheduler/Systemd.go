package scheduler

import (
	"fmt"
	"os/exec"
	"time"
)

type SystemdScheduler struct{}

func (sd *SystemdScheduler) Schedule(task Task) error {
	delay := max(time.Until(task.ExecuteAt), 0)
	delayStr := fmt.Sprintf("%.3fs", delay.Seconds())
	unitName := fmt.Sprintf("pomo-notify-%s", task.ID)

	// pomoPath, err := exec.LookPath("pomo")
	// if err != nil {
	// 	return fmt.Errorf("could not find pomo executable: %v", err)
	// }
	//
	args := []string{
		"--user", "--collect", "--unit=" + unitName, "--on-active=" + delayStr, "--timer-property=AccuracySec=100ms",
		task.Binary,
	}

	args = append(args, task.Args...)

	cmd := exec.Command("systemd-run", args...)

	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("Failed to schedule notification: %v: %s", err, out)
	}

	fmt.Printf("You'll be notified at %s (via systemd timer %s)\n", task.ExecuteAt.Format("15:04:05"), unitName)
	return nil
}

func (sd *SystemdScheduler) Cancel(taskID string) error {
	panic("todo")
}
