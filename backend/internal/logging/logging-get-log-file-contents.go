package logging

import (
	"aura/internal/masking"
	"os"
)

func Logging_GetLogFileContents() (string, StandardError) {
	Err := NewStandardError()

	filePath := GetTodayLogFile()

	if filePath == "" {
		Err.Message = "Failed to get the current log file path"
		Err.HelpText = "Ensure the logging system is properly configured and the log file path is valid."
		return "", Err
	}

	// Read the log file using os
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		Err.Message = "Failed to read the log file"
		Err.HelpText = "Ensure the log file exists and is accessible."
		Err.Details = map[string]any{
			"filePath": filePath,
			"error":    err.Error(),
		}
		return "", Err
	}

	// Convert the file content to a string
	logContent := string(fileContent)

	// Redact sensitive information from the log content
	logContent = masking.Masking_IpAddress(logContent)

	return logContent, Err
}
