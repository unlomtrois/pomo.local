package cli

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"pomo.local/internal/pomo"
	"pomo.local/internal/scheduler"
	"pomo.local/internal/utils"
)

// StartCommand is basically a wrapper around notify-send
type StartCommand struct {
	topic    string
	message  string
	duration time.Duration
	hint     string
	useToggl bool
	useEmail bool
}

func ParseStart(args []string) *StartCommand {
	cmd := StartCommand{}
	fs := flag.NewFlagSet("start", flag.ExitOnError)
	fs.StringVar(&cmd.topic, "t", "", "Topic of your pomodoro session")
	fs.StringVar(&cmd.message, "m", "Pomodoro session is ended!", "Notification message")
	fs.DurationVar(&cmd.duration, "d", 25*time.Minute, "Timer duration")
	fs.StringVar(&cmd.hint, "hint", utils.HintDefault, "Hint the same as notify-send hint")
	fs.BoolVar(&cmd.useToggl, "toggl", false, "Use toggl integration?")
	fs.BoolVar(&cmd.useEmail, "email", false, "Send email when the session is over?")
	fs.Parse(args)
	return &cmd
}

func (cmd *StartCommand) Run() error {
	session := pomo.NewSession(cmd.topic, cmd.duration)

	s, err := scheduler.NewDefault()
	if err != nil {
		return err
	}

	bin, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not find pomo executable: %v", err)
	}

	args := []string{
		"notify",
		"--summary", "Pomodoro",
		"--body", cmd.message,
		"--hint", cmd.hint,
	}
	if cmd.useEmail {
		args = append(args, "--email")
	}

	task := scheduler.Task{
		ID:        strconv.FormatInt(time.Now().Unix(), 16),
		ExecuteAt: session.StopTime,
		Binary:    bin,
		Args:      args,
	}

	if err := s.Schedule(task); err != nil {
		return err
	}

	return nil
}
