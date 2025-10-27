package api

import (
	"aura/internal/logging"
	"os"
	"path"

	"gopkg.in/yaml.v3"
)

func (config *Config) SaveToFile() logging.StandardError {
	Err := logging.NewStandardError()

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "/config"
	}

	// Create the config.yaml file in the specified directory
	yamlFile := path.Join(configPath, "config.yaml")

	config.Logging.File = ""
	config.MediaServer.UserID = ""

	// Save the new config to a file
	data, marshalErr := yaml.Marshal(config)
	if marshalErr != nil {
		Err.Message = "Failed to marshal config to YAML"
		Err.HelpText = "Make sure the config structure is valid."
		Err.Details = map[string]any{
			"error": marshalErr.Error(),
		}
		return Err
	}
	err := os.WriteFile(yamlFile, data, 0644)
	if err != nil {
		Err.Message = "Failed to save config to file"
		Err.HelpText = "Ensure the application has write permissions to the config directory."
		Err.Details = map[string]any{
			"error": err.Error(),
		}
		return Err
	}

	return Err
}
