package cli

import (
	"flag"

	"pomo.local/internal/notifier"
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
	fs := flag.NewFlagSet("notify", flag.ExitOnError)
	fs.StringVar(&cmd.title, "t", "Pomodoro session is ended", "Title")
	fs.StringVar(&cmd.message, "m", "", "Notification message")
	fs.StringVar(&cmd.hint, "hint", utils.HintDefault, "Hint the same as notify-send hint")
	fs.Parse(args)
	return &cmd
}

func (cmd *NotifyCommand) Run() error {
	notifier := &notifier.LibnotifyNotifier{}
	return notifier.Notify(cmd.title, cmd.message, cmd.hint)
}
