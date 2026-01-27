package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"pomo.local/internal/pomo"
	"pomo.local/internal/utils"

	"time"
)

// it is filled by -ldflags="-X main.version=$(VERSION)"" in makefile
var version string = "dev"

// Config holds all CLI configuration options
type Config struct {
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

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

// registerCommonFlags registers flags shared between start and rest commands
func registerCommonFlags(fs *flag.FlagSet, cfg *Config) {
	fs.BoolVar(&cfg.NoNotify, "no-notify", false, "Don't notify")
	fs.BoolVar(&cfg.MuteNotifySound, "mute", false, "Mute notify sound")
	fs.StringVar(&cfg.NotifySound, "notify-sound", "", "Notify sound")
	fs.BoolVar(&cfg.SaveInToggl, "toggl", false, "Save in Toggl")
	fs.BoolVar(&cfg.SaveToCsv, "csv", false, "Save to csv")
	fs.StringVar(&cfg.TogglToken, "token", "", "Toggl token")
	fs.IntVar(&cfg.TogglWorkspaceID, "workspace", 0, "Toggl workspace ID")
	fs.IntVar(&cfg.TogglUserID, "user", 0, "Toggl user ID")
}

// run executes the pomodoro timer with the given configuration
func run(cfg *Config) {
	if cfg.Duration <= 0 {
		fatal("Error: duration must be positive")
	}

	pomodoro := pomo.NewPomodoro(cfg.Title, cfg.Message, cfg.Duration)

	if cfg.SaveToCsv {
		if err := pomo.InitCsv("pomodoro.csv"); err != nil {
			fatal("Error initializing pomodoro.csv: %v", err)
		}
		if err := pomodoro.Save("pomodoro.csv"); err != nil {
			fatal("Error saving pomodoro: %v", err)
		}
	}

	if cfg.SaveInToggl {
		if cfg.TogglToken == "" {
			fatal("Error: toggl token is required")
		}
		if cfg.TogglWorkspaceID == 0 {
			fatal("Error: toggl workspace id is required")
		}
		if cfg.TogglUserID == 0 {
			fatal("Error: toggl user id is required")
		}
		if err := pomodoro.SaveInToggl(cfg.TogglToken, cfg.TogglWorkspaceID, cfg.TogglUserID); err != nil {
			fatal("Error saving in Toggl: %v", err)
		}
		fmt.Println("Pomodoro saved in Toggl")
	}

	if cfg.NoNotify && !cfg.SaveInToggl && !cfg.SaveToCsv {
		fmt.Fprintln(os.Stderr, "no action to perform (set --toggl or --csv)")
		os.Exit(0)
	}

	fmt.Printf("🍅 Pomodoro timer set for %s\n", utils.ShortDuration(cfg.Duration))

	if cfg.NoNotify {
		fmt.Println("No notification")
		os.Exit(0)
	}

	if cfg.NotifySound != "" {
		fmt.Println("Using custom notify sound:", cfg.NotifySound)
		cfg.NotifySound = filepath.Clean(cfg.NotifySound)
		if _, err := os.Stat(cfg.NotifySound); err != nil {
			fatal("Error: notify sound file does not exist: %v", err)
		}
	}

	if err := pomodoro.Notify(cfg.MuteNotifySound, cfg.NotifySound); err != nil {
		fatal("Error notifying: %v", err)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "Usage: pomo <command> [options]")
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

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cfg := &Config{}

	switch os.Args[1] {
	case "start":
		fs := flag.NewFlagSet("start", flag.ExitOnError)
		fs.DurationVar(&cfg.Duration, "d", 25*time.Minute, "Timer duration")
		fs.StringVar(&cfg.Title, "t", "Pomodoro Timer", "Notification title")
		fs.StringVar(&cfg.Message, "m", "Pomodoro finished! Time for a break.", "Notification message")
		registerCommonFlags(fs, cfg)
		fs.Parse(os.Args[2:])

	case "rest":
		fs := flag.NewFlagSet("rest", flag.ExitOnError)
		fs.DurationVar(&cfg.Duration, "d", 5*time.Minute, "Timer duration")
		fs.StringVar(&cfg.Title, "t", "Break Timer", "Notification title")
		fs.StringVar(&cfg.Message, "m", "Break finished! Time for a pomodoro.", "Notification message")
		registerCommonFlags(fs, cfg)
		fs.Parse(os.Args[2:])

	default:
		flag.BoolVar(&cfg.ShowVersion, "version", false, "Show current version")
		flag.Parse()
		if cfg.ShowVersion {
			fmt.Fprintln(os.Stderr, version)
			os.Exit(0)
		}
		printUsage()
		os.Exit(1)
	}

	run(cfg)
}
