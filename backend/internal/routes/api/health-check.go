package routes_api

import (
	"aura/internal/api"
	"net/http"
	"time"
)

// HealthCheck handles HTTP requests for checking the server's health status.
//
// Method: GET
// Endpoints:
//   - /
//   - /api
//   - /api/health
//
// It responds with a JSON object containing the server status, elapsed time, and a message.
// This endpoint can be used for uptime monitoring and basic diagnostics.
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	// Get the start time
	startTime := time.Now()

	api.Util_Response_SendJson(w, http.StatusOK, api.JSONResponse{
		Status:  "success",
		Elapsed: api.Util_ElapsedTime(startTime),
		Data:    "Server is online",
	})
}
