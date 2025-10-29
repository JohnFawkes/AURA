package api

import (
	"aura/internal/logging"
	"context"
	"fmt"
	"os"
	"path"

	"gopkg.in/yaml.v3"
)

func Config_LoadYamlConfig(ctx context.Context) {
	ctx, ld := logging.CreateLoggingContext(ctx, "Config - Load YAML Config")

	// Top-level action
	action := ld.AddAction("Load YAML config", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, action)
	defer action.Complete()

	// Use an environment variable to determine the config path
	// By default, it will use /config
	// This is useful for testing and local development
	// In Docker, the config path is set to /config
	// Sub-action: Check for config.yaml file
	ctx, sub := logging.AddSubActionToContext(ctx, "Check for config.yaml file", logging.LevelTrace)
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "/config"
	}

	yamlFile := path.Join(configPath, "config.yaml")
	if _, err := os.Stat(yamlFile); os.IsNotExist(err) {
		yamlFile = path.Join(configPath, "config.yml")
		if _, err := os.Stat(yamlFile); os.IsNotExist(err) {
			sub.SetError(fmt.Sprintf("Config file not found in '%s'", configPath), err.Error(), nil)
			ld.Status = logging.StatusError
			return
		}
	}
	sub.Complete()

	// Sub-action: Read config.yaml file
	ctx, sub = logging.AddSubActionToContext(ctx, "Read config.yaml file", logging.LevelTrace)
	data, err := os.ReadFile(yamlFile)
	if err != nil {
		sub.SetError(fmt.Sprintf("failed to read config.yaml file: %s", err.Error()), err.Error(), nil)
		ld.Status = logging.StatusError
		return
	}
	sub.Complete()

	// Sub-action: Parse config.yaml file
	var config Config
	ctx, sub = logging.AddSubActionToContext(ctx, "Parse config.yaml file", logging.LevelTrace)
	if err := yaml.Unmarshal(data, &config); err != nil {
		sub.SetError(fmt.Sprintf("failed to parse config.yaml file: %s", err.Error()), err.Error(), nil)
		ld.Status = logging.StatusError
		return
	}
	sub.Complete()

	Global_Config = &config
	Global_Config_Loaded = true
}
