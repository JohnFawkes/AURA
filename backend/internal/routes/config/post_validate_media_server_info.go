package route_config

import (
	"aura/internal/config"
	"aura/internal/logging"
	"aura/internal/modals"
	mediaserver_shared "aura/internal/server/shared"
	"aura/internal/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

func ValidateMediaServerNewInfoConnection(w http.ResponseWriter, r *http.Request) {
	logging.LOG.Trace(r.URL.Path)
	startTime := time.Now()
	Err := logging.NewStandardError()

	// Get the media server information from the request
	var mediaServerInfo modals.Config_MediaServer
	if err := json.NewDecoder(r.Body).Decode(&mediaServerInfo); err != nil {
		Err.Message = "Failed to decode request body"
		Err.HelpText = "Ensure the request body is valid JSON"
		Err.Details = fmt.Sprintf("Error: %v", err)
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}

	// Store the current values for Type, URL and Token
	currentType := config.Global.MediaServer.Type
	currentURL := config.Global.MediaServer.URL
	currentToken := config.Global.MediaServer.Token

	// Change the Global values temporarily
	config.Global.MediaServer.Type = mediaServerInfo.Type
	config.Global.MediaServer.URL = mediaServerInfo.URL
	if !strings.HasPrefix(mediaServerInfo.Token, "***") {
		config.Global.MediaServer.Token = mediaServerInfo.Token
	}

	var mediaServer mediaserver_shared.MediaServer
	switch config.Global.MediaServer.Type {
	case "Plex":
		mediaServer = &mediaserver_shared.PlexServer{}
	case "Emby", "Jellyfin":
		mediaServer = &mediaserver_shared.EmbyJellyServer{}
	default:
		Err.Message = "Unsupported media server type"
		Err.HelpText = fmt.Sprintf("The media server type '%s' is not supported.", config.Global.MediaServer.Type)
		Err.Details = fmt.Sprintf("Received media server type: %s", config.Global.MediaServer.Type)
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}

	// Restore the previous values
	defer func() {
		config.Global.MediaServer.Type = currentType
		config.Global.MediaServer.URL = currentURL
		config.Global.MediaServer.Token = currentToken
	}()

	status, Err := mediaServer.GetMediaServerStatus()
	if Err.Message != "" {
		logging.LOG.Warn(Err.Message)
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}

	// Respond with a success message
	utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
		Status:  "success",
		Elapsed: utils.ElapsedTime(startTime),
		Data:    status,
	})
}
