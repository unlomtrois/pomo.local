package commands

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/adrg/xdg"
	"github.com/spf13/cobra"
	"pomo.local/internal/pomo"
)

var activeCmd = &cobra.Command{
	Use:   "active",
	Short: "Check or manage active pomodoro session",
	RunE:  runActive,
}

func init() {
	rootCmd.AddCommand(activeCmd)

	activeCmd.Flags().BoolP("verbose", "v", false, "Verbose output")
	activeCmd.Flags().Bool("remove", false, "Remove outdated active session")
}

func runActive(cmd *cobra.Command, _ []string) error {
	verbose, _ := cmd.Flags().GetBool("verbose")
	if verbose {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	remove, _ := cmd.Flags().GetBool("remove")

	activeSessionPath, err := xdg.StateFile("pomo/active_session.json")
	if err != nil {
		return nil
	}

	slog.Debug("Read active_session:", "path", activeSessionPath)
	data, err := os.ReadFile(activeSessionPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No active pomodoro session")
			return nil
		}
		return err
	}

	var session pomo.Session
	if err := json.Unmarshal(data, &session); err != nil {
		return err
	}
	fmt.Printf("Active session topic: %s, ends at: %s\n", session.Topic, session.StopTime.Format("15:04:05"))

	if time.Now().Compare(session.StopTime) > 0 {
		slog.Warn("active session is outdated")
		if remove {
			slog.Info("Removing active_session file:", "path", activeSessionPath)
			if err := os.Remove(activeSessionPath); err != nil {
				return err
			}
			slog.Info("Successfully removed active_session file:", "path", activeSessionPath)
		} else {
			slog.Info("You can remove it by adding --remove flag")
		}
	}

	return nil
}
