package routes_logging

import (
	"aura/internal/api"
	"aura/internal/logging"
	"encoding/json"
	"fmt"

	"net/http"
	"time"
)

func ClearOldLogs(w http.ResponseWriter, r *http.Request) {
	logging.LOG.Trace(r.URL.Path)
	startTime := time.Now()
	Err := logging.NewStandardError()

	// Get the request body
	var ClearToday struct {
		ClearToday bool `json:"clearToday"`
	}

	decodeErr := json.NewDecoder(r.Body).Decode(&ClearToday)
	if decodeErr != nil {
		Err.Message = "Failed to decode request body"
		Err.Details = map[string]any{
			"error":      decodeErr.Error(),
			"clearToday": ClearToday.ClearToday,
		}
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		return
	}

	clearCount, Err := api.Logging_ClearOldLogs(ClearToday.ClearToday)
	if Err.Message != "" {
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		return
	}

	if clearCount == 0 {
		api.Util_Response_SendJson(w, http.StatusOK, api.JSONResponse{
			Status:  "success",
			Elapsed: api.Util_ElapsedTime(startTime),
			Data:    "No old log files to clear",
		})
		return
	}

	logging.LOG.Info(fmt.Sprintf("Cleared %d old log files from %s", clearCount, logging.LogFolder))

	// Return a JSON response
	api.Util_Response_SendJson(w, http.StatusOK, api.JSONResponse{
		Status:  "success",
		Elapsed: api.Util_ElapsedTime(startTime),
		Data:    fmt.Sprintf("Cleared %d old log files successfully", clearCount),
	})
}
