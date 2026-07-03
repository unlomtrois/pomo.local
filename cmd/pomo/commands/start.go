package commands

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/spf13/cobra"
	"pomo.local/internal/client"
	"pomo.local/internal/utils"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Set a new pomodoro timer",
	RunE:  runStart,
}

func init() {
	rootCmd.AddCommand(startCmd)

	startCmd.Flags().StringP("topic", "t", "", "Topic of your pomodoro session")
	startCmd.Flags().StringP("message", "m", "Pomodoro session is ended!", "Notification message")
	startCmd.Flags().DurationP("duration", "d", 25*time.Minute, "Timer duration")
	startCmd.Flags().String("hint", utils.HintDefault, "Hint the same as notify-send hint")
	startCmd.Flags().Bool("email", false, "Send email when the session is over")
	startCmd.Flags().BoolP("verbose", "v", false, "Verbose output")
}

func runStart(cmd *cobra.Command, _ []string) error {
	verbose, _ := cmd.Flags().GetBool("verbose")
	if verbose {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	topic, _ := cmd.Flags().GetString("topic")
	message, _ := cmd.Flags().GetString("message")
	duration, _ := cmd.Flags().GetDuration("duration")
	hint, _ := cmd.Flags().GetString("hint")
	useEmail, _ := cmd.Flags().GetBool("email")

	return executeStart(topic, message, duration, hint, useEmail)
}

// executeStart starts a session through the daemon, auto-spawning it if needed.
// It is shared with the `rest` command.
func executeStart(topic, message string, duration time.Duration, hint string, useEmail bool) error {
	ctx := context.Background()

	c, err := ensureDaemon(ctx, client.DefaultAddr)
	if err != nil {
		return err
	}

	sess, err := c.StartSession(ctx, client.StartParams{
		Topic:    topic,
		Duration: duration.String(),
		Source:   "cli",
		Message:  message,
		Hint:     hint,
		Email:    useEmail,
	})
	if errors.Is(err, client.ErrActiveExists) {
		return fmt.Errorf("you can only have 1 active pomodoro session at once")
	}
	if err != nil {
		return err
	}

	fmt.Printf("Started %q — you'll be notified at %s",
		sess.Topic, sess.StopTime.Local().Format("15:04:05"))
	if useEmail {
		fmt.Print(", and via email")
	}
	fmt.Println()
	return nil
}
