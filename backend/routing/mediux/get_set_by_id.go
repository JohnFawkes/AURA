package routes_mediux

import (
	"aura/logging"
	"aura/mediux"
	"aura/models"
	"aura/utils/httpx"
	"net/http"
)

func GetSetByID(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Get MediUX Set By ID", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	actionGetQueryParams := logAction.AddSubAction("Get all query params", logging.LevelTrace)
	// Get the following information from the URL
	// Set ID
	// Library Section
	// Item Type
	// TMDB ID
	setID := r.URL.Query().Get("set_id")
	setType := r.URL.Query().Get("set_type")
	tmdbID := r.URL.Query().Get("tmdb_id")
	itemLibraryTitle := r.URL.Query().Get("item_library_title")
	// Validate the set ID, library section, item type, and TMDB ID
	if setID == "" || setType == "" || itemLibraryTitle == "" || (setType != "show" && setType != "movie" && setType != "collection") {
		actionGetQueryParams.SetError("Missing Query Parameters", "One or more required query parameters are missing or invalid",
			map[string]any{
				"set_id":           setID,
				"set_type":         setType,
				"valid_item_types": []string{"show", "movie", "collection"},
			})
		httpx.SendResponse(w, ld, nil)
		return
	}
	actionGetQueryParams.Complete()

	var response struct {
		Set           models.SetRef                  `json:"set"`
		IncludedItems map[string]models.IncludedItem `json:"included_items"`
	}

	switch setType {
	case "show":
		showSet, includedItems, Err := mediux.GetShowSetByID(ctx, setID, itemLibraryTitle)
		if Err.Message != "" {
			httpx.SendResponse(w, ld, nil)
			return
		}
		response.Set = showSet
		response.IncludedItems = includedItems
	case "movie":
		movieSet, includedItems, Err := mediux.GetMovieSetByID(ctx, setID, itemLibraryTitle)
		if Err.Message != "" {
			httpx.SendResponse(w, ld, nil)
			return
		}
		response.Set = movieSet
		response.IncludedItems = includedItems

	case "collection":
		collectionSet, includedItems, Err := mediux.GetMovieCollectionSetByID(ctx, setID, tmdbID, itemLibraryTitle)
		if Err.Message != "" {
			httpx.SendResponse(w, ld, nil)
			return
		}
		response.Set = collectionSet
		response.IncludedItems = includedItems
	}

	httpx.SendResponse(w, ld, response)
}
