package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/adrg/xdg"
	"pomo.local/internal/pomo"
)

// StartCommand is basically a wrapper around notify-send
type ActiveCommand struct {
	topic    string
	message  string
	duration time.Duration
	hint     string
	useToggl bool
	useEmail bool
}

func ParseActive(args []string) *ActiveCommand {
	cmd := ActiveCommand{}
	fs := flag.NewFlagSet("active", flag.ExitOnError)
	// fs.StringVar(&cmd.topic, "t", "", "Topic of your pomodoro session")
	// fs.StringVar(&cmd.message, "m", "Pomodoro session is ended!", "Notification message")
	// fs.DurationVar(&cmd.duration, "d", 25*time.Minute, "Timer duration")
	// fs.StringVar(&cmd.hint, "hint", utils.HintDefault, "Hint the same as notify-send hint")
	// fs.BoolVar(&cmd.useToggl, "toggl", false, "Use toggl integration?")
	// fs.BoolVar(&cmd.useEmail, "email", false, "Send email when the session is over?")
	fs.Parse(args)
	return &cmd
}

func (cmd *ActiveCommand) Run() error {
	path, err := xdg.StateFile("pomo/active_session.json")
	if err != nil {
		return nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No active pomodoro session")
			return nil
		}
		return err
	}

	var session pomo.Session
	if err := json.Unmarshal(data, &session); err != nil {
		return err
	}
	fmt.Printf("Active session: %s, ends at: %s\n", session.Topic, session.StopTime.Format("15:04:05"))

	return nil
}
