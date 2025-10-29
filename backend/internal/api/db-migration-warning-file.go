package api

import (
	"aura/internal/logging"
	"fmt"
	"os"
	"path"
)

func DB_Migration_CreateWarningFile(version int, warning logging.LogErrorInfo) {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "/config"
	}
	filePath := path.Join(configPath, fmt.Sprintf("migration_warning_v%d.txt", version))

	// Create the file if it doesn't exist, else append to it
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logging.LOGGER.Error().Timestamp().Err(err).Msg("Failed to create or open migration warning file")
		return
	}
	defer file.Close()

	// Write the warning details to the file
	content := fmt.Sprintf("Message: %s\nItem: %s\nDetails: %v\n", warning.Message, warning.Help, warning.Detail)
	logging.LOGGER.Warn().Timestamp().Int("version", version).Str("filePath", filePath).Interface("content", content).Msg("Database migration warning created")

	if _, err := file.WriteString(content + "\n"); err != nil {
		logging.LOGGER.Error().Timestamp().Err(err).Msg("Failed to write to migration warning file")
	}
}
