package routes_db

import (
	"aura/internal/api"
	"aura/internal/logging"
	"context"
	"net/http"
)

func AddItem(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Add Item To Database", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	// Parse the request body to get the DBMediaItemWithPosterSets
	var saveItem api.DBMediaItemWithPosterSets
	Err := api.DecodeRequestBodyJSON(ctx, r.Body, &saveItem, "DBMediaItemWithPosterSets")
	if Err.Message != "" {
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}

	// Validate the JSON structure
	validateAction := logAction.AddSubAction("Validate Save Item", logging.LevelDebug)
	// Make sure it contains a TMDB_ID and LibraryTitle
	if saveItem.TMDB_ID == "" || saveItem.LibraryTitle == "" {
		validateAction.SetError("Missing Required Fields", "TMDB_ID or LibraryTitle is empty",
			map[string]any{
				"body": saveItem,
			})
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}

	// Validate that there is a MediaItem
	if saveItem.MediaItem.TMDB_ID == "" {
		validateAction.SetError("Missing Media Item Field", "MediaItem.TMDB_ID is empty",
			map[string]any{
				"body": saveItem,
			})
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}

	// Validate that there is at least one PosterSet
	if len(saveItem.PosterSets) == 0 {
		validateAction.SetError("Missing Poster Set", "At least one PosterSet is required",
			map[string]any{
				"body": saveItem,
			})
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}

	// Validate each PosterSetDetail
	for _, ps := range saveItem.PosterSets {
		if ps.PosterSetID == "" {
			validateAction.SetError("Missing PosterSetDetail Field", "PosterSetDetail.PosterSetID is empty",
				map[string]any{
					"body": saveItem,
				})
			api.Util_Response_SendJSON(w, ld, nil)
			return
		}
	}
	validateAction.Complete()

	// Response with a success message
	// logging.LOGGER.Info().Timestamp().Str("title", saveItem.MediaItem.Title).Str("library", saveItem.LibraryTitle).Msgf("Adding %s to database successfully", saveItem.MediaItem.Title)
	// api.Util_Response_SendJSON(w, ld, saveItem)
	// return

	// Save the item to the database
	Err = api.DB_InsertAllInfoIntoTables(ctx, saveItem)
	if Err.Message != "" {
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}

	inDB, posterSummary, Err := api.DB_CheckIfMediaItemExists(ctx, saveItem.TMDB_ID, saveItem.LibraryTitle)
	if Err.Message == "" && inDB && len(posterSummary) > 0 {
		addToCacheAction := logAction.AddSubAction("Add Item To Cache", logging.LevelDebug)
		saveItem.MediaItem.DBSavedSets = posterSummary
		saveItem.MediaItem.ExistInDatabase = true
		// Update the in-memory cache
		api.Global_Cache_LibraryStore.UpdateMediaItem(saveItem.LibraryTitle, &saveItem.MediaItem)
		addToCacheAction.Complete()
	}

	// Handle any labels and tags asynchronously
	go api.SR_CallHandleTags(context.Background(), saveItem.MediaItem)

	api.Util_Response_SendJSON(w, ld, saveItem)
}
