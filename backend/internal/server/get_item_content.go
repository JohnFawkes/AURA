package mediaserver

import (
	"aura/internal/cache"
	"aura/internal/config"
	"aura/internal/logging"
	mediaserver_shared "aura/internal/server/shared"
	"aura/internal/utils"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

func GetItemContent(w http.ResponseWriter, r *http.Request) {
	logging.LOG.Trace(r.URL.Path)
	startTime := time.Now()
	Err := logging.NewStandardError()

	// Get the ratingKey from the URL
	ratingKey := chi.URLParam(r, "ratingKey")
	if ratingKey == "" {
		Err.Message = "Missing rating key in URL"
		Err.HelpText = "Ensure the URL contains a valid rating key."
		Err.Details = fmt.Sprintf("Received ratingKey: %s", ratingKey)
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}

	// Get the library section title from the query parameters
	sectionTitle := r.URL.Query().Get("sectionTitle")
	if sectionTitle == "" {
		Err.Message = "Missing section title in query parameters"
		Err.HelpText = "Ensure the URL contains a valid sectionTitle query parameter."
		Err.Details = fmt.Sprintf("Received sectionTitle: %s", sectionTitle)
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
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

	itemInfo, Err := mediaServer.FetchItemContent(ratingKey, sectionTitle)
	if Err.Message != "" {
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}

	if itemInfo.RatingKey == "" {
		Err.Message = "Item content not found"
		Err.HelpText = "Ensure the rating key is valid and the item exists in the media server."
		Err.Details = fmt.Sprintf("No content found for ratingKey: %s in sectionTitle: %s", ratingKey, sectionTitle)
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}

	// Update the cache with the item content
	cache.LibraryCacheStore.UpdateMediaItem(sectionTitle, &itemInfo)

	// Respond with a success message
	utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
		Status:  "success",
		Elapsed: utils.ElapsedTime(startTime),
		Data:    itemInfo,
	})
}
