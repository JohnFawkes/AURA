package routes_db

import (
	"aura/internal/api"
	"aura/internal/logging"
	"net/http"
)

func DeleteItem(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Delete Item From Database", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	// Get the TMDB_ID from the query parameters
	queryParamsAction := logAction.AddSubAction("Get TMDB_ID from Query Parameters", logging.LevelDebug)
	defer queryParamsAction.Complete()
	tmdbID := r.URL.Query().Get("tmdbID")
	if tmdbID == "" {
		queryParamsAction.SetError("Missing TMDB_ID", "TMDB_ID query parameter is required",
			map[string]any{
				"params": r.URL.Query(),
			})
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}

	// Get the LibraryTitle from the query parameters
	libraryTitle := r.URL.Query().Get("libraryTitle")
	if libraryTitle == "" {
		queryParamsAction.SetError("Missing LibraryTitle", "LibraryTitle query parameter is required",
			map[string]any{
				"params": r.URL.Query(),
			})
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}

	// Delete the item from the database
	Err := api.DB_DeleteItem(ctx, tmdbID, libraryTitle)
	if Err.Message != "" {
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}

	// Get the media item from the cache and then update its DBSavedSets and update the cache
	mediaItem, inCache := api.Global_Cache_LibraryStore.GetMediaItemFromSectionByTMDBID(libraryTitle, tmdbID)
	inDB, posterSummary, Err := api.DB_CheckIfMediaItemExists(ctx, tmdbID, libraryTitle)
	if Err.Message == "" && inDB && len(posterSummary) > 0 {
		// Get the media item from the cache and then update its DBSavedSets and update the cache
		if inCache {
			updateCacheAction := logAction.AddSubAction("Update Media Item Cache After Deletion", logging.LevelDebug)
			mediaItem.DBSavedSets = posterSummary
			mediaItem.ExistInDatabase = true
			api.Global_Cache_LibraryStore.UpdateMediaItem(libraryTitle, mediaItem)
			updateCacheAction.Complete()
		}
	} else if Err.Message == "" && !inDB {
		// If the item is not in the DB, remove it from the cache
		if inCache {
			removeCacheAction := logAction.AddSubAction("Remove Media Item From Cache After Deletion", logging.LevelDebug)
			mediaItem.DBSavedSets = []api.PosterSetSummary{}
			mediaItem.ExistInDatabase = false
			api.Global_Cache_LibraryStore.UpdateMediaItem(libraryTitle, mediaItem)
			removeCacheAction.Complete()
		}
	}

	api.Util_Response_SendJSON(w, ld, "success")
}
