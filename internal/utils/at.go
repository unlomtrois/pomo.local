package utils

import "os/exec"

func HasAt() bool {
	_, err := exec.LookPath("at")
	return err == nil
}
