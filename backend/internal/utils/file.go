package utils

import (
	"aura/internal/logging"
	"fmt"
	"os"
)

func CheckFolderExists(folderPath string) logging.StandardError {
	Err := logging.NewStandardError()
	_, err := os.Stat(folderPath)
	if os.IsNotExist(err) {
		// Folder does not exist, create it
		err := os.MkdirAll(folderPath, 0755)
		if err != nil {
			Err.Message = "Failed to create folder"
			Err.HelpText = fmt.Sprintf("Ensure the path %s is accessible and writable.", folderPath)
			Err.Details = fmt.Sprintf("Error creating folder: %v", err)
			return Err
		}
	}
	return logging.StandardError{}
}

func CheckIfImageExists(imagePath string) bool {
	// Check if the image file exists
	_, err := os.Stat(imagePath)
	return !os.IsNotExist(err)
}
