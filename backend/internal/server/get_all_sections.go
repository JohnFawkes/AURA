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

	var allSections []modals.LibrarySection
	for _, library := range config.Global.MediaServer.Libraries {
		found, logErr := mediaServer.FetchLibrarySectionInfo(&library)
		if logErr.Err != nil {
			logging.LOG.Warn(logErr.Log.Message)
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
		utils.SendErrorJSONResponse(w, http.StatusInternalServerError, logging.ErrorLog{
			Err: fmt.Errorf("no sections found in %s", config.Global.MediaServer.Type),
			Log: logging.Log{
				Message: fmt.Sprintf("No sections found in %s", config.Global.MediaServer.Type),
				Elapsed: utils.ElapsedTime(startTime),
			},
		})
		return
	}

	// Respond with a success message
	utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
		Status:  "success",
		Message: fmt.Sprintf("Fetched all sections from %s", config.Global.MediaServer.Type),
		Elapsed: utils.ElapsedTime(startTime),
		Data:    allSections,
	})
}
