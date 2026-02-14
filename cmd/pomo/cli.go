package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"pomo.local/internal/pomo"
)

// CLIConfig holds all CLI configuration options
type CLIConfig struct { // TODO: I notice it becoming hard to manage args in internal/pomo, because I cannot simply import cmd/pomo/cli this thing from there. Do something about it later
	Timer struct {
		Duration time.Duration
		Title    string
		Message  string
	}
	Toggl struct {
		Enabled     bool
		Token       string
		WorkspaceID int
		UserID      int
	}
	Notifications struct {
		Disabled bool
		Mute     bool
		Sound    string
		Hint     string
	}
	SaveToCsv   bool
	ShowVersion bool
	Verbose     bool
}

// registerCommonFlags registers flags shared between start and rest commands
func registerCommonFlags(fs *flag.FlagSet, cfg *CLIConfig, fileCfg *pomo.FileConfig) {
	// Notifications
	fs.BoolVar(&cfg.Notifications.Disabled, "no-notify", false, "Don't notify")
	fs.BoolVar(&cfg.Notifications.Mute, "mute", fileCfg.Notifications.Mute, "Mute notify sound")
	fs.StringVar(&cfg.Notifications.Sound, "notify-sound", fileCfg.Notifications.Sound, "Notify sound")
	// Toggl
	fs.BoolVar(&cfg.Toggl.Enabled, "toggl", false, "Save in Toggl")
	fs.StringVar(&cfg.Toggl.Token, "token", fileCfg.Toggl.Token, "Toggl token")
	fs.IntVar(&cfg.Toggl.WorkspaceID, "workspace", fileCfg.Toggl.WorkspaceID, "Toggl workspace ID")
	fs.IntVar(&cfg.Toggl.UserID, "user", fileCfg.Toggl.UserID, "Toggl user ID")
	// Other
	fs.BoolVar(&cfg.SaveToCsv, "csv", false, "Save to csv")
	fs.BoolVar(&cfg.Verbose, "verbose", false, "--verbose")
}

// parseStartCommand parses flags for the "start" subcommand
func parseStartCommand(cfg *CLIConfig, fileCfg *pomo.FileConfig) {
	// Parse duration from file config, fallback to 25m
	defaultDuration := 25 * time.Minute
	if fileCfg.Pomodoro.DefaultDuration != "" {
		if d, err := time.ParseDuration(fileCfg.Pomodoro.DefaultDuration); err == nil {
			defaultDuration = d
		}
	}

	// Use file config for message default
	defaultMessage := "Pomodoro finished! Time for a break."
	if fileCfg.Pomodoro.DefaultMessage != "" {
		defaultMessage = fileCfg.Pomodoro.DefaultMessage
	}

	defaultTitle := "focus"

	startCmd := flag.NewFlagSet("start", flag.ExitOnError)
	startCmd.DurationVar(&cfg.Timer.Duration, "d", defaultDuration, "Timer duration")
	startCmd.StringVar(&cfg.Timer.Title, "t", defaultTitle, "Title")
	startCmd.StringVar(&cfg.Timer.Message, "m", defaultMessage, "Notification message")
	registerCommonFlags(startCmd, cfg, fileCfg)
	startCmd.Parse(os.Args[2:])
}

// parseRestCommand parses flags for the "rest" subcommand
func parseRestCommand(cfg *CLIConfig, fileCfg *pomo.FileConfig) {
	// Parse duration from file config, fallback to 5m
	defaultDuration := 5 * time.Minute
	if fileCfg.Rest.DefaultDuration != "" {
		if d, err := time.ParseDuration(fileCfg.Rest.DefaultDuration); err == nil {
			defaultDuration = d
		}
	}

	// Use file config for message default
	defaultMessage := "Break finished! Time for a pomodoro."
	if fileCfg.Rest.DefaultMessage != "" {
		defaultMessage = fileCfg.Rest.DefaultMessage
	}
	defaultTitle := "break"

	restCmd := flag.NewFlagSet("rest", flag.ExitOnError)
	restCmd.DurationVar(&cfg.Timer.Duration, "d", defaultDuration, "Timer duration")
	restCmd.StringVar(&cfg.Timer.Title, "t", defaultTitle, "Title")
	restCmd.StringVar(&cfg.Timer.Message, "m", defaultMessage, "Notification message")
	registerCommonFlags(restCmd, cfg, fileCfg)
	restCmd.Parse(os.Args[2:])
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
	fmt.Fprintln(os.Stderr, "Usage: pomo <command> [options]")
	fmt.Fprintln(os.Stderr, "Commands:")
	fmt.Fprintln(os.Stderr, "  start - Set a new pomodoro timer")
	fmt.Fprintln(os.Stderr, "  rest - Set a rest timer")
	fmt.Fprintln(os.Stderr, "  notify - Immediate notify")
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
