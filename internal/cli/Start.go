package cli

import (
	"flag"
	"fmt"
	"os/exec"
	"time"

	"pomo.local/internal/pomo"
	"pomo.local/internal/scheduler"
	"pomo.local/internal/utils"
)

// StartCommand is basically a wrapper around notify-send
type StartCommand struct {
	title    string
	message  string
	duration time.Duration
	hint     string
	useToggl bool
}

func ParseStart(args []string) *StartCommand {
	cmd := StartCommand{}
	fs := flag.NewFlagSet("start", flag.ExitOnError)
	fs.StringVar(&cmd.title, "t", "Pomodoro session is ended", "Title")
	fs.StringVar(&cmd.message, "m", "", "Notification message")
	fs.DurationVar(&cmd.duration, "d", 25*time.Minute, "Timer duration")
	fs.StringVar(&cmd.hint, "hint", utils.HintDefault, "Hint the same as notify-send hint")
	fs.BoolVar(&cmd.useToggl, "toggl", false, "Use toggl integration?")
	fs.Parse(args)
	return &cmd
}

func (cmd *StartCommand) Run() error {
	pomodoro := pomo.NewPomodoro(cmd.title, cmd.message, cmd.duration)

	var notifier scheduler.Scheduler

	if utils.HasSystemd() {
		notifier = &scheduler.SystemdScheduler{}
	} else if utils.HasAt() {
		notifier = &scheduler.AtScheduler{}
	} else {
		return fmt.Errorf("neither 'systemd-run' nor 'at' found. Please install one of them for background notifications")
	}

	pomoPath, err := exec.LookPath("pomo")
	if err != nil {
		return fmt.Errorf("could not find pomo executable: %v", err)
	}

	task := scheduler.Task{
		ID:        "0",
		ExecuteAt: pomodoro.StopTime,
		Binary:    pomoPath,
		Args:      []string{"notify", "-t=" + pomodoro.Title, "-m=" + pomodoro.Message, "--hint=" + cmd.hint},
	}

	if err := notifier.Schedule(task); err != nil {
		return err
	}

	return nil
}
