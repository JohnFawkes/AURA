package api

import (
	"aura/internal/logging"
	"fmt"
	"os"
	"path"
)

func DB_Migration_CreateWarningFile(version int, warning logging.StandardError) {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "/config"
	}
	filePath := path.Join(configPath, fmt.Sprintf("migration_warning_v%d.txt", version))

	// Create the file if it doesn't exist, else append to it
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logging.LOG.Error(fmt.Sprintf("Failed to create/open migration warning file: %v", err))
		return
	}
	defer file.Close()

	// Write the warning details to the file
	content := fmt.Sprintf("Message: %s\nItem: %s\nDetails: %v\n", warning.Message, warning.HelpText, warning.Details)
	logging.LOG.Warn(fmt.Sprintf("Migration Warning (v%d): %s - Details: %v", version, warning.Message, warning.Details))

	if _, err := file.WriteString(content + "\n"); err != nil {
		logging.LOG.Error(fmt.Sprintf("Failed to write to migration warning file: %v", err))
		return
	}
}
