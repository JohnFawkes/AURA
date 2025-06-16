package mediaserver

import (
	"aura/internal/config"
	"aura/internal/logging"
	"aura/internal/modals"
	mediaserver_shared "aura/internal/server/shared"
	"aura/internal/utils"
	"fmt"
	"net/http"
	"time"
)

func GetAllSections(w http.ResponseWriter, r *http.Request) {
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
		Err.HelpText = "Ensure the media server type is either Plex, Emby, or Jellyfin."
		Err.Details = fmt.Sprintf("Received media server type: %s", config.Global.MediaServer.Type)
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}

	var allSections []modals.LibrarySection
	for _, library := range config.Global.MediaServer.Libraries {
		found, Err := mediaServer.FetchLibrarySectionInfo(&library)
		if Err.Message != "" {
			logging.LOG.Warn(Err.Message)
			continue
		}
		if !found {
			logging.LOG.Warn(fmt.Sprintf("Library section '%s' not found in %s", library.Name, config.Global.MediaServer.Type))
			continue
		}
		if library.Type != "movie" && library.Type != "show" {
			logging.LOG.Warn(fmt.Sprintf("Library section '%s' is not a movie/show section", library.Name))
			continue
		}

		var section modals.LibrarySection
		section.ID = library.SectionID
		section.Type = library.Type
		section.Title = library.Name
		allSections = append(allSections, section)
	}

	if len(allSections) == 0 {

		Err.Message = "No sections found"
		Err.HelpText = fmt.Sprintf("Ensure that the media server has sections configured for %s.", config.Global.MediaServer.Type)
		Err.Details = fmt.Sprintf("No sections found in %s for the configured libraries", config.Global.MediaServer.Type)
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
