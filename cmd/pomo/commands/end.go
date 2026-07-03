package commands

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"pomo.local/internal/client"
	"pomo.local/internal/utils"
)

var endCmd = &cobra.Command{
	Use:     "end [topic]",
	Aliases: []string{"stop"},
	Short:   "End the active session, recording elapsed time; optionally (re)name it",
	Long: "Ends the active session and records how long it actually ran. Pass a " +
		"topic to set it (or override the one given at start).",
	Args: cobra.MaximumNArgs(1),
	RunE: runEnd,
}

func init() {
	rootCmd.AddCommand(endCmd)
}

func runEnd(_ *cobra.Command, args []string) error {
	ctx := context.Background()

	c, err := ensureDaemon(ctx, client.DefaultAddr)
	if err != nil {
		return err
	}

	var topic string
	if len(args) == 1 {
		topic = args[0]
	}

	sess, err := c.EndActive(ctx, topic)
	if errors.Is(err, client.ErrNoActive) {
		return fmt.Errorf("no active session to end")
	}
	if err != nil {
		return err
	}

	elapsed := utils.ShortDuration(time.Duration(sess.Duration).Round(time.Second))
	if sess.Topic != "" {
		fmt.Printf("Ended %q after %s\n", sess.Topic, elapsed)
	} else {
		fmt.Printf("Ended session after %s\n", elapsed)
	}
	return nil
}
