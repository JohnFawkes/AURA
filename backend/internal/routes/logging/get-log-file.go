package routes_logging

import (
	"aura/internal/api"
	"aura/internal/logging"
	"aura/internal/masking"
	"net/http"
	"os"
)

func GetLogFile(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Get Log File", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	// Read the log file contents
	ctx, logAction = logging.AddSubActionToContext(ctx, "Read Log File Contents", logging.LevelDebug)
	fileContent, err := os.ReadFile(logging.LogFilePath)
	if err != nil {
		logAction.SetError("Failed to read log file", "Make sure the log file exists and is readable",
			map[string]any{
				"error": err.Error(),
				"path":  logging.LogFilePath,
			})
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}

	// Convert the file into a string
	logContent := string(fileContent)

	// Redact sensitive information
	logContent = masking.Masking_IpAddress(logContent)

	logAction.Complete()
	api.Util_Response_SendJSON(w, ld, logContent)
}
