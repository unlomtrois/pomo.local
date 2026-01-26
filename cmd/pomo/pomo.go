package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"pomo.local/internal/pomo"

	"time"
)

// it is filled by -ldflags="-X main.version=$(VERSION)"" in makefile
var version string = "dev"

func main() {
	if version == "dev" {
		log.Println("Warning: running development build")
	}

	startCmd := flag.NewFlagSet("start", flag.ExitOnError)
	restCmd := flag.NewFlagSet("rest", flag.ExitOnError)

	var duration int
	var title string
	var message string
	var saveToCsv bool
	var noNotify bool
	var saveInToggl bool
	var togglToken string
	var toggleWorkspaceId int
	var toggleUserId int
	var showVersion bool

	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: pomo <command> [options]")
		fmt.Fprintln(os.Stderr, "Commands:")
		fmt.Fprintln(os.Stderr, "  start - Set a new pomodoro timer")
		fmt.Fprintln(os.Stderr, "  rest - Set a rest timer")
		fmt.Fprintln(os.Stderr, "Options:")
		fmt.Fprintln(os.Stderr, "  -d duration - Duration in minutes (default: 25 for new, 5 for rest)")
		fmt.Fprintln(os.Stderr, "  -m message - Notification message (default: Pomodoro finished! Time for a break for new, Break finished! Time for a pomodoro for rest)")
		fmt.Fprintln(os.Stderr, "  -t title - Notification title (default: Pomodoro Timer for new, Break Timer for rest)")
		fmt.Fprintln(os.Stderr, "  --toggl - Save in Toggl")
		fmt.Fprintln(os.Stderr, "  --csv - Save to csv")
		fmt.Fprintln(os.Stderr, "  --no-notify - Don't notify")
		fmt.Fprintln(os.Stderr, "  --token <token> - Toggl token")
		fmt.Fprintln(os.Stderr, "  --workspace <workspaceId> - Toggl workspace ID")
		fmt.Fprintln(os.Stderr, "  --user <userId> - Toggl user ID")
		fmt.Fprintln(os.Stderr, "  --version - shows current version")
	}

	if len(os.Args) < 2 {
		flag.Usage()
		os.Exit(1)
	}

	switch os.Args[1] { // todo: refactor
	case "start":
		startCmd.IntVar(&duration, "d", 25, "Duration in minutes (default: 25)")
		startCmd.StringVar(&title, "t", "Pomodoro Timer", "Notification title")
		startCmd.StringVar(&message, "m", "Pomodoro finished! Time for a break.", "Notification message")
		startCmd.BoolVar(&noNotify, "no-notify", false, "Don't notify")
		startCmd.BoolVar(&saveInToggl, "toggl", false, "Save in Toggl")
		startCmd.BoolVar(&saveToCsv, "csv", false, "Save to csv")
		startCmd.StringVar(&togglToken, "token", "", "Toggl token")
		startCmd.IntVar(&toggleWorkspaceId, "workspace", 0, "Toggl workspace ID")
		startCmd.IntVar(&toggleUserId, "user", 0, "Toggl user ID")
		startCmd.Parse(os.Args[2:])
	case "rest":
		restCmd.IntVar(&duration, "d", 5, "Duration in minutes (default: 5)")
		restCmd.StringVar(&title, "t", "Break Timer", "Notification title")
		restCmd.BoolVar(&saveInToggl, "toggl", false, "Save in Toggl")
		restCmd.BoolVar(&saveToCsv, "csv", false, "Save to csv")
		restCmd.StringVar(&message, "m", "Break finished! Time for a pomodoro.", "Notification message")
		restCmd.BoolVar(&noNotify, "no-notify", false, "Don't notify")
		restCmd.StringVar(&togglToken, "token", "", "Toggl token")
		restCmd.IntVar(&toggleWorkspaceId, "workspace", 0, "Toggl workspace ID")
		restCmd.IntVar(&toggleUserId, "user", 0, "Toggl user ID")
		restCmd.Parse(os.Args[2:])
	default:
		flag.BoolVar(&showVersion, "version", false, "--version - show current version")
		flag.Parse()
		if !showVersion {
			flag.Usage()
			os.Exit(1)
		}
	}

	if showVersion {
		fmt.Fprintln(os.Stderr, version)
		os.Exit(0)
	}

	if duration <= 0 {
		fmt.Fprintln(os.Stderr, "Error: duration must be positive")
		os.Exit(1)
	}

	pomodoro := pomo.NewPomodoro(title, message, time.Duration(duration)*time.Minute)

	if saveToCsv {
		if err := pomo.InitCsv("pomodoro.csv"); err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing pomodoro.csv: %v\n", err)
			os.Exit(1)
		}
		if err := pomodoro.Save("pomodoro.csv"); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving pomodoro: %v\n", err)
			os.Exit(1)
		}
	}

	if saveInToggl {
		if togglToken == "" {
			fmt.Fprintln(os.Stderr, "Error: toggl token is required")
			os.Exit(1)
		}
		if toggleWorkspaceId == 0 {
			fmt.Fprintln(os.Stderr, "Error: toggl workspace id is required")
			os.Exit(1)
		}
		if toggleUserId == 0 {
			fmt.Fprintln(os.Stderr, "Error: toggl user id is required")
			os.Exit(1)
		}
		if err := pomodoro.SaveInToggl(togglToken, toggleWorkspaceId, toggleUserId); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving in Toggl: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Pomodoro saved in Toggl")
	}

	if noNotify && !saveInToggl && !saveToCsv {
		fmt.Fprintln(os.Stderr, "no action to perform (set --toggl or --csv)")
		os.Exit(0)
	} else {
		fmt.Printf("🍅 Pomodoro timer set for %d minutes\n", duration)
	}

	if noNotify {
		fmt.Println("No notification")
		os.Exit(0)
	}

	if err := pomodoro.Notify(); err != nil {
		fmt.Fprintf(os.Stderr, "Error notifying: %v\n", err)
		os.Exit(1)
	}
}
