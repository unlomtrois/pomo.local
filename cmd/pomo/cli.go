package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

// CLIConfig holds all CLI configuration options
type CLIConfig struct {
	Duration         time.Duration
	Title            string
	Message          string
	SaveToCsv        bool
	NoNotify         bool
	MuteNotifySound  bool
	SaveInToggl      bool
	TogglToken       string
	TogglWorkspaceID int
	TogglUserID      int
	NotifySound      string
	ShowVersion      bool
}

// registerCommonFlags registers flags shared between start and rest commands
func registerCommonFlags(fs *flag.FlagSet, cfg *CLIConfig) {
	fs.BoolVar(&cfg.NoNotify, "no-notify", false, "Don't notify")
	fs.BoolVar(&cfg.MuteNotifySound, "mute", false, "Mute notify sound")
	fs.StringVar(&cfg.NotifySound, "notify-sound", "", "Notify sound")
	fs.BoolVar(&cfg.SaveInToggl, "toggl", false, "Save in Toggl")
	fs.BoolVar(&cfg.SaveToCsv, "csv", false, "Save to csv")
	fs.StringVar(&cfg.TogglToken, "token", "", "Toggl token")
	fs.IntVar(&cfg.TogglWorkspaceID, "workspace", 0, "Toggl workspace ID")
	fs.IntVar(&cfg.TogglUserID, "user", 0, "Toggl user ID")
}

// parseStartCommand parses flags for the "start" subcommand
func parseStartCommand(cfg *CLIConfig) {
	fs := flag.NewFlagSet("start", flag.ExitOnError)
	fs.DurationVar(&cfg.Duration, "d", 25*time.Minute, "Timer duration")
	fs.StringVar(&cfg.Message, "m", "Pomodoro finished! Time for a break.", "Notification message")
	registerCommonFlags(fs, cfg)
	fs.Parse(os.Args[2:])

	if fs.NArg() > 0 {
		cfg.Title = fs.Arg(0)
	} else {
		cfg.Title = "focus"
	}
}

// parseRestCommand parses flags for the "rest" subcommand
func parseRestCommand(cfg *CLIConfig) {
	fs := flag.NewFlagSet("rest", flag.ExitOnError)
	fs.DurationVar(&cfg.Duration, "d", 5*time.Minute, "Timer duration")
	fs.StringVar(&cfg.Message, "m", "Break finished! Time for a pomodoro.", "Notification message")
	registerCommonFlags(fs, cfg)
	fs.Parse(os.Args[2:])

	if fs.NArg() > 0 {
		cfg.Title = fs.Arg(0)
	} else {
		cfg.Title = "break"
	}
}

// parseVersionFlag parses the --version flag and returns true if version was requested
func parseVersionFlag() bool {
	var showVersion bool
	flag.BoolVar(&showVersion, "version", false, "Show current version")
	flag.Usage = printUsage
	flag.Parse()
	return showVersion
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "Usage: pomo <command> [options] \"your current focus\" ")
	fmt.Fprintln(os.Stderr, "Commands:")
	fmt.Fprintln(os.Stderr, "  start - Set a new pomodoro timer")
	fmt.Fprintln(os.Stderr, "  rest - Set a rest timer")
	fmt.Fprintln(os.Stderr, "Options:")
	fmt.Fprintln(os.Stderr, "  -d duration - Valid time units are \"ns\", \"us\" (or \"µs\"), \"ms\", \"s\", \"m\", \"h\".")
	fmt.Fprintln(os.Stderr, "  -m message - Notification message (default: Pomodoro finished! Time for a break for start, Break finished! Time for a pomodoro for rest)")
	fmt.Fprintln(os.Stderr, "  -t title - Notification title (default: Pomodoro Timer for start, Break Timer for rest)")
	fmt.Fprintln(os.Stderr, "  --toggl - Save in Toggl")
	fmt.Fprintln(os.Stderr, "  --csv - Save to csv")
	fmt.Fprintln(os.Stderr, "  --no-notify - Don't notify")
	fmt.Fprintln(os.Stderr, "  --mute - Mute notify sound")
	fmt.Fprintln(os.Stderr, "  --notify-sound <path/to/sound> - Notify sound")
	fmt.Fprintln(os.Stderr, "  --token <token> - Toggl token")
	fmt.Fprintln(os.Stderr, "  --workspace <workspaceId> - Toggl workspace ID")
	fmt.Fprintln(os.Stderr, "  --user <userId> - Toggl user ID")
	fmt.Fprintln(os.Stderr, "  --version - shows current version")
}
