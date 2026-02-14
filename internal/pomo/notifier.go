package pomo

import (
	"fmt"
	"os/exec"
	"time"
)

type Notifier interface {
	NotifyLater(p *Pomodoro, hint string) error
}

type SystemdNotifier struct{}

func (sd *SystemdNotifier) NotifyLater(p *Pomodoro, hint string) error {
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

type AtNotifier struct{}

func (n *AtNotifier) NotifyLater(p *Pomodoro, hint string) error {
	if p.Duration < 1*time.Minute {
		fmt.Println("Warning: \"at\" does not support sub-minute precision")
	}

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
