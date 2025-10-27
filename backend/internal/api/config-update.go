package api

import (
	"aura/internal/logging"
)

func (config *Config) Config_UpdateConfig() logging.StandardError {
	Global_Config = config
	Err := logging.NewStandardError()

	Err = config.SaveToFile()
	if Err.Message != "" {
		return Err
	}

	logging.SetLogLevel(config.Logging.Level)

	Global_Config_Loaded = true
	Global_Config_Valid = true
	Global_Config_MediuxValid = true
	Global_Config_MediaServerValid = true

	return Err
}
