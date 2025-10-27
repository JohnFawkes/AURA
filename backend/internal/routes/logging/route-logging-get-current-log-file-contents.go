package routes_logging

import (
	"aura/internal/api"
	"aura/internal/logging"
	"net/http"
	"time"
)

func GetCurrentLogFileContents(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	logging.LOG.Trace(r.URL.Path)
	Err := logging.NewStandardError()

	logContent, Err := logging.Logging_GetLogFileContents()
	if Err.Message != "" {
		logging.LOG.Warn(Err.Message)
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		return
	}

	// Respond with a success message
	api.Util_Response_SendJson(w, http.StatusOK, api.JSONResponse{
		Status:  "success",
		Elapsed: api.Util_ElapsedTime(startTime),
		Data:    logContent,
	})
}
