// Package config provides configuration types and loading for pomo.
package config

import (
	"github.com/adrg/xdg"
)

// Config is implemented by all pomo configuration types.
type Config interface {
	Save() error
	Load() error
}

var configDirFunc = xdg.ConfigFile
