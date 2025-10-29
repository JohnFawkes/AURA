package api

import (
	"aura/internal/logging"
	"context"
	"fmt"
	"os"
	"time"
)

func Util_File_CheckIfFileExists(filePath string) bool {
	// Check if the file exists
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

func Util_File_CheckFolderExists(ctx context.Context, folderPath string) logging.LogErrorInfo {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Checking if folder exists", logging.LevelTrace)
	defer logAction.Complete()

	_, err := os.Stat(folderPath)
	if os.IsNotExist(err) {
		// Folder does not exist, create it
		err := os.MkdirAll(folderPath, 0755)
		if err != nil {
			logAction.SetError("Failed to create folder", fmt.Sprintf("Ensure the path %s is accessible and writable.", folderPath), map[string]any{
				"error":      err.Error(),
				"folderPath": folderPath,
			})
			return *logAction.Error
		}
	}
	return logging.LogErrorInfo{}
}

func Util_File_ClearFilesFromFolder(ctx context.Context, folderPath string, daysToClear int64) (int, logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Clearing files from folder", logging.LevelInfo)
	defer logAction.Complete()

	files, err := os.ReadDir(folderPath)
	if err != nil {
		logAction.SetError("Failed to read folder", fmt.Sprintf("Ensure the path %s is accessible.", folderPath), map[string]any{
			"error":      err.Error(),
			"folderPath": folderPath,
		})
		return 0, *logAction.Error
	}

	if len(files) == 0 {
		return 0, logging.LogErrorInfo{}
	}

	clearCount := 0
	for _, file := range files {
		if file.IsDir() {
			subFolderCount, Err := Util_File_ClearFilesFromFolder(ctx, fmt.Sprintf("%s/%s", folderPath, file.Name()), daysToClear)
			if Err.Message != "" {
				logAction.AppendWarning("message", "Failed to clear files from subfolder")
				logAction.AppendResult("subfolder", file.Name())
				logAction.AppendResult("error", Err)
				continue
			}
			clearCount += subFolderCount
			continue
		}

		filePath := fmt.Sprintf("%s/%s", folderPath, file.Name())

		// If daysToClear is set to 0, we clear all files
		if daysToClear == 0 {
			err := os.Remove(filePath)
			if err != nil {
				logAction.AppendWarning("message", "Failed to delete file")
				logAction.AppendResult("file", file.Name())
				logAction.AppendResult("error", err.Error())
				continue
			}
			clearCount++
			continue
		}

		fileInfo, err := os.Stat(filePath)
		if err != nil {
			logAction.AppendWarning("message", "Failed to get info for file")
			logAction.AppendResult("file", file.Name())
			logAction.AppendResult("error", err.Error())
			continue
		}

		// Check if the file is older than the specified number of days
		if time.Since(fileInfo.ModTime()) > time.Duration(daysToClear)*24*time.Hour {
			err := os.Remove(filePath)
			if err != nil {
				logAction.AppendWarning("message", "Failed to delete file")
				logAction.AppendResult("file", file.Name())
				logAction.AppendResult("error", err.Error())
				continue
			}
			clearCount++
		}
	}

	return clearCount, logging.LogErrorInfo{}
}
