package config

import "pomo.local/internal/utils"

type Config interface {
	Save() error
	Load() error
}

var configDirFunc = utils.GetConfigDir
