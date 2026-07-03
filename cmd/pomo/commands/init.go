package commands

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"pomo.local/internal/project"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Mark the current directory as a pomo project",
	Long: "Creates a .pomo marker (like git's .git) in the current directory so " +
		"sessions started here are tagged with this project. Adds a self-ignoring " +
		".gitignore so it stays invisible to git.",
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(_ *cobra.Command, _ []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	p, err := project.Init(cwd)
	if errors.Is(err, project.ErrExists) {
		return fmt.Errorf("%s is already a pomo project", cwd)
	}
	if err != nil {
		return err
	}

	fmt.Printf("Initialized pomo project %q (id %s)\n", p.Name, p.ID)
	return nil
}
