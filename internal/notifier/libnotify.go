package notifier

import (
	"fmt"
	"os/exec"
)

type LibnotifyNotifier struct{}

func (n *LibnotifyNotifier) Notify(summary, body, hint string) error {
	notify := exec.Command("notify-send", summary, body, "--hint", hint)

	if notify.Err != nil {
		return fmt.Errorf("error finding notify-send: %v", notify.Err)
	}

	if err := notify.Run(); err != nil {
		return fmt.Errorf("error running notify: %v", err)
	}

	return nil
}
