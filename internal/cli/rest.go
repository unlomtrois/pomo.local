package cli

import (
	"flag"
	"time"

	"pomo.local/internal/utils"
)

// RestCommand is basically an alias to [StartCommand]
type RestCommand struct {
	duration time.Duration
}

func ParseRest(args []string) *RestCommand {
	cmd := RestCommand{}
	fs := flag.NewFlagSet("rest", flag.ExitOnError)
	fs.DurationVar(&cmd.duration, "d", 5*time.Minute, "Timer duration")
	fs.Parse(args)
	return &cmd
}

func (cmd *RestCommand) Run() error {
	start := &StartCommand{
		topic:    "Rest",
		message:  "Break is over, get back to work!",
		duration: cmd.duration,
		hint:     utils.HintDefault,
		useToggl: false,
	}

	return start.Run()
}
