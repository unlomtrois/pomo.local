package commands

import (
	"log/slog"
	"os"

	"github.com/adrg/xdg"
	"github.com/spf13/cobra"
	"pomo.local/internal/mail"
	"pomo.local/internal/notifier"
	"pomo.local/internal/utils"
)

var notifyCmd = &cobra.Command{
	Use:   "notify",
	Short: "Send an immediate notification",
	RunE:  runNotify,
}

func init() {
	rootCmd.AddCommand(notifyCmd)

	notifyCmd.Flags().String("summary", "Pomodoro session is ended", "Notification title")
	notifyCmd.Flags().String("body", "", "Notification message")
	notifyCmd.Flags().String("hint", utils.HintDefault, "Hint the same as notify-send hint")
	notifyCmd.Flags().Bool("email", false, "Send also email")
}

func runNotify(cmd *cobra.Command, args []string) error {
	summary, _ := cmd.Flags().GetString("summary")
	body, _ := cmd.Flags().GetString("body")
	hint, _ := cmd.Flags().GetString("hint")
	useEmail, _ := cmd.Flags().GetBool("email")

	n := &notifier.LibnotifyNotifier{}
	if err := n.Notify(summary, body, hint); err != nil {
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

	if useEmail {
		slog.Debug("Sending mail")
		if err := mail.SendMail(summary, body); err != nil {
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
