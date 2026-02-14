package utils

import (
	"os"
)

func HasSystemd() bool {
	_, err := os.Stat("/run/systemd/system")
	return err == nil
}
