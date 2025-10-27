package api

import (
	"aura/internal/logging"
	"fmt"
	"os"
	"path"

	"gopkg.in/yaml.v3"
)

func Config_LoadYamlConfig() {
	logging.LOG.Debug("Attempting to load configuration from YAML file...")

	// Use an environment variable to determine the config path
	// By default, it will use /config
	// This is useful for testing and local development
	// In Docker, the config path is set to /config
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "/config"
	}

	// Check for a config.yaml or config.yml file
	yamlFile := path.Join(configPath, "config.yaml")
	if _, err := os.Stat(yamlFile); os.IsNotExist(err) {
		yamlFile = path.Join(configPath, "config.yml")
		if _, err := os.Stat(yamlFile); os.IsNotExist(err) {
			logging.LOG.Error(fmt.Sprintf("Config file not found in '%s'", configPath))
			return
		}
	}

	// Read the YAML file
	data, err := os.ReadFile(yamlFile)
	if err != nil {
		logging.LOG.Error(fmt.Sprintf("failed to read config.yaml file: %s", err.Error()))
		return
	}

	// Parse the YAML file into a Config struct
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		logging.LOG.Error(fmt.Sprintf("failed to parse config.yaml file: %s", err.Error()))
		return
	}

	Global_Config = &config
	Global_Config_Loaded = true
}
