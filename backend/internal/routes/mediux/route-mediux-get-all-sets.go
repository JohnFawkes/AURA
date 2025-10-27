package routes_mediux

import (
	"aura/internal/api"
	"aura/internal/logging"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

// GetAllSets handles HTTP requests to retrieve all poster sets for a given TMDB ID and item type.
// It extracts the TMDB ID and item type from the URL parameters, validates them, and fetches
// the corresponding poster sets from the Mediux service. If successful, it responds with the
// retrieved data in JSON format. In case of errors, it sends an appropriate error response.
//
// Parameters:
//   - w: The HTTP response writer used to send the response.
//   - r: The HTTP request containing the URL parameters.
//
// URL Parameters:
//   - tmdbID: The TMDB ID of the item for which poster sets are being retrieved.
//   - itemType: The type of the item (e.g., movie, show).
//
// Responses:
//   - 200 OK: If the poster sets are successfully retrieved, the response contains a JSON object
//     with the status, message, elapsed time, and the retrieved data.
//   - 500 Internal Server Error: If there is an error (e.g., missing parameters or a failure
//     during data retrieval), the response contains an error message and elapsed time.
func GetAllSets(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	logging.LOG.Trace(r.URL.Path)

	// Create a new StandardError with details about the error
	Err := logging.NewStandardError()

	// Get the TMDB ID from the URL
	tmdbID := chi.URLParam(r, "tmdbID")
	itemType := chi.URLParam(r, "itemType")
	librarySection := chi.URLParam(r, "librarySection")
	if tmdbID == "" || itemType == "" || librarySection == "" {
		Err.Message = "Missing TMDB ID, Item Type or Library Section in URL Parameters"
		Err.HelpText = "Ensure the TMDB ID, Item Type and Library Section are provided in path parameters."
		Err.Details = map[string]any{
			"url":            r.URL.Path,
			"method":         r.Method,
			"tmdbID":         tmdbID,
			"itemType":       itemType,
			"librarySection": librarySection,
		}
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		return
	}

	// If cache is empty, return false
	if api.Global_Cache_LibraryStore.IsEmpty() {
		Err.Message = "Backend cache is empty"
		Err.HelpText = "Try refreshing the cache from the Home Page"
		Err.Details = "This typically happens when the backend cache is not initialized or has been cleared. Example on application restart."
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		return
	}

	logging.LOG.Debug(fmt.Sprintf("Fetching all sets for TMDB ID: %s, item type: %s, library section: %s", tmdbID, itemType, librarySection))

	posterSets, Err := api.Mediux_FetchAllSets(tmdbID, itemType, librarySection)
	if Err.Message != "" {
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		return
	}

	if len(posterSets) == 0 {
		Err.Message = "No sets found for the provided TMDB ID and Item Type"
		Err.HelpText = "Ensure the TMDB ID and Item Type are correct and that sets exist for this item."
		Err.Details = map[string]any{
			"tmdbID":         tmdbID,
			"itemType":       itemType,
			"librarySection": librarySection,
		}
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		return
	}

	// Respond with a success message
	api.Util_Response_SendJson(w, http.StatusOK, api.JSONResponse{
		Status:  "success",
		Elapsed: api.Util_ElapsedTime(startTime),
		Data:    posterSets,
	})
}
