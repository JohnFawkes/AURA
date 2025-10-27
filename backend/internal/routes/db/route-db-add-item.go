package routes_db

import (
	"aura/internal/api"
	"aura/internal/logging"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// AddItem handles requests to add a new media item along with its poster sets to the database.
//
// Method: POST
//
// Endpoint: /api/db/add/item
//
// It expects a JSON body containing the media item details and associated poster sets.
// The request body should include a TMDB_ID, LibraryTitle, MediaItem, and at least one PosterSet.
//
// On success, it responds with a confirmation message. On failure, it returns an error message.
func AddItem(w http.ResponseWriter, r *http.Request) {
	logging.LOG.Trace(r.URL.Path)
	startTime := time.Now()
	Err := logging.NewStandardError()

	// Parse the request body to get the DBMediaItemWithPosterSets
	var saveItem api.DBMediaItemWithPosterSets

	if err := json.NewDecoder(r.Body).Decode(&saveItem); err != nil {
		Err.Message = "Failed to decode request body"
		Err.HelpText = "Ensure the request body is a valid JSON object."
		Err.Details = map[string]any{
			"error": err.Error(),
			"body":  r.Body,
		}
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		return
	}

	// Validate the request body
	// Make sure it contains a TMDB_ID and LibraryTitle
	if saveItem.TMDB_ID == "" || saveItem.LibraryTitle == "" {
		Err.Message = "Missing required fields"
		Err.HelpText = "TMDB_ID and LibraryTitle are required."
		Err.Details = map[string]any{
			"body": saveItem,
		}
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		return
	}

	// Validate that there is a MediaItem
	if saveItem.MediaItem.TMDB_ID == "" {
		Err.Message = "No MediaItem provided"
		Err.HelpText = "A valid MediaItem is required."
		Err.Details = map[string]any{
			"body": saveItem,
		}
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		return
	}

	// Validate that there is at least one PosterSet
	if len(saveItem.PosterSets) == 0 {
		Err.Message = "No PosterSets provided"
		Err.HelpText = "At least one PosterSet is required."
		Err.Details = map[string]any{
			"body": saveItem,
		}
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		return
	}

	// Validate each PosterSetDetail
	for _, ps := range saveItem.PosterSets {
		if ps.PosterSetID == "" {
			Err.Message = "Invalid PosterSetDetail"
			Err.HelpText = "Each PosterSetDetail must have a valid PosterSetID and PosterSetJSON."
			Err.Details = map[string]any{
				"body": saveItem,
				"ps":   ps,
			}
			api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
			return
		}
	}

	// Save the item to the database
	Err = api.DB_InsertAllInfoIntoTables(saveItem)
	if Err.Message != "" {
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		return
	}

	inDB, posterSummary, Err := api.DB_CheckIfMediaItemExists(saveItem.TMDB_ID, saveItem.LibraryTitle)
	if Err.Message == "" && inDB && len(posterSummary) > 0 {
		logging.LOG.Debug(fmt.Sprintf("MediaItem %s (%s) now has %d poster sets in the DB:\n%v", saveItem.MediaItem.Title, saveItem.MediaItem.TMDB_ID, len(posterSummary), posterSummary))
		saveItem.MediaItem.DBSavedSets = posterSummary
		saveItem.MediaItem.ExistInDatabase = true
		// Update the in-memory cache
		api.Global_Cache_LibraryStore.UpdateMediaItem(saveItem.LibraryTitle, &saveItem.MediaItem)
	}

	// Respond with a success message
	api.Util_Response_SendJson(w, http.StatusOK, api.JSONResponse{
		Status:  "success",
		Elapsed: api.Util_ElapsedTime(startTime),
		Data:    saveItem,
	})
}
