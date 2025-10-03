package mediaserver

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

func GetAllLibrariesWithRequest(w http.ResponseWriter, r *http.Request) {
	logging.LOG.Trace(r.URL.Path)
	startTime := time.Now()
	Err := logging.NewStandardError()

	// Get the media server information from the request
	var mediaServerInfo modals.Config_MediaServer
	if err := json.NewDecoder(r.Body).Decode(&mediaServerInfo); err != nil {
		Err.Message = "Failed to decode request body"
		Err.HelpText = "Ensure the request body is valid JSON"
		Err.Details = map[string]any{
			"error": err.Error(),
		}
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}

	// Store the current values for Type, URL and Token
	currentType := config.Global.MediaServer.Type
	currentURL := config.Global.MediaServer.URL
	currentToken := config.Global.MediaServer.Token
	currentLibraries := config.Global.MediaServer.Libraries
	currentUserID := config.Global.MediaServer.UserID

	// Change the Global values temporarily
	config.Global.MediaServer.Type = mediaServerInfo.Type
	config.Global.MediaServer.URL = mediaServerInfo.URL
	config.Global.MediaServer.Libraries = mediaServerInfo.Libraries
	config.Global.MediaServer.UserID = mediaServerInfo.UserID
	if !strings.HasPrefix(mediaServerInfo.Token, "***") {
		config.Global.MediaServer.Token = mediaServerInfo.Token
	}

	// Restore the previous values
	defer func() {
		config.Global.MediaServer.Type = currentType
		config.Global.MediaServer.URL = currentURL
		config.Global.MediaServer.Token = currentToken
		config.Global.MediaServer.Libraries = currentLibraries
		config.Global.MediaServer.UserID = currentUserID
	}()

	allSections, Err := CallFetchLibrarySectionNames()
	if Err.Message != "" {
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}

	if len(allSections) == 0 {
		Err.Message = "No sections found"
		Err.HelpText = fmt.Sprintf("Ensure that the media server has sections configured for %s.", config.Global.MediaServer.Type)
		Err.Details = map[string]any{
			"error": fmt.Sprintf("No sections found in %s for the configured libraries", config.Global.MediaServer.Type),
		}
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}

	// Respond with a success message
	utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
		Status:  "success",
		Elapsed: utils.ElapsedTime(startTime),
		Data:    allSections,
	})
}

func CallFetchLibrarySectionNames() ([]string, logging.StandardError) {

	var mediaServer mediaserver_shared.MediaServer
	switch config.Global.MediaServer.Type {
	case "Plex":
		mediaServer = &mediaserver_shared.PlexServer{}
	case "Emby", "Jellyfin":
		mediaServer = &mediaserver_shared.EmbyJellyServer{}
	default:
		Err := logging.NewStandardError()
		Err.Message = "Unsupported media server type"
		Err.HelpText = "Supported types are: Plex, Emby, Jellyfin"
		Err.Details = map[string]any{
			"error": fmt.Sprintf("Received media server type: %s", config.Global.MediaServer.Type),
		}
		return nil, Err
	}

	allOptions, Err := mediaServer.FetchLibrarySectionOptions()
	if Err.Message != "" {
		return nil, Err
	}

	return allOptions, logging.StandardError{}
}
