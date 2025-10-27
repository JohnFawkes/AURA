package api

import (
	"aura/internal/logging"
	"fmt"
	"os"
)

func Logging_ClearOldLogs(clearToday bool) (int, logging.StandardError) {

	clearCount := 0
	// Check if the log folder exists
	Err := Util_File_CheckFolderExists(logging.LogFolder)
	if Err.Message != "" {
		return 0, Err
	}

	// If clearToday is true, clear just today's log file
	if clearToday {
		logging.LOG.Trace("Clearing today's log file")
		todayLogFile := logging.GetTodayLogFile()
		// If the file exists, clear it
		if _, err := os.Stat(todayLogFile); err == nil {
			err = os.Remove(todayLogFile)
			if err != nil {
				Err.Message = fmt.Sprintf("Failed to clear today's log file: %s", err.Error())
				return 0, Err
			}
		}
		return 1, Err
	}

	clearCount, Err = Util_File_ClearFilesFromFolder(logging.LogFolder, 3)
	if Err.Message != "" {
		return clearCount, Err
	}

	return clearCount, Err
}
