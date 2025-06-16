package health

import (
	"aura/internal/logging"
	"aura/internal/utils"
	"fmt"
	"net/http"
	"os"
	"time"
)

func ClearLogOldFiles(w http.ResponseWriter, r *http.Request) {
	logging.LOG.Trace(r.URL.Path)
	startTime := time.Now()
	Err := logging.NewStandardError()

	// Check if the log folder exists
	Err = utils.CheckFolderExists(logging.LogFolder)
	if Err.Message != "" {
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}

	// Go through the log files and remove those older than 7 days
	files, _ := os.ReadDir(logging.LogFolder)
	if len(files) == 0 {
		logging.LOG.Warn(fmt.Sprintf("No log files found in %s", logging.LogFolder))
		utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
			Status:  "success",
			Elapsed: utils.ElapsedTime(startTime),
			Data:    "No log files to clear",
		})
		return
	}

	clearCount := 0
	for _, file := range files {
		if file.IsDir() {
			continue // Skip directories
		}
		filePath := fmt.Sprintf("%s/%s", logging.LogFolder, file.Name())
		logging.LOG.Debug(fmt.Sprintf("Checking file: %s", filePath))
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			logging.LOG.Error(fmt.Sprintf("Error getting file info for %s: %v", filePath, err))
			continue // Skip this file if there's an error
		}
		// Check if the file is older than 7 days
		if time.Since(fileInfo.ModTime()) > 7*24*time.Hour {
			logging.LOG.Debug(fmt.Sprintf("Removing old log file: %s", filePath))
			if err := os.Remove(filePath); err != nil {
				logging.LOG.Error(fmt.Sprintf("Error removing file %s: %v", filePath, err))
				continue // Skip incrementing clearCount if there's an error
			}
			clearCount++
		}
	}

	if clearCount == 0 {
		utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
			Status:  "success",
			Elapsed: utils.ElapsedTime(startTime),
			Data:    "No old log files to clear",
		})
		return
	}

	// Return a JSON response
	utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
		Status:  "success",
		Elapsed: utils.ElapsedTime(startTime),
		Data:    fmt.Sprintf("Cleared %d old log files successfully", clearCount),
	})
}
