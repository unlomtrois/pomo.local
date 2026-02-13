package utils

import (
	"fmt"
	"os/exec"
)

func Notify(summary string, body string, hint string) error {
	cmd := exec.Command("notify-send", summary, body, "--hint", hint)
	if cmd.Err != nil {
		return fmt.Errorf("Error finding notify-send: %v\n", cmd.Err)
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Error running notify: %v\n", err)
	}

	return nil
}
