package commands

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "pomo",
	Short: "Pomodoro timer CLI",
}

// Execute runs the root cobra command.
func Execute() error {
	return rootCmd.Execute()
}

// SetVersion sets the CLI version string shown in --version output.
func SetVersion(v string) {
	rootCmd.Version = v
}
