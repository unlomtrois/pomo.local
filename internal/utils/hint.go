package utils

import (
	"fmt"
)

var (
	HintMute    = "boolean:suppress-sound:true"
	HintDefault = "string:sound-name:complete"
)

func BuildHint(muteNotifySound bool, notifySoundFile string) string {
	var Hint = HintDefault
	if muteNotifySound {
		Hint = HintMute
	} else if notifySoundFile != "" {
		Hint = fmt.Sprintf("string:sound-file:%s", notifySoundFile)
	}
	return Hint
}
