package api

import (
	"aura/internal/logging"
	"context"
	"os"
	"path"

	"gopkg.in/yaml.v3"
)

func (config *Config) SaveToFile(ctx context.Context) logging.LogErrorInfo {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Saving Config to File", logging.LevelDebug)
	defer logAction.Complete()

	// Sub-action: Determine config path
	subActionDeterminePath := logAction.AddSubAction("Determine Config Path", logging.LevelTrace)
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "/config"
	}
	yamlFile := path.Join(configPath, "config.yaml")
	config.MediaServer.UserID = ""
	subActionDeterminePath.Complete()

	// Sub-action: Marshal config to YAML
	subActionMarshal := logAction.AddSubAction("Marshal Config to YAML", logging.LevelTrace)
	data, marshalErr := yaml.Marshal(config)
	if marshalErr != nil {
		subActionMarshal.SetError("Failed to marshal config to YAML", marshalErr.Error(), nil)
		logAction.Status = logging.StatusError
		return *subActionMarshal.Error
	}
	subActionMarshal.Complete()

	// Sub-action: Write config to file
	subActionWrite := logAction.AddSubAction("Write Config to File", logging.LevelTrace)
	if writeErr := os.WriteFile(yamlFile, data, 0644); writeErr != nil {
		subActionWrite.SetError("Failed to write config to file", writeErr.Error(), nil)
		logAction.Status = logging.StatusError
		return *subActionWrite.Error
	}
	subActionWrite.Complete()

	return logging.LogErrorInfo{}
}
