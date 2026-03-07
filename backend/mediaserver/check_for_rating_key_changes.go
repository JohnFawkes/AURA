package mediaserver

import (
	"aura/cache"
	"aura/config"
	"aura/database"
	"aura/logging"
	"aura/models"
	"aura/notification"
	"context"
	"fmt"
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
		logging.LOGGER.Trace().Timestamp().Str("tmdb_id", dbItem.TMDB_ID).Str("library_title", dbItem.LibraryTitle).Str("title", dbItem.Title).Msg("Checking MediaItem for RatingKey changes")
		cachedItem, found := cache.LibraryStore.GetMediaItemFromSectionByTMDBID(dbItem.LibraryTitle, dbItem.TMDB_ID)
		if !found {
			logging.LOGGER.Warn().Timestamp().Str("tmdb_id", dbItem.TMDB_ID).Str("library_title", dbItem.LibraryTitle).Msg("MediaItem not found in cache while checking for RatingKey changes")
			sendNotFoundNotification(dbItem)
			continue
		}
		if dbItem.RatingKey != cachedItem.RatingKey {
			logging.LOGGER.Trace().Timestamp().Str("tmdb_id", dbItem.TMDB_ID).Str("library_title", dbItem.LibraryTitle).
				Msgf("Updating MediaItem RatingKey from %s to %s", dbItem.RatingKey, cachedItem.RatingKey)
			database.UpdateMediaItem(ctx, *cachedItem)
			logAction.AppendResult("updated_items", map[string]any{
				"tmdb_id":        dbItem.TMDB_ID,
				"library_title":  dbItem.LibraryTitle,
				"old_rating_key": dbItem.RatingKey,
				"new_rating_key": cachedItem.RatingKey,
			})
		}
	}

	return logging.LogErrorInfo{}
}

func sendNotFoundNotification(mediaItem models.MediaItem) {
	if len(config.Current.Notifications.Providers) == 0 || !config.Current.Notifications.Enabled {
		return
	}

	title := fmt.Sprintf("Media Item Not Found: %s", mediaItem.Title)
	message := fmt.Sprintf("The media item '%s' (TMDB ID: %s) in library '%s' could not be found in the cache. This may indicate an issue with the media server or that the item has been removed. This was noticed during the Rating Key Changes job. Please double check if this item exists. If it doesn't remove it from the Saved Sets.", mediaItem.Title, mediaItem.TMDB_ID, mediaItem.LibraryTitle)
	imageURL := ""

	ctx, ld := logging.CreateLoggingContext(context.Background(), "Notification - Send Not Found Alert")
	action := ld.AddAction("Sending Not Found Notification", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, action)
	defer ld.Log()
	defer action.Complete()

	// Send a notification to all configured providers
	// Send a notification to all configured providers
	for _, provider := range config.Current.Notifications.Providers {
		if provider.Enabled {
			switch provider.Provider {
			case "Discord":
				notification.SendDiscordMessage(
					ctx,
					provider.Discord,
					message,
					imageURL,
					title,
				)
			case "Pushover":
				notification.SendPushoverMessage(
					ctx,
					provider.Pushover,
					message,
					imageURL,
					title,
				)
			case "Gotify":
				notification.SendGotifyMessage(
					ctx,
					provider.Gotify,
					message,
					imageURL,
					title,
				)
			case "Webhook":
				notification.SendWebhookMessage(
					ctx,
					provider.Webhook,
					message,
					imageURL,
					title,
				)
			}
		}
	}
}
