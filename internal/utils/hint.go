package utils

import (
	"fmt"
)

var (
	// HintMute is the notify-send hint value that suppresses notification sounds.
	HintMute    = "boolean:suppress-sound:true"
	// HintDefault is the notify-send hint value that plays the default completion sound.
	HintDefault = "string:sound-name:complete"
)

// BuildHint returns the notify-send hint string based on sound preferences.
func BuildHint(muteNotifySound bool, notifySoundFile string) string {
	var Hint = HintDefault
	if muteNotifySound {
		Hint = HintMute
	} else if notifySoundFile != "" {
		Hint = fmt.Sprintf("string:sound-file:%s", notifySoundFile)
	}
	return Hint
}
