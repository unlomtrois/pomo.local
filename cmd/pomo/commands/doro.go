package commands

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/spf13/cobra"
	"pomo.local/internal/client"
	"pomo.local/internal/project"
	"pomo.local/internal/utils"
)

var doroCmd = &cobra.Command{
	Use:   "doro [topic]",
	Short: "Start a pomodoro session (fixed timer)",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runDoro,
}

func init() {
	rootCmd.AddCommand(doroCmd)

	doroCmd.Flags().StringP("message", "m", "Pomodoro session is ended!", "Notification message")
	doroCmd.Flags().DurationP("duration", "d", 25*time.Minute, "Timer duration")
	doroCmd.Flags().String("hint", utils.HintDefault, "Hint the same as notify-send hint")
	doroCmd.Flags().Bool("email", false, "Send email when the session is over")
	doroCmd.Flags().BoolP("verbose", "v", false, "Verbose output")
}

func runDoro(cmd *cobra.Command, args []string) error {
	verbose, _ := cmd.Flags().GetBool("verbose")
	if verbose {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	var topic string
	if len(args) == 1 {
		topic = args[0]
	}
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

	c, err := ensureDaemon(ctx, client.DefaultAddr())
	if err != nil {
		return err
	}

	params := client.StartParams{
		Topic:    topic,
		Duration: duration.String(),
		Source:   "cli",
		Message:  message,
		Hint:     hint,
		Email:    useEmail,
	}
	// Tag the session with the nearest .pomo project, if we're inside one.
	proj := discoverProject()
	if proj != nil {
		params.Project = proj.ID
		params.ProjectName = proj.Name
	}

	sess, err := c.StartSession(ctx, params)
	if errors.Is(err, client.ErrActiveExists) {
		return fmt.Errorf("you can only have 1 active pomodoro session at once")
	}
	if err != nil {
		return err
	}

	fmt.Printf("[%s] Started %q", short(sess.Hash), sess.Topic)
	if proj != nil {
		fmt.Printf(" in project %q", proj.Name)
	}
	fmt.Printf(" — you'll be notified at %s", sess.StopTime.Local().Format("15:04:05"))
	if useEmail {
		fmt.Print(", and via email")
	}
	fmt.Println()
	return nil
}

// discoverProject finds the nearest .pomo project from the working directory,
// or returns nil if we're not inside one.
func discoverProject() *project.Project {
	cwd, err := os.Getwd()
	if err != nil {
		return nil
	}
	p, err := project.Find(cwd)
	if err != nil {
		return nil
	}
	return p
}
