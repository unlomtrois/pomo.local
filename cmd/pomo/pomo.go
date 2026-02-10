package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"pomo.local/internal/pomo"
	"pomo.local/internal/utils"
)

// it is filled by -ldflags="-X main.version=$(VERSION)"" in makefile
var version string = "dev"

var configDir string // XDG_CONFIG_HOME
var dataDir string   // XDG_DATA_HOME

func init() {
	var err error
	if configDir, err = utils.GetConfigDir(); err != nil {
		fatal("Error getting config directory: %v", err)
	}
	if dataDir, err = utils.GetDataDir(); err != nil {
		fatal("Error getting data directory: %v", err)
	}
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

// run executes the pomodoro timer with the given configuration
func run(cfg *CLIConfig) {
	if cfg.Timer.Duration < 5*time.Minute {
		fatal("please, focus more than 5 minutes")
	}

	pomodoro := pomo.NewPomodoro(cfg.Timer.Title, cfg.Timer.Message, cfg.Timer.Duration)

	if cfg.SaveToCsv {
		if err := pomo.InitCsv("pomodoro.csv"); err != nil {
			fatal("Error initializing pomodoro.csv: %v", err)
		}
		if err := pomodoro.Save("pomodoro.csv"); err != nil {
			fatal("Error saving pomodoro: %v", err)
		}
	}

	if cfg.Toggl.Enabled {
		if cfg.Toggl.Token == "" {
			fatal("Error: toggl token is required")
		}
		if cfg.Toggl.WorkspaceID == 0 {
			fatal("Error: toggl workspace id is required")
		}
		if cfg.Toggl.UserID == 0 {
			fatal("Error: toggl user id is required")
		}
		if err := pomodoro.SaveInToggl(cfg.Toggl.Token, cfg.Toggl.WorkspaceID, cfg.Toggl.UserID); err != nil {
			fatal("Error saving in Toggl: %v", err)
		}
		fmt.Println("Pomodoro saved in Toggl")
	}

	if cfg.Notifications.Disabled && !cfg.Toggl.Enabled && !cfg.SaveToCsv {
		fmt.Fprintln(os.Stderr, "no action to perform (set --toggl or --csv)")
		os.Exit(0)
	}

	fmt.Printf("🍅 Pomodoro timer set for %s\n", utils.ShortDuration(cfg.Timer.Duration))

	if cfg.Notifications.Disabled {
		fmt.Println("No notification")
		os.Exit(0)
	}

	if cfg.Notifications.Sound != "" {
		fmt.Println("Using custom notify sound:", cfg.Notifications.Sound)
		cfg.Notifications.Sound = filepath.Clean(cfg.Notifications.Sound)
		if _, err := os.Stat(cfg.Notifications.Sound); err != nil {
			fatal("Error: notify sound file does not exist: %v", err)
		}
	}

	if err := pomodoro.Notify(cfg.Notifications.Mute, cfg.Notifications.Sound); err != nil {
		fatal("Error notifying: %v", err)
	}
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cfg := &CLIConfig{}

	// Load file config (creates default if not exists)
	fileCfg, err := pomo.LoadConfig(configDir)
	if err != nil {
		fatal("Error loading config: %v", err)
	}

	switch os.Args[1] {
	case "start":
		parseStartCommand(cfg, fileCfg)
	case "rest":
		parseRestCommand(cfg, fileCfg)
	default:
		if parseVersionFlag() {
			fmt.Fprintln(os.Stderr, version)
			os.Exit(0)
		}
		printUsage()
		os.Exit(1)
	}

	run(cfg)
}
