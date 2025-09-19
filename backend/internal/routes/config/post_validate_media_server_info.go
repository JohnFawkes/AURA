package route_config

import (
	"aura/internal/config"
	"aura/internal/logging"
	"aura/internal/modals"
	mediaserver_shared "aura/internal/server/shared"
	"aura/internal/utils"
	"aura/internal/utils/masking"
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
	currentUserID := config.Global.MediaServer.UserID

	// Change the Global values temporarily
	config.Global.MediaServer.Type = mediaServerInfo.Type
	config.Global.MediaServer.URL = mediaServerInfo.URL
	if !strings.HasPrefix(mediaServerInfo.Token, "***") {
		config.Global.MediaServer.Token = mediaServerInfo.Token
	}
	config.Global.MediaServer.UserID = "" // Clear UserID to force re-fetch

	// Restore the previous values
	defer func() {
		config.Global.MediaServer.Type = currentType
		config.Global.MediaServer.URL = currentURL
		config.Global.MediaServer.Token = currentToken
		config.Global.MediaServer.UserID = currentUserID
		logging.LOG.Debug("Restored original media server configuration values")
		logging.LOG.Debug(fmt.Sprintf("Type: %s, URL: %s, Token: %s, UserID: %s", currentType, currentURL, currentToken, currentUserID))
	}()

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

	_, Err = mediaServer.GetMediaServerStatus()
	if Err.Message != "" {
		logging.LOG.Warn(Err.Message)
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}

	Err = mediaserver_shared.InitUserID()
	if Err.Message != "" {
		logging.LOG.Warn(Err.Message)
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}

	// Set the UserID in the response
	// if Jellyfin/Emby, we can get it from the server
	// if Plex, we set it to the existing value (it won't change)
	if config.Global.MediaServer.Type == "Emby" || config.Global.MediaServer.Type == "Jellyfin" {
		mediaServerInfo.UserID = config.Global.MediaServer.UserID
	}

	// For security, mask the token in the response
	mediaServerInfo.Token = masking.MaskToken(mediaServerInfo.Token)

	// Respond with a success message
	utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
		Status:  "success",
		Elapsed: utils.ElapsedTime(startTime),
		Data:    mediaServerInfo,
	})
}
