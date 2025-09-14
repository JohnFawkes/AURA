package config

import (
	"aura/internal/logging"
	"aura/internal/modals"
	"os"
	"path"

	"gopkg.in/yaml.v3"
)

func UpdateConfig(newConfig modals.Config) logging.StandardError {
	Global = &newConfig
	Err := logging.NewStandardError()

	Err = SaveConfigToFile(newConfig)
	if Err.Message != "" {
		return Err
	}

	logging.SetLogLevel(newConfig.Logging.Level)

	ConfigLoaded = true
	ConfigValid = true
	ConfigMediuxValid = true
	ConfigMediaServerValid = true

	return Err
}

func SaveConfigToFile(newConfig modals.Config) logging.StandardError {
	Err := logging.NewStandardError()

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "/config"
	}

	// Create the config.yaml file in the specified directory
	yamlFile := path.Join(configPath, "config.yaml")

	newConfig.Logging.File = ""
	newConfig.MediaServer.UserID = ""

	// Save the new config to a file
	data, marshalErr := yaml.Marshal(newConfig)
	if marshalErr != nil {
		Err.Message = "Failed to marshal config to YAML"
		return Err
	}
	err := os.WriteFile(yamlFile, data, 0644)
	if err != nil {
		Err.Message = "Failed to save config to file"
		return Err
	}

	return Err
}
