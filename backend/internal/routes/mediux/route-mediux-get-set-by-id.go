package routes_mediux

import (
	"aura/internal/api"
	"aura/internal/logging"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

// Route_Mediux_GetSetByID handles the API request to fetch a specific poster set from Mediux by its ID.
//
// Takes in the following:
// - setID as a URL parameter
//
// - librarySection as a query parameter (the library section name, e.g., "Movies" or "TV Shows")
//
// - tmdbID as a query parameter (the TMDB ID of the item associated with the set)
//
// - itemType as a query parameter (either "movie", "show", or "collection")
//
// Returns a JSON response with the poster set details as a `PosterSet` object.
func GetSetByID(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	logging.LOG.Trace(r.URL.Path)
	Err := logging.NewStandardError()

	// Get the set ID from the URL
	setID := chi.URLParam(r, "setID")
	if setID == "" {
		Err.Message = "Missing setID in URL"
		Err.HelpText = "Ensure the setID is provided in the URL path."
		Err.Details = map[string]any{
			"url":    r.URL.Path,
			"method": r.Method,
			"setID":  setID,
		}
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		return
	}

	// Get the librarySection, tmdbID and itemType from the query parameters
	librarySection := r.URL.Query().Get("librarySection")
	tmdbID := r.URL.Query().Get("tmdbID")
	itemType := r.URL.Query().Get("itemType")
	if librarySection == "" || tmdbID == "" || itemType == "" {
		Err.Message = "Missing librarySection, tmdbID or itemType in query parameters"
		Err.HelpText = "Ensure the librarySection, tmdbID and itemType are provided in query parameters."
		Err.Details = map[string]any{
			"url":            r.URL.Path,
			"method":         r.Method,
			"librarySection": librarySection,
			"tmdbID":         tmdbID,
			"itemType":       itemType,
		}
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		return
	}

	var updatedSet api.PosterSet

	switch itemType {
	case "show":
		updatedSet, Err = api.Mediux_FetchShowSetByID(librarySection, tmdbID, setID)
		if Err.Message != "" {
			api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
			return
		}
		if updatedSet.ID == "" {
			Err.Message = "No show set found for the provided ID"
			Err.HelpText = "Ensure the set ID is correct and the show set exists in the Mediux database."
			Err.Details = map[string]any{
				"setID":          setID,
				"librarySection": librarySection,
				"tmdbID":         tmdbID,
			}
			api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
			return
		}
	case "movie":
		updatedSet, Err = api.Mediux_FetchMovieSetByID(librarySection, tmdbID, setID)
		if Err.Message != "" {
			api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
			return
		}
		if updatedSet.ID == "" {
			Err.Message = "No movie set found for the provided ID"
			Err.HelpText = "Ensure the set ID is correct and the movie set exists in the Mediux database."
			Err.Details = map[string]any{
				"setID":          setID,
				"librarySection": librarySection,
				"tmdbID":         tmdbID,
			}
			api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
			return
		}
	case "collection":
		updatedSet, Err = api.Mediux_FetchCollectionSetByID(librarySection, tmdbID, setID)
		if Err.Message != "" {
			api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
			return
		}
		if updatedSet.ID == "" {
			Err.Message = "No collection set found for the provided ID"
			Err.HelpText = "Ensure the set ID is correct and the collection set exists in the Mediux database."
			Err.Details = map[string]any{
				"setID":          setID,
				"librarySection": librarySection,
				"tmdbID":         tmdbID,
			}
			api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
			return
		}
	default:
		Err.Message = "Invalid item type provided"
		Err.HelpText = "Ensure the item type is either 'movie', 'show' or 'collection'."
		Err.Details = map[string]any{
			"itemType": itemType,
		}
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		return
	}

	api.Util_Response_SendJson(w, http.StatusOK, api.JSONResponse{
		Status:  "success",
		Elapsed: api.Util_ElapsedTime(startTime),
		Data:    updatedSet,
	})
}
