package commands

import (
	"github.com/spf13/cobra"
	"pomo.local/internal/config"
	"pomo.local/internal/utils"
)

var restCmd = &cobra.Command{
	Use:   "rest",
	Short: "Start a rest/break timer",
	RunE:  runRest,
}

func init() {
	rootCmd.AddCommand(restCmd)

	restCmd.Flags().DurationP("duration", "d", config.LoadSettings().RestDuration.Std(), "Timer duration")
	restCmd.Flags().Bool("email", false, "Send email when the session is over")
}

func runRest(cmd *cobra.Command, _ []string) error {
	duration, _ := cmd.Flags().GetDuration("duration")
	useEmail, _ := cmd.Flags().GetBool("email")

	return executeStart("Rest", "Break is over, get back to work!", duration, utils.HintDefault, useEmail)
}
