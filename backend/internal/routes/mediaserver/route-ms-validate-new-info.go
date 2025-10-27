package routes_ms

import (
	"aura/internal/api"
	"aura/internal/logging"
	"aura/internal/masking"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

func ValidateNewInfo(w http.ResponseWriter, r *http.Request) {
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

	// If the token is masked, get the actual token from the global config
	if !strings.HasPrefix(mediaServerInfo.Token, "***") {
		mediaServerInfo.Token = api.Global_Config.MediaServer.Token
	}

	mediaServer, Err := api.GetMediaServerInterface(mediaServerInfo)
	if Err.Message != "" {
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		return
	}

	_, Err = mediaServer.GetMediaServerStatus(mediaServerInfo)
	if Err.Message != "" {
		logging.LOG.Warn(Err.Message)
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		return
	}

	Err = api.MediaServer_Init(mediaServerInfo)
	if Err.Message != "" {
		logging.LOG.Warn(Err.Message)
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		return
	}

	// Set the UserID in the response
	// if Jellyfin/Emby, we can get it from the server
	// if Plex, we set it to the existing value (it won't change)
	if api.Global_Config.MediaServer.Type == "Emby" || api.Global_Config.MediaServer.Type == "Jellyfin" {
		mediaServerInfo.UserID = api.Global_Config.MediaServer.UserID
	}

	// For security, mask the token in the response
	mediaServerInfo.Token = masking.Masking_Token(mediaServerInfo.Token)

	// Respond with a success message
	api.Util_Response_SendJson(w, http.StatusOK, api.JSONResponse{
		Status:  "success",
		Elapsed: api.Util_ElapsedTime(startTime),
		Data:    mediaServerInfo,
	})
}
