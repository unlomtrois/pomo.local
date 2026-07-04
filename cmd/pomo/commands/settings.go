package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"pomo.local/internal/config"
	"pomo.local/internal/utils"
)

var settingsCmd = &cobra.Command{
	Use:   "settings",
	Short: "Show pomo settings (durations); --init writes a default file to edit",
	RunE:  runSettings,
}

func init() {
	rootCmd.AddCommand(settingsCmd)

	settingsCmd.Flags().Bool("init", false, "Write a default settings.json if none exists")
}

func runSettings(cmd *cobra.Command, _ []string) error {
	path := config.SettingsPath()

	if doInit, _ := cmd.Flags().GetBool("init"); doInit {
		if config.SettingsExist() {
			fmt.Printf("settings already exist at %s\n", path)
		} else if err := config.DefaultSettings().Save(); err != nil {
			return err
		} else {
			fmt.Printf("wrote default settings to %s\n", path)
		}
	}

	s := config.LoadSettings()
	source := "defaults (no file — run `pomo settings --init` to customize)"
	if config.SettingsExist() {
		source = path
	}
	fmt.Printf("settings  %s\n", source)
	fmt.Printf("doro      %s\n", utils.ShortDuration(s.DoroDuration.Std()))
	fmt.Printf("long      %s\n", utils.ShortDuration(s.LongDuration.Std()))
	fmt.Printf("rest      %s\n", utils.ShortDuration(s.RestDuration.Std()))
	return nil
}
