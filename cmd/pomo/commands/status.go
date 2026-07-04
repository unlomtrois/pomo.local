package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/adrg/xdg"
	"github.com/spf13/cobra"
	"pomo.local/internal/client"
	"pomo.local/internal/utils"
)

var statusCmd = &cobra.Command{
	Use:     "status",
	Aliases: []string{"stats", "health"},
	Short:   "Show daemon and database status",
	// Deliberately does NOT auto-spawn the daemon — the point is to report
	// whether it's actually running.
	RunE: runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(_ *cobra.Command, _ []string) error {
	// Database file (inspected locally; independent of the daemon).
	dbPath := filepath.Join(xdg.DataHome, "pomo", "pomo.db")
	if fi, err := os.Stat(dbPath); err == nil {
		fmt.Printf("database  %s (%s)\n", dbPath, humanSize(fi.Size()))
	} else {
		fmt.Printf("database  %s (not created yet)\n", dbPath)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	c := client.New(client.DefaultAddr)
	if err := c.Health(ctx); err != nil {
		fmt.Printf("daemon    not running at %s\n", client.DefaultAddr)
		fmt.Println("          (any pomo command auto-spawns it; or run `pomo daemon`)")
		return nil
	}

	info, err := c.Status(ctx)
	if err != nil {
		// Reachable but stats failed (e.g. older daemon without /api/stats).
		fmt.Printf("daemon    running at %s (stats unavailable: %v)\n", client.DefaultAddr, err)
		return nil
	}

	version := info.Version
	if version == "" {
		version = "unknown"
	}
	fmt.Printf("daemon    running (%s) at %s\n", version, client.DefaultAddr)
	fmt.Printf("sessions  %d total", info.Sessions.TotalSessions)
	if s := info.Sessions.ByStatus; len(s) > 0 {
		fmt.Printf(" — %d done, %d cancelled, %d active",
			s["done"], s["cancelled"], s["active"])
	}
	fmt.Println()
	fmt.Printf("projects  %d\n", info.Sessions.Projects)
	fmt.Printf("tracked   %s (across done sessions)\n",
		utils.ShortDuration(time.Duration(info.Sessions.TrackedNanos).Round(time.Second)))

	if a := info.Active; a != nil {
		topic := a.Topic
		if topic == "" {
			topic = "(no topic)"
		}
		if a.Open {
			fmt.Printf("active    [%s] %s (open, since %s)\n",
				short(a.Hash), topic, a.StartTime.Local().Format("15:04"))
		} else {
			fmt.Printf("active    [%s] %s (ends %s)\n",
				short(a.Hash), topic, a.StopTime.Local().Format("15:04"))
		}
	} else {
		fmt.Println("active    none")
	}
	return nil
}

// humanSize formats a byte count as B/KB/MB.
func humanSize(n int64) string {
	switch {
	case n >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(n)/(1<<20))
	case n >= 1<<10:
		return fmt.Sprintf("%.1f KB", float64(n)/(1<<10))
	default:
		return fmt.Sprintf("%d B", n)
	}
}
