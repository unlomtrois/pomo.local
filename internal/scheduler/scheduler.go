package scheduler

import (
	"fmt"
	"os/exec"
)

type Scheduler interface {
	Schedule(task Task) error
	Cancel(taskID string) error
}

func hasSystemd() bool {
	_, err := exec.LookPath("systemd-run")
	return err == nil
}

func hasAt() bool {
	_, err := exec.LookPath("at")
	return err == nil
}

func NewDefault() (Scheduler, error) {
	if hasSystemd() {
		return &SystemdScheduler{}, nil
	}

	if hasAt() {
		return &AtScheduler{}, nil
	}

	return nil, fmt.Errorf("neither 'systemd-run' nor 'at' found. Please install one of them for background notifications")
}
