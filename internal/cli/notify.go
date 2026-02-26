package cli

import (
	"flag"

	"pomo.local/internal/mail"
	"pomo.local/internal/notifier"
	"pomo.local/internal/utils"
)

type NotifyCommand struct {
	summary  string
	body     string
	hint     string
	useEmail bool
}

func ParseNotify(args []string) *NotifyCommand {
	cmd := NotifyCommand{}
	fs := flag.NewFlagSet("notify", flag.ExitOnError)
	fs.StringVar(&cmd.summary, "summary", "Pomodoro session is ended", "Title")
	fs.StringVar(&cmd.body, "body", "", "Notification message")
	fs.StringVar(&cmd.hint, "hint", utils.HintDefault, "Hint the same as notify-send hint")
	fs.BoolVar(&cmd.useEmail, "email", false, "Send also email")
	fs.Parse(args)
	return &cmd
}

func (cmd *NotifyCommand) Run() error {
	notifier := &notifier.LibnotifyNotifier{}
	if err := notifier.Notify(cmd.summary, cmd.body, cmd.hint); err != nil {
		return err
	}

	if !cmd.useEmail {
		return nil
	}

	if err := mail.SendMail(cmd.summary, cmd.body); err != nil {
		return err
	}

	return nil
}
