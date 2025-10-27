package routes_api

import (
	"aura/internal/api"
	"aura/internal/logging"
	"net/http"
	"time"
)

// NotFound handles requests to undefined routes.
//
// Method: ANY
//
// Endpoint: Any undefined route
//
// It returns a JSON response indicating that the route was not found.
func NotFound(w http.ResponseWriter, r *http.Request) {
	logging.LOG.Warn("Route not found: " + r.URL.Path)
	startTime := time.Now()

	api.Util_Response_SendJson(w, http.StatusNotFound, api.JSONResponse{
		Status:  "error",
		Elapsed: api.Util_ElapsedTime(startTime),
		Data:    "Route not found",
	})
}
