package main

import (
	"os"

	"pomo.local/cmd/pomo/commands"
)

// it is filled by -ldflags="-X main.version=$(VERSION)"" in makefile
var version = "dev"

func main() {
	commands.SetVersion(version)
	if err := commands.Execute(); err != nil {
		os.Exit(1)
	}
}
