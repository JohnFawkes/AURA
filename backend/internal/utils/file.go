package utils

import (
	"aura/internal/logging"
	"fmt"
	"os"
	"time"
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
			Err.Details = map[string]any{
				"error":      err.Error(),
				"folderPath": folderPath,
			}
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

func ClearFilesFromFolder(folderPath string, daysToClear int64) (int, logging.StandardError) {
	Err := logging.NewStandardError()
	files, err := os.ReadDir(folderPath)
	if err != nil {
		Err.Message = "Failed to read folder"
		Err.HelpText = fmt.Sprintf("Ensure the path '%s' is accessible.", folderPath)
		Err.Details = map[string]any{
			"error":      err.Error(),
			"folderPath": folderPath,
		}
		return 0, Err
	}

	logging.LOG.Trace(fmt.Sprintf("Checking files in folder: %s", folderPath))

	if len(files) == 0 {
		logging.LOG.Warn(fmt.Sprintf("No files found in '%s'", folderPath))
		return 0, logging.StandardError{}
	}

	clearCount := 0
	for _, file := range files {
		if file.IsDir() {
			subFolderCount, Err := ClearFilesFromFolder(fmt.Sprintf("%s/%s", folderPath, file.Name()), daysToClear)
			if Err.Message != "" {
				logging.LOG.Error(fmt.Sprintf("Error clearing subfolder %s: %v", file.Name(), Err))
				continue // Skip this subfolder if there's an error
			}
			clearCount += subFolderCount
			continue // Skip directories, we already handled them
		}

		filePath := fmt.Sprintf("%s/%s", folderPath, file.Name())

		// If the daysToClear is 0, we assume we want to clear all files
		if daysToClear == 0 {
			logging.LOG.Debug(fmt.Sprintf("Removing file: %s", file.Name()))
			if err := os.Remove(filePath); err != nil {
				logging.LOG.Error(fmt.Sprintf("Error removing file %s: %v", filePath, err))
				continue // Skip incrementing clearCount if there's an error
			}
			clearCount++
			continue // Skip further checks for this file
		}

		fileInfo, err := os.Stat(filePath)
		if err != nil {
			logging.LOG.Error(fmt.Sprintf("Error getting file info for %s: %v", file.Name(), err))
			continue // Skip this file if there's an error
		}
		// Check if the file is older than the specified number of days
		if time.Since(fileInfo.ModTime()) > time.Duration(daysToClear)*24*time.Hour {
			logging.LOG.Debug(fmt.Sprintf("Removing old file: %s", file.Name()))
			if err := os.Remove(filePath); err != nil {
				logging.LOG.Error(fmt.Sprintf("Error removing file %s: %v", file.Name(), err))
				continue // Skip incrementing clearCount if there's an error
			}
			clearCount++
		}
	}

	return clearCount, logging.StandardError{}
}
