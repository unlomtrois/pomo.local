package cli

import (
	"flag"
	"fmt"
	"os/exec"

	"pomo.local/internal/utils"
)

// NotifyCommand is basically a wrapper around notify-send
type NotifyCommand struct {
	title   string
	message string
	hint    string
}

func ParseNotify(args []string) *NotifyCommand {
	cmd := NotifyCommand{}
	notifyCmd := flag.NewFlagSet("notify", flag.ExitOnError)
	notifyCmd.StringVar(&cmd.title, "t", "Pomodoro session is ended", "Title")
	notifyCmd.StringVar(&cmd.message, "m", "", "Notification message")
	notifyCmd.StringVar(&cmd.hint, "hint", utils.HintDefault, "Hint the same as notify-send hint")
	notifyCmd.Parse(args)
	return &cmd
}

func (cmd *NotifyCommand) Run() error {
	notify := exec.Command("notify-send", cmd.title, cmd.message, "--hint", cmd.hint)

	if notify.Err != nil {
		return fmt.Errorf("Error finding notify-send: %v\n", notify.Err)
	}

	if err := notify.Run(); err != nil {
		return fmt.Errorf("Error running notify: %v\n", err)
	}

	return nil
}
