package mediaserver

import (
	"fmt"
	"net/http"
	"poster-setter/internal/config"
	"poster-setter/internal/logging"
	"poster-setter/internal/modals"
	mediaserver_shared "poster-setter/internal/server/shared"
	"poster-setter/internal/utils"
	"time"
)

func GetAllSectionItems(w http.ResponseWriter, r *http.Request) {
	logging.LOG.Trace(r.URL.Path)
	startTime := time.Now()

	// Get the following information from the URL
	// Section ID
	// Section Title
	// Section Type
	// Item Start Index
	sectionID := r.URL.Query().Get("sectionID")
	sectionTitle := r.URL.Query().Get("sectionTitle")
	sectionType := r.URL.Query().Get("sectionType")
	sectionStartIndex := r.URL.Query().Get("sectionStartIndex")

	logging.LOG.Trace(fmt.Sprintf("Section Start Index: '%s'", sectionStartIndex))

	// Validate the section ID, title, type, and start index
	if sectionID == "" || sectionTitle == "" || sectionType == "" || sectionStartIndex == "" {
		logErr := logging.ErrorLog{
			Err: fmt.Errorf("missing section ID, title, type, or start index"),
			Log: logging.Log{
				Message: "Missing section ID, title, type, or start index",
				Elapsed: utils.ElapsedTime(startTime),
			},
		}
		utils.SendErrorJSONResponse(w, http.StatusBadRequest, logErr)
		return
	}

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

	var section modals.LibrarySection
	section.ID = sectionID
	section.Title = sectionTitle
	section.Type = sectionType

	// Fetch the section items from the media server
	// starting from the specified index
	mediaItems, totalSize, logErr := mediaServer.FetchLibrarySectionItems(section, sectionStartIndex)
	if logErr.Err != nil {
		logging.LOG.Warn(logErr.Log.Message)
		utils.SendErrorJSONResponse(w, http.StatusInternalServerError, logErr)
		return
	}
	if len(mediaItems) == 0 {
		logging.LOG.Warn(fmt.Sprintf("Library section '%s' is empty", section.Title))
		logErr := logging.ErrorLog{
			Err: fmt.Errorf("library section '%s' is empty", section.Title),
			Log: logging.Log{
				Message: fmt.Sprintf("Library section '%s' is empty", section.Title),
				Elapsed: utils.ElapsedTime(startTime),
			},
		}
		utils.SendErrorJSONResponse(w, http.StatusInternalServerError, logErr)
		return
	}

	section.MediaItems = mediaItems
	section.TotalSize = totalSize

	utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
		Status:  "success",
		Message: "Fetched all section items successfully",
		Elapsed: utils.ElapsedTime(startTime),
		Data:    section,
	})
}
