package mediaserver

import (
	"aura/cache"
	"aura/database"
	"aura/logging"
	"context"
)

func CheckForRatingKeyChanges(ctx context.Context) (Err logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Checking for MediaItem RatingKey changes", logging.LevelDebug)
	defer logAction.Complete()

	// Get all MediaItems from the database
	dbMediaItems, logErr := database.GetAllMediaItems(ctx)
	if logErr.Message != "" {
		logAction.SetError("Failed to retrieve MediaItems from database", "", map[string]any{"error": logErr.Message})
		return *logAction.Error
	}

	// Compare each DB MediaItem's RatingKey with the Cache's RatingKey
	// If different, update the DB MediaItem's RatingKey
	for _, dbItem := range dbMediaItems {
		cachedItem, found := cache.LibraryStore.GetMediaItemFromSectionByTMDBID(dbItem.LibraryTitle, dbItem.TMDB_ID)
		if !found {
			continue
		}
		if dbItem.RatingKey != cachedItem.RatingKey {
			logging.LOGGER.Trace().Timestamp().Str("tmdb_id", dbItem.TMDB_ID).Str("library_title", dbItem.LibraryTitle).
				Msgf("Updating MediaItem RatingKey from %s to %s", dbItem.RatingKey, cachedItem.RatingKey)
			database.UpdateMediaItem(ctx, *cachedItem)
		}
	}

	return logging.LogErrorInfo{}
}
