package routes_config

import (
	"aura/internal/api"
	"net/http"
	"time"
)

// GetSanitizedConfig handles requests to retrieve the sanitized server configuration.
//
// Method: GET
//
// Endpoint: /api/config
//
// It responds with a JSON object containing the sanitized configuration data,
// excluding sensitive information such as passwords and API keys.
func GetSanitizedConfig(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	safeConfigData := api.Global_Config.Sanitize()

	api.Util_Response_SendJson(w, http.StatusOK, api.JSONResponse{
		Status:  "success",
		Elapsed: api.Util_ElapsedTime(startTime),
		Data:    safeConfigData,
	})
}
