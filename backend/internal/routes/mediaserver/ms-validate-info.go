package routes_ms

import (
	"aura/internal/api"
	"aura/internal/logging"
	"aura/internal/masking"
	"net/http"
	"strings"
)

func ValidateNewInfo(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Validate Media Server Info", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	// Get the new media server info from the request
	var msConfig api.Config_MediaServer
	Err := api.DecodeRequestBodyJSON(ctx, r.Body, &msConfig, "Config_MediaServer")
	if Err.Message != "" {
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}

	// If the token is masked, get the actual token from the global config
	if strings.HasPrefix(msConfig.Token, "***") {
		msConfig.Token = api.Global_Config.MediaServer.Token
	}

	switch msConfig.Type {
	case "Plex":
		// Get the status
		_, Err = api.CallGetMediaServerStatus(ctx, msConfig)
		if Err.Message != "" {
			api.Util_Response_SendJSON(w, ld, nil)
			return
		}
	case "Emby", "Jellyfin":
		userID, Err := api.CallInitializeMediaServerConnection(ctx, msConfig)
		if Err.Message != "" {
			api.Util_Response_SendJSON(w, ld, nil)
			return
		}
		// Set the UserID in the response
		// if Jellyfin/Emby, we can get it from the server
		// if Plex, we set it to the existing value (it won't change)
		msConfig.UserID = userID
	default:
		logAction.SetError("Unsupported Media Server Type", "The media server type must be Plex, Emby, or Jellyfin", map[string]any{"type": msConfig.Type})
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}

	// Mask the token before sending the response
	msConfig.Token = masking.Masking_Token(msConfig.Token)

	api.Util_Response_SendJSON(w, ld, msConfig)
}
