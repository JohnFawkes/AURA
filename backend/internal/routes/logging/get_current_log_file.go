package route_logging

import (
	"aura/internal/logging"
	"aura/internal/utils"
	"aura/internal/utils/masking"
	"net/http"
	"os"
	"time"
)

func GetCurrentLogFile(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	logging.LOG.Trace(r.URL.Path)
	Err := logging.NewStandardError()

	filePath := logging.GetTodayLogFile()

	if filePath == "" {
		Err.Message = "Failed to get the current log file path"
		Err.HelpText = "Ensure the logging system is properly configured and the log file path is valid."
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
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
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}

	// Convert the file content to a string
	logContent := string(fileContent)

	// Redact sensitive information from the log content
	logContent = masking.RedactIPAddresses(logContent)

	utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
		Status:  "success",
		Elapsed: utils.ElapsedTime(startTime),
		Data:    logContent,
	})
}
