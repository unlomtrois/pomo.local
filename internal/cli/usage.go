package cli

import (
	"fmt"
	"os"
)

func PrintUsage() {
	fmt.Fprintln(os.Stderr, "Usage: pomo <command> [options]")
	fmt.Fprintln(os.Stderr, "Commands:")
	fmt.Fprintln(os.Stderr, "  start - Set a new pomodoro timer")
	fmt.Fprintln(os.Stderr, "  rest - alias to start with 5 minutes")
	fmt.Fprintln(os.Stderr, "  notify - Immediate notify")
	fmt.Fprintln(os.Stderr, "Options:")
	fmt.Fprintln(os.Stderr, "  -d duration - Valid time units are \"s\", \"m\", \"h\".")
	fmt.Fprintln(os.Stderr, "  -m message - Notification message (default: Pomodoro finished! Time for a break for start, Break finished! Time for a pomodoro for rest)")
	fmt.Fprintln(os.Stderr, "  -t title - Notification title (default: Pomodoro Timer for start, Break Timer for rest)")
	fmt.Fprintln(os.Stderr, "  --toggl - Save in Toggl")
	// fmt.Fprintln(os.Stderr, "  --csv - Save to csv")
	// fmt.Fprintln(os.Stderr, "  --no-notify - Don't notify")
	// fmt.Fprintln(os.Stderr, "  --mute - Mute notify sound")
	// fmt.Fprintln(os.Stderr, "  --notify-sound <path/to/sound> - Notify sound")
	// fmt.Fprintln(os.Stderr, "  --token <token> - Toggl token")
	// fmt.Fprintln(os.Stderr, "  --workspace <workspaceId> - Toggl workspace ID")
	// fmt.Fprintln(os.Stderr, "  --user <userId> - Toggl user ID")
	fmt.Fprintln(os.Stderr, "  --version - shows current version")
}
