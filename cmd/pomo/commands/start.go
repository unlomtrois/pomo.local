package commands

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"pomo.local/internal/client"
)

var startCmd = &cobra.Command{
	Use:   "start [topic]",
	Short: "Start an open-ended session (stopwatch); end it with `pomo end`",
	Long: "Starts an open-ended session that runs until `pomo end`/`pomo stop`. " +
		"The topic is optional at start — you can name it (or rename it) when you end.",
	Args: cobra.MaximumNArgs(1),
	RunE: runStart,
}

func init() {
	rootCmd.AddCommand(startCmd)
}

func runStart(_ *cobra.Command, args []string) error {
	ctx := context.Background()

	c, err := ensureDaemon(ctx, client.DefaultAddr())
	if err != nil {
		return err
	}

	params := client.StartParams{
		Source: "cli",
		Open:   true,
	}
	if len(args) == 1 {
		params.Topic = args[0]
	}
	if proj := discoverProject(); proj != nil {
		params.Project = proj.ID
		params.ProjectName = proj.Name
	}

	sess, err := c.StartSession(ctx, params)
	if errors.Is(err, client.ErrActiveExists) {
		return fmt.Errorf("you already have an active session (end it with `pomo end`)")
	}
	if err != nil {
		return err
	}

	if sess.Topic != "" {
		fmt.Printf("[%s] Started %q at %s\n", short(sess.Hash), sess.Topic, sess.StartTime.Local().Format("15:04:05"))
	} else {
		fmt.Printf("[%s] Started at %s — name it when you `pomo end`\n", short(sess.Hash), sess.StartTime.Local().Format("15:04:05"))
	}
	return nil
}
