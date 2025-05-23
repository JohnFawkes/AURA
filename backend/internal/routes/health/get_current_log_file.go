package health

import (
	"aura/internal/logging"
	"aura/internal/utils"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"time"
)

func GetCurrentLogFile(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	logging.LOG.Trace(r.URL.Path)

	filePath := logging.GetTodayLogFile()

	if filePath == "" {
		logging.LOG.Error("Failed to get the current log file path")
		utils.SendErrorJSONResponse(w, http.StatusInternalServerError, logging.ErrorLog{
			Err: fmt.Errorf("failed to get the current log file path"),
			Log: logging.Log{
				Message: "Failed to get the current log file path",
				Elapsed: utils.ElapsedTime(startTime),
			},
		})
		return
	}

	// Read the log file using os
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		utils.SendErrorJSONResponse(w, http.StatusInternalServerError, logging.ErrorLog{
			Err: err,
			Log: logging.Log{
				Message: "Failed to read the log file",
				Elapsed: utils.ElapsedTime(startTime),
			},
		})
		return
	}

	// Convert the file content to a string
	logContent := string(fileContent)

	// Redact sensitive information from the log content
	logContent = redactIPAddresses(logContent)

	utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
		Status:  "success",
		Message: "Current log file content retrieved successfully",
		Elapsed: utils.ElapsedTime(startTime),
		Data:    logContent,
	})
}

func redactIPAddresses(logContent string) string {
	patterns := map[string]string{
		`\b\d{1,3}(\.\d{1,3}){3}\b`: "***REDACTED_IP***", // IP addresses
	}

	for pattern, replacement := range patterns {
		re := regexp.MustCompile(pattern)
		logContent = re.ReplaceAllString(logContent, replacement)
	}
	return logContent
}
