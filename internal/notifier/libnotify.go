package notifier

import (
	"fmt"
	"os/exec"
)

type LibnotifyNotifier struct{}

func (n *LibnotifyNotifier) Notify(title, message string, hint string) error {
	notify := exec.Command("notify-send", title, message, "--hint", hint)

	if notify.Err != nil {
		return fmt.Errorf("Error finding notify-send: %v\n", notify.Err)
	}

	if err := notify.Run(); err != nil {
		return fmt.Errorf("Error running notify: %v\n", err)
	}

	return nil
}
