package scheduler

import (
	"fmt"

	"pomo.local/internal/utils"
)

type Scheduler interface {
	Schedule(task Task) error
	Cancel(taskID string) error
}

func NewDefault() (Scheduler, error) {
	if utils.HasSystemd() {
		return &SystemdScheduler{}, nil
	}

	if utils.HasAt() {
		return &AtScheduler{}, nil
	}

	return nil, fmt.Errorf("neither 'systemd-run' nor 'at' found. Please install one of them for background notifications")
}
