package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/adrg/xdg"
	"pomo.local/internal/pomo"
)

type ActiveCommand struct {
	verbose bool
	remove  bool
}

func ParseActive(args []string) *ActiveCommand {
	cmd := ActiveCommand{}
	fs := flag.NewFlagSet("active", flag.ExitOnError)
	fs.BoolVar(&cmd.verbose, "v", false, "Verbose output, e.g. opened files, making http requests")
	fs.BoolVar(&cmd.verbose, "verbose", false, "Verbose output, e.g. opened files, making http requests")
	fs.BoolVar(&cmd.remove, "remove", false, "Remove active session? e.g. if it is outdated and was not removed automatically")
	fs.Parse(args)

	if cmd.verbose {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	return &cmd
}

func (cmd *ActiveCommand) Run() error {
	activeSessionPath, err := xdg.StateFile("pomo/active_session.json")
	if err != nil {
		return nil
	}

	slog.Debug("Read active session:", "path", activeSessionPath)
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
		if cmd.remove {
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
