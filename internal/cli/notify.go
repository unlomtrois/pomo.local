package cli

import (
	"flag"
	"log/slog"
	"os"

	"github.com/adrg/xdg"
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

	slog.Debug("Removing active_session")
	if err := removeActiveSession(); err != nil {
		return err
	}

	slog.Debug("Removing active_task")
	if err := removeActiveTask(); err != nil {
		return err
	}

	if cmd.useEmail {
		slog.Debug("Sending mail")
		if err := mail.SendMail(cmd.summary, cmd.body); err != nil {
			return err
		}
	}

	return nil
}

func removeActiveSession() error {
	activeSessionPath, err := xdg.StateFile("pomo/active_session.json")
	if err != nil {
		return nil
	}

	slog.Debug("Read active_session:", "path", activeSessionPath)
	if _, err := os.Stat(activeSessionPath); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		slog.Debug("active_session not found:", "path", activeSessionPath)
	}

	if err := os.Remove(activeSessionPath); err != nil {
		return err
	}

	return nil
}

func removeActiveTask() error {
	activeTaskPath, err := xdg.StateFile("pomo/active_task.json")
	if err != nil {
		return nil
	}

	slog.Debug("Read active_task:", "path", activeTaskPath)
	if _, err := os.Stat(activeTaskPath); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		slog.Debug("active_task not found:", "path", activeTaskPath)
	}

	if err := os.Remove(activeTaskPath); err != nil {
		return err
	}

	return nil
}
