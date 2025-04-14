package utils

import (
	"fmt"
	"os"
	"poster-setter/internal/logging"
)

func CheckFolderExists(folderPath string) logging.ErrorLog {
	_, err := os.Stat(folderPath)
	if os.IsNotExist(err) {
		// Folder does not exist, create it
		err := os.MkdirAll(folderPath, 0755)
		if err != nil {
			return logging.ErrorLog{Err: err,
				Log: logging.Log{Message: fmt.Sprintf("Failed to create folder: %s", folderPath)},
			}
		}
	}
	return logging.ErrorLog{}
}

func CheckIfImageExists(imagePath string) bool {
	// Check if the image file exists
	_, err := os.Stat(imagePath)
	if os.IsNotExist(err) {
		return false
	}
	return true
}
