package commands

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/adrg/xdg"
	"github.com/spf13/cobra"
	"pomo.local/internal/pomo"
	"pomo.local/internal/scheduler"
	"pomo.local/internal/utils"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Set a new pomodoro timer",
	RunE:  runStart,
}

func init() {
	rootCmd.AddCommand(startCmd)

	startCmd.Flags().StringP("topic", "t", "", "Topic of your pomodoro session")
	startCmd.Flags().StringP("message", "m", "Pomodoro session is ended!", "Notification message")
	startCmd.Flags().DurationP("duration", "d", 25*time.Minute, "Timer duration")
	startCmd.Flags().String("hint", utils.HintDefault, "Hint the same as notify-send hint")
	startCmd.Flags().Bool("toggl", false, "Use toggl integration")
	startCmd.Flags().Bool("email", false, "Send email when the session is over")
	startCmd.Flags().BoolP("verbose", "v", false, "Verbose output")
}

func runStart(cmd *cobra.Command, _ []string) error {
	verbose, _ := cmd.Flags().GetBool("verbose")
	if verbose {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	topic, _ := cmd.Flags().GetString("topic")
	message, _ := cmd.Flags().GetString("message")
	duration, _ := cmd.Flags().GetDuration("duration")
	hint, _ := cmd.Flags().GetString("hint")
	useEmail, _ := cmd.Flags().GetBool("email")

	return executeStart(topic, message, duration, hint, useEmail, verbose)
}

func executeStart(topic, message string, duration time.Duration, hint string, useEmail, verbose bool) error {
	// check that there is no current session
	if err := checkPomodoroSession(); errors.Is(err, fs.ErrExist) {
		return fmt.Errorf("You can only have 1 active pomodoro session at once.")
	}

	session := pomo.NewSession(topic, duration)
	slog.Debug("Prepared session:", "session", session)

	bin, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not find pomo executable: %v", err)
	}

	notifyArgs := []string{
		"notify",
		"--summary", "Pomodoro",
		"--body", message,
		"--hint", hint,
	}
	if useEmail {
		notifyArgs = append(notifyArgs, "--email")
	}

	task := scheduler.Task{
		ID:        strconv.FormatInt(time.Now().Unix(), 16),
		ExecuteAt: session.StopTime,
		Binary:    bin,
		Args:      notifyArgs,
	}
	slog.Debug("Prepared task:", "task", task)

	s, err := scheduler.NewDefault(verbose)
	if err != nil {
		return err
	}

	if err := s.Schedule(task); err != nil {
		return err
	}

	if err := saveActiveTask(task); err != nil {
		return err
	}

	if err := saveSession(session); err != nil {
		return err
	}

	if err := appendSessionCsv(session); err != nil {
		return err
	}

	return nil
}

func checkPomodoroSession() error {
	slog.Debug("Search for active_session:", "path", "pomo/active_session.json")
	if _, err := xdg.SearchStateFile("pomo/active_session.json"); err != nil {
		slog.Debug("File %s not found: %w", "pomo/active_session.json", err)
		return fs.ErrNotExist
	}
	return fs.ErrExist
}

func saveActiveTask(task scheduler.Task) error {
	path, err := xdg.StateFile("pomo/active_task.json")
	if err != nil {
		return nil
	}

	data, err := json.MarshalIndent(task, "", "    ")
	if err != nil {
		return nil
	}

	if err = os.WriteFile(path, data, 0644); err != nil {
		return err
	}

	slog.Debug("Updated active_task", "path", path)
	return nil
}

func saveSession(session *pomo.Session) error {
	path, err := xdg.StateFile("pomo/active_session.json")
	if err != nil {
		return nil
	}

	data, err := json.MarshalIndent(session, "", "    ")
	if err != nil {
		return nil
	}

	if err = os.WriteFile(path, data, 0644); err != nil {
		return err
	}

	slog.Debug("Updated active_session", "path", path)
	return nil
}

func appendSessionCsv(session *pomo.Session) error {
	sessionsPath, err := xdg.DataFile("pomo/sessions.csv")
	if err != nil {
		return err
	}

	file, err := os.OpenFile(sessionsPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.Write(session.Strings()); err != nil {
		return err
	}

	slog.Debug("Updated session.csv", "path", sessionsPath)
	return nil
}
