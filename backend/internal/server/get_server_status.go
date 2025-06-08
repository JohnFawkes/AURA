package mediaserver

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

	var mediaServer mediaserver_shared.MediaServer
	switch config.Global.MediaServer.Type {
	case "Plex":
		mediaServer = &mediaserver_shared.PlexServer{}
	case "Emby", "Jellyfin":
		mediaServer = &mediaserver_shared.EmbyJellyServer{}
	default:
		logErr := logging.ErrorLog{Err: fmt.Errorf("unsupported media server type: %s", config.Global.MediaServer.Type),
			Log: logging.Log{Message: fmt.Sprintf("Unsupported media server type: %s", config.Global.MediaServer.Type),
				Elapsed: utils.ElapsedTime(startTime),
			},
		}
		utils.SendErrorJSONResponse(w, http.StatusBadRequest, logErr)
		return
	}

	status, logErr := mediaServer.GetMediaServerStatus()
	if logErr.Err != nil {
		logging.LOG.Warn(logErr.Log.Message)
		utils.SendErrorJSONResponse(w, http.StatusInternalServerError, logErr)
		return
	}

	// Respond with a success message
	utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
		Status:  "success",
		Message: "Media server status retrieved successfully",
		Elapsed: utils.ElapsedTime(startTime),
		Data:    status,
	})
}
