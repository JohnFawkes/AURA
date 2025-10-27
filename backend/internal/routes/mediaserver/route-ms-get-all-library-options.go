package routes_ms

import (
	"aura/internal/api"
	"aura/internal/logging"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

func GetAllLibrariesOptions(w http.ResponseWriter, r *http.Request) {
	logging.LOG.Trace(r.URL.Path)
	startTime := time.Now()
	Err := logging.NewStandardError()

	// Get the media server information from the request
	var mediaServerInfo api.Config_MediaServer
	if err := json.NewDecoder(r.Body).Decode(&mediaServerInfo); err != nil {
		Err.Message = "Failed to decode request body"
		Err.HelpText = "Ensure the request body is valid JSON"
		Err.Details = map[string]any{
			"error": err.Error(),
		}
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		return
	}

	// Store the current values for Type, URL and Token
	currentType := api.Global_Config.MediaServer.Type
	currentURL := api.Global_Config.MediaServer.URL
	currentToken := api.Global_Config.MediaServer.Token
	currentLibraries := api.Global_Config.MediaServer.Libraries
	currentUserID := api.Global_Config.MediaServer.UserID

	// Change the Global values temporarily
	api.Global_Config.MediaServer.Type = mediaServerInfo.Type
	api.Global_Config.MediaServer.URL = mediaServerInfo.URL
	api.Global_Config.MediaServer.Libraries = mediaServerInfo.Libraries
	api.Global_Config.MediaServer.UserID = mediaServerInfo.UserID
	if !strings.HasPrefix(mediaServerInfo.Token, "***") {
		api.Global_Config.MediaServer.Token = mediaServerInfo.Token
	}

	// Restore the previous values
	defer func() {
		api.Global_Config.MediaServer.Type = currentType
		api.Global_Config.MediaServer.URL = currentURL
		api.Global_Config.MediaServer.Token = currentToken
		api.Global_Config.MediaServer.Libraries = currentLibraries
		api.Global_Config.MediaServer.UserID = currentUserID
	}()

	allSections, Err := api.CallFetchLibrarySectionOptions()
	if Err.Message != "" {
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		return
	}

	if len(allSections) == 0 {
		Err.Message = "No sections found"
		Err.HelpText = fmt.Sprintf("Ensure that the media server has sections configured for %s.", api.Global_Config.MediaServer.Type)
		Err.Details = map[string]any{
			"error": fmt.Sprintf("No sections found in %s for the configured libraries", api.Global_Config.MediaServer.Type),
		}
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		return
	}

	// Respond with a success message
	api.Util_Response_SendJson(w, http.StatusOK, api.JSONResponse{
		Status:  "success",
		Elapsed: api.Util_ElapsedTime(startTime),
		Data:    allSections,
	})
}
