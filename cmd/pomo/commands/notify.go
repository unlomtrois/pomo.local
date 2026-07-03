package commands

import (
	"log/slog"

	"github.com/spf13/cobra"
	"pomo.local/internal/mail"
	"pomo.local/internal/notifier"
	"pomo.local/internal/utils"
)

// notifyCmd sends an immediate notification. In the daemon model the daemon
// fires session-completion notifications in-process, so this is now a manual
// utility (e.g. testing notify-send/SMTP) rather than part of the timer flow.
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

func runNotify(cmd *cobra.Command, _ []string) error {
	summary, _ := cmd.Flags().GetString("summary")
	body, _ := cmd.Flags().GetString("body")
	hint, _ := cmd.Flags().GetString("hint")
	useEmail, _ := cmd.Flags().GetBool("email")

	n := &notifier.LibnotifyNotifier{}
	if err := n.Notify(summary, body, hint); err != nil {
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
