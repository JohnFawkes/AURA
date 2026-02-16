package routes_ms

import (
	"aura/config"
	"aura/logging"
	"aura/mediaserver"
	"aura/models"
	"aura/utils/httpx"
	"net/http"
)

func GetLibrarySections(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Get Library Sections", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	var allLibrarySections []models.LibrarySection

	for _, libSection := range config.Current.MediaServer.Libraries {
		base := models.LibrarySectionBase{
			ID:    libSection.ID,
			Title: libSection.Title,
			Type:  libSection.Type,
			Path:  libSection.Path,
		}
		librarySection := models.LibrarySection{
			LibrarySectionBase: base,
		}

		found, Err := mediaserver.GetLibrarySectionDetails(ctx, &librarySection)
		if Err.Message != "" || !found {
			continue
		}

		allLibrarySections = append(allLibrarySections, librarySection)
	}

	if len(allLibrarySections) == 0 {
		logAction.SetError("No library sections found", "No valid library sections could be found on the configured Media Server", nil)
		httpx.SendResponse(w, ld, nil)
		return
	}

	httpx.SendResponse(w, ld, allLibrarySections)
}
