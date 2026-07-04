package commands

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"pomo.local/internal/client"
)

var rmCmd = &cobra.Command{
	Use:     "rm <hash>",
	Aliases: []string{"delete"},
	Short:   "Delete a session by its hash (or a short prefix)",
	Args:    cobra.ExactArgs(1),
	RunE:    runRm,
}

func init() {
	rootCmd.AddCommand(rmCmd)
}

func runRm(_ *cobra.Command, args []string) error {
	ctx := context.Background()
	c, err := ensureDaemon(ctx, client.DefaultAddr())
	if err != nil {
		return err
	}

	sess, err := c.DeleteByHash(ctx, args[0])
	if errors.Is(err, client.ErrSessionNotFound) {
		return fmt.Errorf("no session matching %q", args[0])
	}
	if errors.Is(err, client.ErrAmbiguousHash) {
		return fmt.Errorf("ambiguous hash %q — use more characters", args[0])
	}
	if err != nil {
		return err
	}

	topic := sess.Topic
	if topic == "" {
		topic = "(no topic)"
	}
	fmt.Printf("Deleted [%s] %s\n", short(sess.Hash), topic)
	return nil
}
