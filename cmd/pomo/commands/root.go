package commands

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "pomo",
	Short: "Pomodoro timer CLI",
}

// commandGroups maps each subcommand name to a help group. Assigned centrally
// (in Execute) so the command files don't each need to know their group.
var commandGroups = map[string]string{
	// Session lifecycle.
	"doro": "session", "start": "session", "end": "session",
	"cancel": "session", "rest": "session", "active": "session",
	// History by hash.
	"log": "history", "show": "history", "edit": "history", "rm": "history",
	// Projects.
	"init": "project",
	// Daemon & configuration.
	"daemon": "system", "status": "system", "settings": "system",
	"web": "system", "auth": "system", "notify": "system",
}

func assignGroups() {
	rootCmd.AddGroup(
		&cobra.Group{ID: "session", Title: "Session commands:"},
		&cobra.Group{ID: "history", Title: "History commands:"},
		&cobra.Group{ID: "project", Title: "Project commands:"},
		&cobra.Group{ID: "system", Title: "Daemon & config commands:"},
	)
	for _, c := range rootCmd.Commands() {
		if g, ok := commandGroups[c.Name()]; ok {
			c.GroupID = g
		}
	}
	// help & completion stay ungrouped → cobra's "Additional Commands:" section.
}

// Execute runs the root cobra command.
func Execute() error {
	assignGroups()
	return rootCmd.Execute()
}

// SetVersion sets the CLI version string shown in --version output.
func SetVersion(v string) {
	rootCmd.Version = v
}
