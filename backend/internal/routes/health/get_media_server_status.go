package route_health

import (
	"aura/internal/config"
	"aura/internal/logging"
	mediaserver_shared "aura/internal/server/shared"
	"aura/internal/utils"
	"fmt"
	"net/http"
	"time"
)

func GetMediaServerStatus(w http.ResponseWriter, r *http.Request) {
	logging.LOG.Trace(r.URL.Path)
	startTime := time.Now()
	Err := logging.NewStandardError()

	var mediaServer mediaserver_shared.MediaServer
	switch config.Global.MediaServer.Type {
	case "Plex":
		mediaServer = &mediaserver_shared.PlexServer{}
	case "Emby", "Jellyfin":
		mediaServer = &mediaserver_shared.EmbyJellyServer{}
	default:
		Err.Message = "Unsupported media server type"
		Err.HelpText = "Supported types are: Plex, Emby, Jellyfin"
		Err.Details = map[string]any{
			"error": fmt.Sprintf("Received media server type: %s", config.Global.MediaServer.Type),
		}
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}

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
