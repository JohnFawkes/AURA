package mediaserver

import (
	"aura/internal/cache"
	"aura/internal/config"
	"aura/internal/logging"
	"aura/internal/modals"
	mediaserver_shared "aura/internal/server/shared"
	"aura/internal/utils"
	"fmt"
	"net/http"
	"time"
)

func GetAllSectionItems(w http.ResponseWriter, r *http.Request) {
	logging.LOG.Trace(r.URL.Path)
	startTime := time.Now()
	Err := logging.NewStandardError()

	// Get the following information from the URL
	// Section ID
	// Section Title
	// Section Type
	// Item Start Index
	sectionID := r.URL.Query().Get("sectionID")
	sectionTitle := r.URL.Query().Get("sectionTitle")
	sectionType := r.URL.Query().Get("sectionType")
	sectionStartIndex := r.URL.Query().Get("sectionStartIndex")

	// Validate the section ID, title, type, and start index
	if sectionID == "" || sectionTitle == "" || sectionType == "" || sectionStartIndex == "" {

		Err.Message = "Missing required query parameters"
		Err.HelpText = "Ensure the URL contains sectionID, sectionTitle, sectionType, and sectionStartIndex query parameters."
		Err.Details = fmt.Sprintf("Received sectionID: %s, sectionTitle: %s, sectionType: %s, sectionStartIndex: %s",
			sectionID, sectionTitle, sectionType, sectionStartIndex)
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}

	// Fetch the section items from the media server
	section, Err := CallFetchLibrarySectionItems(sectionID, sectionTitle, sectionType, sectionStartIndex)
	if Err.Message != "" {
		logging.LOG.Error(Err.Message)
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}

	utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
		Status:  "success",
		Elapsed: utils.ElapsedTime(startTime),
		Data:    section,
	})
}

func CallFetchLibrarySectionItems(sectionID, sectionTitle, sectionType string, sectionStartIndex string) (modals.LibrarySection, logging.StandardError) {

	var section modals.LibrarySection
	section.ID = sectionID
	section.Title = sectionTitle
	section.Type = sectionType

	var mediaServer mediaserver_shared.MediaServer
	switch config.Global.MediaServer.Type {
	case "Plex":
		mediaServer = &mediaserver_shared.PlexServer{}
	case "Emby", "Jellyfin":
		mediaServer = &mediaserver_shared.EmbyJellyServer{}
	default:
		Err := logging.NewStandardError()

		Err.Message = "Unsupported media server type"
		Err.HelpText = "Ensure the media server type is either Plex, Emby, or Jellyfin."
		Err.Details = fmt.Sprintf("Received media server type: %s", config.Global.MediaServer.Type)
		return section, Err
	}

	// Fetch the section items from the media server
	mediaItems, totalSize, Err := mediaServer.FetchLibrarySectionItems(section, sectionStartIndex)
	if Err.Message != "" {
		logging.LOG.Warn(Err.Message)
		return section, Err
	}
	if len(mediaItems) == 0 {
		logging.LOG.Warn(fmt.Sprintf("Library section '%s' is empty", section.Title))

		Err.Message = "No items found in the library section"
		Err.HelpText = fmt.Sprintf("Ensure the section '%s' has items.", section.Title)
		Err.Details = fmt.Sprintf("No items found for section ID '%s' with title '%s'.", section.ID, section.Title)
		return section, Err
	}
	section.MediaItems = mediaItems
	section.TotalSize = totalSize
	cache.LibraryCacheStore.Update(&section)
	return section, logging.StandardError{}
}
