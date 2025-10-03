package route_logging

import (
	"aura/internal/logging"
	"aura/internal/utils"
	"encoding/json"
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

	// Get the request body
	var ClearToday struct {
		ClearToday bool `json:"clearToday"`
	}

	// Decode the request body into the struct
	err := json.NewDecoder(r.Body).Decode(&ClearToday)
	if err != nil {
		Err := logging.NewStandardError()
		Err.Message = "Failed to decode request body"
		Err.HelpText = "Ensure the request body is a valid JSON object matching the expected structure."
		Err.Details = map[string]any{
			"error": err.Error(),
		}
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}

	if ClearToday.ClearToday {
		logging.LOG.Trace("Clearing today's log files")

		// Get today's log file
		todaysLogFile := logging.GetTodayLogFile()

		// If the log file does not exist, create it
		// If the log file exists, clear it out
		if _, err := os.Stat(todaysLogFile); os.IsNotExist(err) {
			// Create the log file if it does not exist
			file, err := os.Create(todaysLogFile)
			if err != nil {
				Err.Message = "Failed to create today's log file"
				Err.HelpText = "Check file permissions and ensure the log directory is writable."
				Err.Details = map[string]any{
					"filePath": todaysLogFile,
					"error":    err.Error(),
				}
				utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
				return
			}
			defer file.Close()
			logging.LOG.Info(fmt.Sprintf("Created today's log file: %s", todaysLogFile))
		} else {
			// Clear the log file if it exists
			err := os.Truncate(todaysLogFile, 0)
			if err != nil {
				Err.Message = "Failed to clear today's log file"
				Err.HelpText = "Check file permissions and ensure the log directory is writable."
				Err.Details = map[string]any{
					"filePath": todaysLogFile,
					"error":    err.Error(),
				}
				utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
				return
			}
			logging.LOG.Info(fmt.Sprintf("Cleared today's log file: %s", todaysLogFile))
		}
		utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
			Status:  "success",
			Elapsed: utils.ElapsedTime(startTime),
			Data:    "Today's log file cleared successfully",
		})
		return
	}

	clearCount, Err := utils.ClearFilesFromFolder(logging.LogFolder, 3)
	if Err.Message != "" {
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}

	if clearCount == 0 {
		utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
			Status:  "success",
			Elapsed: utils.ElapsedTime(startTime),
			Data:    "No old log files to clear",
		})
		return
	}

	logging.LOG.Info(fmt.Sprintf("Cleared %d old log files from %s", clearCount, logging.LogFolder))

	// Return a JSON response
	utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
		Status:  "success",
		Elapsed: utils.ElapsedTime(startTime),
		Data:    fmt.Sprintf("Cleared %d old log files successfully", clearCount),
	})
}
