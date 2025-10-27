package routes_db

import (
	"aura/internal/api"
	"aura/internal/logging"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

/*
	Route_DB_DeleteItem

Delete a media item and its associated poster sets from the database.

Method: DELETE
URL: /api/db/delete/mediaitem/{tmdbID}/{libraryTitle}

Path Parameters:
- tmdbID (string): The TMDB ID of the media item to delete.
- libraryTitle (string): The title of the library the media item belongs to.
*/
func DeleteItem(w http.ResponseWriter, r *http.Request) {
	logging.LOG.Debug(r.URL.Path)
	startTime := time.Now()
	Err := logging.NewStandardError()

	// Get the tmdbID from the URL
	tmdbID := chi.URLParam(r, "tmdbID")
	libraryTitle := chi.URLParam(r, "libraryTitle")
	if tmdbID == "" {
		Err.Message = "Missing TMDB ID in Request"
		Err.HelpText = "Ensure the request includes a valid tmdbID parameter."
		Err.Details = map[string]any{
			"url":          r.URL.Path,
			"method":       r.Method,
			"tmdbID":       tmdbID,
			"libraryTitle": libraryTitle,
		}
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		return
	}
	if libraryTitle == "" {
		Err.Message = "Missing Library Title in Request"
		Err.HelpText = "Ensure the request includes a valid libraryTitle parameter."
		Err.Details = map[string]any{
			"url":          r.URL.Path,
			"method":       r.Method,
			"tmdbID":       tmdbID,
			"libraryTitle": libraryTitle,
		}
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		return
	}

	Err = api.DB_DeleteItem(tmdbID, libraryTitle)
	if Err.Message != "" {
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		return
	}

	// Get the media item from the cache and then update its DBSavedSets and update the cache
	mediaItem, inCache := api.Global_Cache_LibraryStore.GetMediaItemFromSectionByTMDBID(libraryTitle, tmdbID)
	inDB, posterSummary, Err := api.DB_CheckIfMediaItemExists(tmdbID, libraryTitle)
	if Err.Message == "" && inDB && len(posterSummary) > 0 {
		// Get the media item from the cache and then update its DBSavedSets and update the cache
		if inCache {
			logging.LOG.Debug(fmt.Sprintf("MediaItem %s (%s) now has %d poster sets in the DB:\n%v", mediaItem.Title, mediaItem.TMDB_ID, len(posterSummary), posterSummary))
			mediaItem.DBSavedSets = posterSummary
			mediaItem.ExistInDatabase = true
			api.Global_Cache_LibraryStore.UpdateMediaItem(libraryTitle, mediaItem)
		}
	} else if Err.Message == "" && !inDB {
		// If the item is not in the DB, remove it from the cache
		if inCache {
			logging.LOG.Debug(fmt.Sprintf("MediaItem %s (%s) no longer exists in the DB, removing from cache", mediaItem.Title, mediaItem.TMDB_ID))
			mediaItem.DBSavedSets = []api.PosterSetSummary{}
			mediaItem.ExistInDatabase = false
			api.Global_Cache_LibraryStore.UpdateMediaItem(libraryTitle, mediaItem)
		}
	}

	api.Util_Response_SendJson(w, http.StatusOK, api.JSONResponse{
		Status:  "success",
		Elapsed: api.Util_ElapsedTime(startTime),
		Data:    "success"})
}
