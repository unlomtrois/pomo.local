package config

import (
	"github.com/adrg/xdg"
)

type Config interface {
	Save() error
	Load() error
}

var configDirFunc = xdg.ConfigFile
