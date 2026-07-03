// Package commands implements the pomo CLI subcommands.
package commands

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
	"pomo.local/internal/client"
)

var activeCmd = &cobra.Command{
	Use:   "active",
	Short: "Check or manage the active pomodoro session",
	RunE:  runActive,
}

func init() {
	rootCmd.AddCommand(activeCmd)

	activeCmd.Flags().BoolP("verbose", "v", false, "Verbose output")
	activeCmd.Flags().Bool("remove", false, "Cancel the active session")
}

func runActive(cmd *cobra.Command, _ []string) error {
	verbose, _ := cmd.Flags().GetBool("verbose")
	if verbose {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}
	remove, _ := cmd.Flags().GetBool("remove")

	ctx := context.Background()
	c, err := ensureDaemon(ctx, client.DefaultAddr)
	if err != nil {
		return err
	}

	if remove {
		sess, err := c.StopActive(ctx)
		if errors.Is(err, client.ErrNoActive) {
			fmt.Println("No active pomodoro session")
			return nil
		}
		if err != nil {
			return err
		}
		fmt.Printf("Cancelled active session %q\n", sess.Topic)
		return nil
	}

	sess, err := c.ActiveSession(ctx)
	if errors.Is(err, client.ErrNoActive) {
		fmt.Println("No active pomodoro session")
		return nil
	}
	if err != nil {
		return err
	}

	fmt.Printf("Active session topic: %s, ends at: %s\n",
		sess.Topic, sess.StopTime.Local().Format("15:04:05"))
	return nil
}
