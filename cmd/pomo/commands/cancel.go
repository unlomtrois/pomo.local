package commands

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"pomo.local/internal/client"
)

var cancelCmd = &cobra.Command{
	Use:   "cancel",
	Short: "Cancel the active session (discard it — no time recorded)",
	Long: "Cancels the active session, whether a doro or an open `start` session. " +
		"Unlike `pomo end`, the session is discarded (marked cancelled), not recorded.",
	RunE: runCancel,
}

func init() {
	rootCmd.AddCommand(cancelCmd)
}

func runCancel(_ *cobra.Command, _ []string) error {
	ctx := context.Background()
	c, err := ensureDaemon(ctx, client.DefaultAddr())
	if err != nil {
		return err
	}

	sess, err := c.StopActive(ctx)
	if errors.Is(err, client.ErrNoActive) {
		return fmt.Errorf("no active session to cancel")
	}
	if err != nil {
		return err
	}

	topic := sess.Topic
	if topic == "" {
		topic = "(no topic)"
	}
	fmt.Printf("Cancelled [%s] %s\n", short(sess.Hash), topic)
	return nil
}
