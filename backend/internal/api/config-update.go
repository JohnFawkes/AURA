package api

import (
	"aura/internal/logging"
	"context"
)

func (config *Config) Update(ctx context.Context) logging.LogErrorInfo {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Updating Config", logging.LevelInfo)
	defer logAction.Complete()

	// Sub-action: Set global config
	subActionSetGlobalConfig := logAction.AddSubAction("Set Global Config", logging.LevelTrace)
	Global_Config = config
	subActionSetGlobalConfig.Complete()

	// Sub-action: Save config to file
	Err := config.SaveToFile(ctx)
	if Err.Message != "" {
		logAction.SetError("Failed to save config to file", Err.Message, nil)
		logAction.Status = logging.StatusError
		return Err
	}

	// Sub-action: Set log level
	subActionSetLogLevel := logAction.AddSubAction("Set Log Level", logging.LevelTrace)
	logging.SetLogLevel(config.Logging.Level)
	subActionSetLogLevel.Complete()

	// Sub-action: Set global flags
	subActionSetGlobalFlags := logAction.AddSubAction("Set Global Flags", logging.LevelTrace)
	Global_Config_Loaded = true
	Global_Config_Valid = true
	Global_Config_MediuxValid = true
	Global_Config_MediaServerValid = true
	subActionSetGlobalFlags.Complete()

	return logging.LogErrorInfo{}
}
