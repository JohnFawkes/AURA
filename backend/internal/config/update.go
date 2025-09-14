package config

import (
	"aura/internal/logging"
	"aura/internal/modals"
	"encoding/json"
	"fmt"
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

	jsonConfig, err := json.MarshalIndent(newConfig, "", "  ")
	if err != nil {
		logging.LOG.Error(fmt.Sprintf("Failed to marshal config to JSON: %v", err))
	} else {
		logging.LOG.Info(fmt.Sprintf("Saving New Config (JSON): %s", string(jsonConfig)))
	}

	logging.SetLogLevel(newConfig.Logging.Level)

	ConfigLoaded = true
	ConfigValid = true
	return Err
}

func SaveConfigToFile(newConfig modals.Config) logging.StandardError {
	Err := logging.NewStandardError()

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "/config"
	}

	// Check for a config.yml or config.yaml file
	yamlFile := path.Join(configPath, "config.yml")

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
