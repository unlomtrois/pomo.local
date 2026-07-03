package commands

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"pomo.local/internal/client"
)

var editCmd = &cobra.Command{
	Use:   "edit <hash>",
	Short: "Edit a past session by hash (partial update)",
	Long: "Edits a session resolved by hash prefix. Only the flags you pass change; " +
		"stop time is kept consistent as start + duration.",
	Args: cobra.ExactArgs(1),
	RunE: runEdit,
}

func init() {
	rootCmd.AddCommand(editCmd)

	editCmd.Flags().StringP("topic", "t", "", "Set the topic (pass \"\" to clear)")
	editCmd.Flags().String("start", "", "Set the start time (RFC3339, \"2006-01-02 15:04\", or \"15:04\")")
	editCmd.Flags().DurationP("duration", "d", 0, "Set the duration (e.g. 40m, 1h30m)")
}

func runEdit(cmd *cobra.Command, args []string) error {
	flags := cmd.Flags()
	if !flags.Changed("topic") && !flags.Changed("start") && !flags.Changed("duration") {
		return fmt.Errorf("nothing to edit — pass --topic, --start, and/or --duration")
	}

	var params client.EditParams
	if flags.Changed("topic") {
		topic, _ := flags.GetString("topic")
		params.Topic = &topic
	}
	if flags.Changed("start") {
		raw, _ := flags.GetString("start")
		t, err := parseTimeArg(raw)
		if err != nil {
			return err
		}
		params.Start = &t
	}
	if flags.Changed("duration") {
		d, _ := flags.GetDuration("duration")
		ds := d.String()
		params.Duration = &ds
	}

	ctx := context.Background()
	c, err := ensureDaemon(ctx, client.DefaultAddr)
	if err != nil {
		return err
	}

	sess, err := c.EditByHash(ctx, args[0], params)
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
	fmt.Printf("[%s] %s — %s to %s\n", short(sess.Hash), topic,
		sess.StartTime.Local().Format("2006-01-02 15:04"),
		sess.StopTime.Local().Format("15:04"))
	return nil
}

// parseTimeArg accepts a few human-friendly local time formats, plus RFC3339.
// A bare "15:04" means that time today.
func parseTimeArg(s string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	for _, layout := range []string{"2006-01-02 15:04:05", "2006-01-02 15:04"} {
		if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			return t, nil
		}
	}
	if t, err := time.ParseInLocation("15:04", s, time.Local); err == nil {
		now := time.Now()
		return time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), 0, 0, time.Local), nil
	}
	return time.Time{}, fmt.Errorf("could not parse time %q (try \"2006-01-02 15:04\" or \"15:04\")", s)
}
