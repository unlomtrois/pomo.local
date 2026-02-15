package cli

import (
	"flag"
	"fmt"

	"github.com/zalando/go-keyring"
	"pomo.local/internal/config"
	"pomo.local/internal/mail"
	"pomo.local/internal/notifier"
	"pomo.local/internal/utils"
)

// NotifyCommand is basically a wrapper around notify-send
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
	fs.BoolVar(&cmd.useEmail, "mail", false, "Send also email")
	fs.Parse(args)
	return &cmd
}

func (cmd *NotifyCommand) Run() error {
	notifier := &notifier.LibnotifyNotifier{}
	if err := notifier.Notify(cmd.summary, cmd.body, cmd.hint); err != nil {
		return err
	}

	if cmd.useEmail {
		fmt.Println("sending email...")
		cfg := &config.MailConfig{}
		if err := cfg.Load(); err != nil {
			return err
		}

		if err := cfg.Validate(); err != nil {
			return fmt.Errorf("Invalid config (try \"pomo auth --mail\" again): %v", err)
		}

		pass, err := keyring.Get("pomo-smtp", cfg.Sender)
		if err != nil {
			return fmt.Errorf("Failed to get keyring for sender: %v, %w. (try to \"pomo auth --mail\" again)", cfg.Sender, err)
		}

		fmt.Println("send email to", cfg.Receiver)
		err = mail.Send(cfg, pass, "Pomo:"+cmd.summary, cmd.body)
		if err != nil {
			return fmt.Errorf("Failed to send an email: %v", err)
		}

		fmt.Println("Mail sended!")
	}

	return nil
}
