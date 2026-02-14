package main

import (
	"fmt"
	"os"

	"pomo.local/internal/cli"
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

func main() {
	if len(os.Args) < 2 {
		cli.PrintUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "start":
		if err := cli.ParseStart(os.Args[2:]).Run(); err != nil {
			fatal("Error running \"pomo start\": %w", err)
		}
	case "rest":
		if err := cli.ParseRest(os.Args[2:]).Run(); err != nil {
			fatal("Error running \"pomo start\": %w", err)
		}
		os.Exit(0)
	case "notify":
		if err := cli.ParseNotify(os.Args[2:]).Run(); err != nil {
			fatal("Error running pomo notify: %w", err)
		}
		os.Exit(0)
	case "auth":
		if err := cli.ParseAuth(os.Args[2:]).Run(); err != nil {
			fatal("Failed to auth: %w", err)
		}
		os.Exit(0)
	default:
		if cli.ParseVersionFlag() {
			fmt.Fprintln(os.Stderr, version)
			os.Exit(0)
		}
		fmt.Fprintln(os.Stderr, "unknown command!")
		cli.PrintUsage()
		os.Exit(1)
	}
}
