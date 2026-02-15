package cli

import (
	"fmt"
	"os"
)

func PrintUsage() {
	fmt.Fprintln(os.Stderr, "Usage: pomo <command> [options]")
	fmt.Fprintln(os.Stderr, "Commands:")
	fmt.Fprintln(os.Stderr, "  start - Set a new pomodoro timer")
	fmt.Fprintln(os.Stderr, "  rest - Alias to \"start\" command, but session is not saved/tracked")
	fmt.Fprintln(os.Stderr, "  notify - Immediate notify, for convenience")
	fmt.Fprintln(os.Stderr, "  auth - Setup configs to work with integrations")
	fmt.Fprintln(os.Stderr, "Options:")
	fmt.Fprintln(os.Stderr, "  -d duration - Valid time units are \"s\", \"m\", \"h\".")
	fmt.Fprintln(os.Stderr, "  -m message - Notification message (default: Pomodoro finished! Time for a break for start, Break finished! Time for a pomodoro for rest)")
	fmt.Fprintln(os.Stderr, "  -t title - Notification title (default: Pomodoro Timer for start, Break Timer for rest)")
	fmt.Fprintln(os.Stderr, "  --toggl - Save in Toggl")
	fmt.Fprintln(os.Stderr, "  --email - Notify yourself via email (e.g. in case you went for lunch, notify that break is over)")
	// fmt.Fprintln(os.Stderr, "  --csv - Save to csv")
	// fmt.Fprintln(os.Stderr, "  --no-notify - Don't notify")
	// fmt.Fprintln(os.Stderr, "  --mute - Mute notify sound")
	// fmt.Fprintln(os.Stderr, "  --notify-sound <path/to/sound> - Notify sound")
	// fmt.Fprintln(os.Stderr, "  --token <token> - Toggl token")
	// fmt.Fprintln(os.Stderr, "  --workspace <workspaceId> - Toggl workspace ID")
	// fmt.Fprintln(os.Stderr, "  --user <userId> - Toggl user ID")
	fmt.Fprintln(os.Stderr, "  -v, --version - shows current version")
}
