package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"pomo.local/internal/client"
	"pomo.local/internal/utils"
)

var logCmd = &cobra.Command{
	Use:   "log",
	Short: "List recent sessions (git log-style); scoped to the current project",
	Long: "Lists recent sessions newest-first. Inside a .pomo project it shows that " +
		"project's sessions; use --all to list across all projects.",
	RunE: runLog,
}

func init() {
	rootCmd.AddCommand(logCmd)

	logCmd.Flags().IntP("number", "n", 20, "Limit the number of sessions shown")
	logCmd.Flags().Bool("all", false, "Show sessions across all projects")
}

func runLog(cmd *cobra.Command, _ []string) error {
	limit, _ := cmd.Flags().GetInt("number")
	all, _ := cmd.Flags().GetBool("all")

	ctx := context.Background()
	c, err := ensureDaemon(ctx, client.DefaultAddr())
	if err != nil {
		return err
	}

	var projectExtID, scopeNote string
	if !all {
		if proj := discoverProject(); proj != nil {
			projectExtID = proj.ID
			scopeNote = " in project " + proj.Name
		}
	}

	sessions, err := c.ListSessions(ctx, limit, projectExtID)
	if err != nil {
		return err
	}
	if len(sessions) == 0 {
		fmt.Printf("No sessions%s\n", scopeNote)
		return nil
	}

	for _, s := range sessions {
		topic := s.Topic
		if topic == "" {
			topic = "(no topic)"
		}
		dur := utils.ShortDuration(time.Duration(s.Duration).Round(time.Second))
		when := s.StartTime.Local().Format("Mon 02 Jan 15:04")
		marker := statusMarker(s)
		fmt.Printf("%s [%s] %-6s %s  %s\n", marker, short(s.Hash), dur, when, topic)
	}
	return nil
}

// statusMarker is a one-glyph hint: running, cancelled, or done.
func statusMarker(s client.Session) string {
	switch s.Status {
	case "active":
		return "▶"
	case "cancelled":
		return "✗"
	default:
		return " "
	}
}
