package cli

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/adrg/xdg"
	"pomo.local/internal/pomo"
	"pomo.local/internal/scheduler"
	"pomo.local/internal/utils"
)

// StartCommand is basically a wrapper around notify-send
type StartCommand struct {
	topic    string
	message  string
	duration time.Duration
	hint     string
	useToggl bool
	useEmail bool
}

func ParseStart(args []string) *StartCommand {
	cmd := StartCommand{}
	fs := flag.NewFlagSet("start", flag.ExitOnError)
	fs.StringVar(&cmd.topic, "t", "", "Topic of your pomodoro session")
	fs.StringVar(&cmd.message, "m", "Pomodoro session is ended!", "Notification message")
	fs.DurationVar(&cmd.duration, "d", 25*time.Minute, "Timer duration")
	fs.StringVar(&cmd.hint, "hint", utils.HintDefault, "Hint the same as notify-send hint")
	fs.BoolVar(&cmd.useToggl, "toggl", false, "Use toggl integration?")
	fs.BoolVar(&cmd.useEmail, "email", false, "Send email when the session is over?")
	fs.Parse(args)
	return &cmd
}

func (cmd *StartCommand) Run() error {
	session := pomo.NewSession(cmd.topic, cmd.duration)

	bin, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not find pomo executable: %v", err)
	}

	args := []string{
		"notify",
		"--summary", "Pomodoro",
		"--body", cmd.message,
		"--hint", cmd.hint,
	}
	if cmd.useEmail {
		args = append(args, "--email")
	}

	task := scheduler.Task{
		ID:        strconv.FormatInt(time.Now().Unix(), 16),
		ExecuteAt: session.StopTime,
		Binary:    bin,
		Args:      args,
	}

	s, err := scheduler.NewDefault()
	if err != nil {
		return err
	}

	if err := s.Schedule(task); err != nil {
		return err
	}

	// --- json task ---
	{
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

		fmt.Println("put task in: ", path)
	}

	// --- json session ---
	{
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

		fmt.Println("put session in: ", path)
	}

	// --- csv session ---
	{
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

		fmt.Println("put session in: ", sessionsPath)
	}

	return nil
}
