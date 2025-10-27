package routes_mediux

import (
	"aura/internal/api"
	"aura/internal/logging"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

func ValidateToken(w http.ResponseWriter, r *http.Request) {
	logging.LOG.Trace(r.URL.Path)
	startTime := time.Now()
	Err := logging.NewStandardError()

	// Get the media server information from the request
	var mediuxInfo api.Config_Mediux
	if err := json.NewDecoder(r.Body).Decode(&mediuxInfo); err != nil {
		Err.Message = "Failed to decode request body"
		Err.HelpText = "Ensure the request body is valid JSON"
		Err.Details = map[string]any{
			"error": err.Error(),
		}
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		return
	}

	// Store the current values for Token
	currentToken := ""
	if api.Global_Config.Mediux.Token != "" {
		currentToken = api.Global_Config.Mediux.Token
	}

	// Change the Global values temporarily
	if !strings.HasPrefix(mediuxInfo.Token, "***") {
		api.Global_Config.Mediux.Token = mediuxInfo.Token
	}

	// Restore the previous values
	defer func() {
		api.Global_Config.Mediux.Token = currentToken
	}()

	Err = api.Mediux_ValidateToken(mediuxInfo.Token)
	if Err.Message != "" {
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		// Restore the previous values
		api.Global_Config.Mediux.Token = currentToken
		return
	}

	// Respond with a success message
	api.Util_Response_SendJson(w, http.StatusOK, api.JSONResponse{
		Status:  "success",
		Elapsed: api.Util_ElapsedTime(startTime),
		Data:    "New MediUX token is ok",
	})
}
