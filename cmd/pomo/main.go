package main

import (
	"flag"
	"fmt"
	"os"

	"pomo.local/internal/pomo"

	"time"
)

func main() {
	addCmd := flag.NewFlagSet("add", flag.ExitOnError)
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

	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: pomo <command> [options]")
		fmt.Fprintln(os.Stderr, "Commands:")
		fmt.Fprintln(os.Stderr, "  add - Set a new pomodoro timer")
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
	}

	if len(os.Args) < 2 {
		flag.Usage()
		os.Exit(1)
	}

	switch os.Args[1] { // todo: refactor
	case "add":
		addCmd.IntVar(&duration, "d", 25, "Duration in minutes (default: 25)")
		addCmd.StringVar(&title, "t", "Pomodoro Timer", "Notification title")
		addCmd.StringVar(&message, "m", "Pomodoro finished! Time for a break.", "Notification message")
		addCmd.BoolVar(&noNotify, "no-notify", false, "Don't notify")
		addCmd.BoolVar(&saveInToggl, "toggl", false, "Save in Toggl")
		addCmd.BoolVar(&saveToCsv, "csv", false, "Save to csv")
		addCmd.StringVar(&togglToken, "token", "", "Toggl token")
		addCmd.IntVar(&toggleWorkspaceId, "workspace", 0, "Toggl workspace ID")
		addCmd.IntVar(&toggleUserId, "user", 0, "Toggl user ID")
		addCmd.Parse(os.Args[2:])
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
		flag.Usage()
		os.Exit(1)
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
