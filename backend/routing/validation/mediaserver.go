package routes_validation

import (
	"aura/config"
	"aura/logging"
	"aura/mediaserver"
	"aura/utils/httpx"
	"fmt"
	"net/http"
)

func ValidateMediaServerInfo(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Validate Media Server Info", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	// Get the Media Server Info from the request body
	var mediaServerInfo config.Config_MediaServer
	Err := httpx.DecodeRequestBodyToJSON(ctx, r.Body, &mediaServerInfo, "Media Server Info")
	if Err.Message != "" {
		httpx.SendResponse(w, ld, nil)
		return
	}

	// If the Media Server Token is masked, retrieve the actual token from the config
	if config.IsMaskedField(mediaServerInfo.ApiToken) {
		mediaServerInfo.ApiToken = config.Current.MediaServer.ApiToken
	}

	switch mediaServerInfo.Type {
	case "Plex":
	case "Emby", "Jellyfin":
	default:
		logAction.SetError("Unsupported Media Server type: "+mediaServerInfo.Type, "", nil)
		httpx.SendResponse(w, ld, nil)
		return
	}

	connectionOk, serverName, serverVersion, Err := mediaserver.TestConnection(ctx, &mediaServerInfo)
	if Err.Message != "" {
		httpx.SendResponse(w, ld, nil)
		return
	} else if !connectionOk {
		logAction.SetError("Failed to connect to media server with provided information", "Please check the media server settings and try again", map[string]any{
			"media_server_info": mediaServerInfo,
		})
		httpx.SendResponse(w, ld, nil)
		return
	}

	var response struct {
		Valid       bool                      `json:"valid"`
		Message     string                    `json:"message"`
		MediaServer config.Config_MediaServer `json:"media_server"`
	}
	response.MediaServer = mediaServerInfo

	if mediaServerInfo.Type != "Plex" {
		adminUserID, Err := mediaserver.GetAdminUser(ctx, &mediaServerInfo)
		if Err.Message != "" {
			httpx.SendResponse(w, ld, nil)
			return
		}
		response.MediaServer.UserID = adminUserID
	}

	response.Valid = connectionOk
	response.Message = fmt.Sprintf("Successfully connected to %s server '%s' (version %s)", mediaServerInfo.Type, serverName, serverVersion)

	httpx.SendResponse(w, ld, response)
}
