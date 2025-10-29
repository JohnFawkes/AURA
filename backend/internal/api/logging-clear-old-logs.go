package api

import (
	"aura/internal/logging"
	"context"
	"os"
)

func Logging_ClearOldLogs(ctx context.Context, clearToday bool) (clearCount int, Err logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Clear Old Logs", logging.LevelDebug)
	defer logAction.Complete()
	clearCount = 0

	// Check if the log folder exists
	Err = Util_File_CheckFolderExists(ctx, logging.LogFolder)
	if Err.Message != "" {
		return 0, Err
	}

	// If clearToday is true, clear just today's log file
	if clearToday {
		// If the file exists, truncate it to zero length
		if _, err := os.Stat(logging.LogFilePath); err == nil {
			f, err := os.OpenFile(logging.LogFilePath, os.O_WRONLY|os.O_TRUNC, 0644)
			if err != nil {
				logAction.SetError("Failed to clear today's log file", "", map[string]any{
					"error": err.Error(),
					"path":  logging.LogFilePath,
				})
				return 0, *logAction.Error
			}
			// Ensure changes are flushed to disk before closing
			if err := f.Sync(); err != nil {
				logAction.SetError("Failed to sync cleared log file", "", map[string]any{
					"error": err.Error(),
					"path":  logging.LogFilePath,
				})
				f.Close()
				return 0, *logAction.Error
			}
			f.Close()
		}
		return 1, Err
	}

	clearCount, Err = Util_File_ClearFilesFromFolder(ctx, logging.LogFolder, 3)
	if Err.Message != "" {
		return clearCount, Err
	}

	return clearCount, logging.LogErrorInfo{}
}
