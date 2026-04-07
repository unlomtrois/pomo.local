// Package notifier provides desktop notification implementations.
package notifier

import (
	"fmt"
	"os/exec"
)

// LibnotifyNotifier sends desktop notifications via notify-send.
type LibnotifyNotifier struct{}

// Notify sends a desktop notification with the given summary, body, and hint.
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
