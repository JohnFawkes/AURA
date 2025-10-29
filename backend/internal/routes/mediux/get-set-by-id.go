package routes_mediux

import (
	"aura/internal/api"
	"aura/internal/logging"
	"net/http"
)

func GetSetByID(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Get Set By ID", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	actionGetQueryParams := logAction.AddSubAction("Get all query params", logging.LevelTrace)
	// Get the following information from the URL
	// Set ID
	// Library Section
	// Item Type
	// TMDB ID
	setID := r.URL.Query().Get("setID")
	librarySection := r.URL.Query().Get("librarySection")
	itemType := r.URL.Query().Get("itemType")
	tmdbID := r.URL.Query().Get("tmdbID")

	// Validate the set ID, library section, item type, and TMDB ID
	if setID == "" || librarySection == "" || itemType == "" || tmdbID == "" || (itemType != "show" && itemType != "movie" && itemType != "collection") {
		actionGetQueryParams.SetError("Missing Query Parameters", "One or more required query parameters are missing or invalid",
			map[string]any{
				"setID":          setID,
				"librarySection": librarySection,
				"itemType":       itemType,
				"tmdbID":         tmdbID,
				"validItemTypes": []string{"show", "movie", "collection"},
			})
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}
	actionGetQueryParams.Complete()

	var updatedSet api.PosterSet
	var Err logging.LogErrorInfo
	switch itemType {
	case "show":
		updatedSet, Err = api.Mediux_FetchShowSetByID(ctx, librarySection, tmdbID, setID)
		if Err.Message != "" {
			api.Util_Response_SendJSON(w, ld, nil)
			return
		}
	case "movie":
		updatedSet, Err = api.Mediux_FetchMovieSetByID(ctx, librarySection, tmdbID, setID)
		if Err.Message != "" {
			api.Util_Response_SendJSON(w, ld, nil)
			return
		}
	case "collection":
		updatedSet, Err = api.Mediux_FetchCollectionSetByID(ctx, librarySection, tmdbID, setID)
		if Err.Message != "" {
			api.Util_Response_SendJSON(w, ld, nil)
			return
		}
	}

	api.Util_Response_SendJSON(w, ld, updatedSet)
}
