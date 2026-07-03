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

var showCmd = &cobra.Command{
	Use:   "show <hash>",
	Short: "Show a session by its hash (or a short prefix, git-style)",
	Args:  cobra.ExactArgs(1),
	RunE:  runShow,
}

func init() {
	rootCmd.AddCommand(showCmd)
}

func runShow(_ *cobra.Command, args []string) error {
	ctx := context.Background()
	c, err := ensureDaemon(ctx, client.DefaultAddr)
	if err != nil {
		return err
	}

	sess, err := c.SessionByHash(ctx, args[0])
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
	fmt.Printf("hash     %s\n", sess.Hash)
	fmt.Printf("topic    %s\n", topic)
	fmt.Printf("started  %s\n", sess.StartTime.Local().Format("2006-01-02 15:04:05"))
	if sess.Status != "active" {
		fmt.Printf("ended    %s\n", sess.StopTime.Local().Format("2006-01-02 15:04:05"))
		fmt.Printf("duration %s\n", utils.ShortDuration(time.Duration(sess.Duration).Round(time.Second)))
	}
	fmt.Printf("status   %s\n", sess.Status)
	fmt.Printf("source   %s\n", sess.Source)
	return nil
}

// short returns the git-style short form of a session hash.
func short(hash string) string {
	if len(hash) >= 7 {
		return hash[:7]
	}
	return hash
}
