package routes_logging

import (
	"aura/internal/api"
	"aura/internal/logging"
	"fmt"
	"net/http"
)

func ClearLogFile(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Clear Log File", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	// Get the request body
	var ClearToday struct {
		ClearToday bool `json:"clearToday"`
	}
	Err := api.DecodeRequestBodyJSON(ctx, r.Body, &ClearToday, "ClearToday")
	if Err.Message != "" {
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}

	// Clear the log file
	clearCount, Err := api.Logging_ClearOldLogs(ctx, ClearToday.ClearToday)
	if Err.Message != "" {
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}

	if clearCount == 0 {
		api.Util_Response_SendJSON(w, ld, "No old log files to clear")
		return
	}

	api.Util_Response_SendJSON(w, ld, fmt.Sprintf("Cleared %d old log files", clearCount))
}
