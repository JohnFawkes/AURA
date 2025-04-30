package mediaserver

import (
	"fmt"
	"net/http"
	"poster-setter/internal/config"
	"poster-setter/internal/logging"
	mediaserver_shared "poster-setter/internal/server/shared"
	"poster-setter/internal/utils"
	"time"

	"github.com/go-chi/chi/v5"
)

func GetItemContent(w http.ResponseWriter, r *http.Request) {
	logging.LOG.Trace(r.URL.Path)
	startTime := time.Now()

	// Get the ratingKey from the URL
	ratingKey := chi.URLParam(r, "ratingKey")
	if ratingKey == "" {
		utils.SendErrorJSONResponse(w, http.StatusInternalServerError, logging.ErrorLog{
			Err: fmt.Errorf("missing rating key"),
			Log: logging.Log{
				Message: "Missing rating key in URL",
				Elapsed: utils.ElapsedTime(startTime),
			},
		})
		return
	}

	// Get the library section title from the query parameters
	sectionTitle := r.URL.Query().Get("sectionTitle")
	if sectionTitle == "" {
		utils.SendErrorJSONResponse(w, http.StatusInternalServerError, logging.ErrorLog{
			Err: fmt.Errorf("missing section title"),
			Log: logging.Log{
				Message: "Missing section title in query parameters",
				Elapsed: utils.ElapsedTime(startTime),
			},
		})
		return
	}

	var mediaServer mediaserver_shared.MediaServer
	switch config.Global.MediaServer.Type {
	case "Plex":
		mediaServer = &mediaserver_shared.PlexServer{}
	case "Emby", "Jellyfin":
		mediaServer = &mediaserver_shared.EmbyJellyServer{}
	default:
		logErr := logging.ErrorLog{
			Err: fmt.Errorf("unsupported media server type: %s", config.Global.MediaServer.Type),
			Log: logging.Log{Message: fmt.Sprintf("Unsupported media server type: %s", config.Global.MediaServer.Type),
				Elapsed: utils.ElapsedTime(startTime),
			},
		}
		utils.SendErrorJSONResponse(w, http.StatusBadRequest, logErr)
		return
	}

	itemInfo, logErr := mediaServer.FetchItemContent(ratingKey, sectionTitle)
	if logErr.Err != nil {
		utils.SendErrorJSONResponse(w, http.StatusInternalServerError, logErr)
		return
	}

	if itemInfo.RatingKey == "" {
		utils.SendErrorJSONResponse(w, http.StatusInternalServerError, logging.ErrorLog{
			Err: fmt.Errorf("bad rating key"),
			Log: logging.Log{
				Message: "No item found with the given rating key",
				Elapsed: utils.ElapsedTime(startTime),
			},
		})
		return
	}

	// Respond with a success message
	utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
		Status:  "success",
		Message: fmt.Sprintf("Retrieved item content from %s", config.Global.MediaServer.Type),
		Elapsed: utils.ElapsedTime(startTime),
		Data:    itemInfo,
	})
}
