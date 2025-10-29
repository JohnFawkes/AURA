package routes_mediux

import (
	"aura/internal/api"
	"aura/internal/logging"
	"net/http"
)

func GetAllSets(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Get All Sets", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	actionGetQueryParams := ld.AddAction("Get all query params", logging.LevelTrace)
	tmdbID := r.URL.Query().Get("tmdbID")
	itemType := r.URL.Query().Get("itemType")
	librarySection := r.URL.Query().Get("librarySection")
	if tmdbID == "" || itemType == "" || librarySection == "" {
		actionGetQueryParams.SetError("Missing Query Parameters", "One or more required query parameters are missing",
			map[string]any{
				"tmdbID":         tmdbID,
				"itemType":       itemType,
				"librarySection": librarySection,
			})
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}
	actionGetQueryParams.Complete()

	// If cache is empty, return false
	actionCheckCache := ld.AddAction("Check library cache", logging.LevelTrace)
	if api.Global_Cache_LibraryStore.IsEmpty() {
		actionCheckCache.SetError("Library Cache Empty", "The library cache is empty, cannot fetch sets", nil)
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}
	actionCheckCache.Complete()

	posterSets, Err := api.Mediux_FetchAllSets(ctx, tmdbID, itemType, librarySection)
	if Err.Message != "" {
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}

	api.Util_Response_SendJSON(w, ld, posterSets)
}
