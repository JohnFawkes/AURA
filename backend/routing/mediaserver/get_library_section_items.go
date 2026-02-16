package routes_ms

import (
	"aura/cache"
	"aura/logging"
	"aura/mediaserver"
	"aura/models"
	"aura/utils/httpx"
	"net/http"
	"time"
)

func GetLibrarySectionItems(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Get Library Section Items", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	actionGetQueryParams := logAction.AddSubAction("Get all query params", logging.LevelTrace)
	// Get the following information from the URL
	// Section ID
	// Section Title
	// Section Type
	// Item Start Index
	sectionID := r.URL.Query().Get("section_id")
	sectionTitle := r.URL.Query().Get("section_title")
	sectionType := r.URL.Query().Get("section_type")
	sectionStartIndex := r.URL.Query().Get("section_start_index")

	// Validate the section ID, title, type, and start index
	if sectionID == "" || sectionTitle == "" || sectionType == "" || sectionStartIndex == "" {
		actionGetQueryParams.SetError("Missing Query Parameters", "One or more required query parameters are missing",
			map[string]any{
				"section_id":          sectionID,
				"section_title":       sectionTitle,
				"section_type":        sectionType,
				"section_start_index": sectionStartIndex,
			})
		httpx.SendResponse(w, ld, nil)
		return
	}
	actionGetQueryParams.Complete()

	// Fetch the section items from the media server
	librarySection := models.LibrarySection{
		LibrarySectionBase: models.LibrarySectionBase{
			ID:    sectionID,
			Title: sectionTitle,
			Type:  sectionType,
		},
	}
	mediaItems, totalSize, Err := mediaserver.GetLibrarySectionItems(ctx, librarySection, sectionStartIndex, "")
	if Err.Message != "" {
		httpx.SendResponse(w, ld, nil)
		return
	}

	librarySection.MediaItems = mediaItems
	librarySection.TotalSize = totalSize
	cache.LibraryStore.LastFullUpdate = time.Now().Unix()
	httpx.SendResponse(w, ld, librarySection)
}
