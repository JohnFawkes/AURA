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

func CheckForMediaItemChanges(ctx context.Context) (Err logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Checking for Media Item changes", logging.LevelDebug)
	defer logAction.Complete()

	// Get all MediaItems from the database
	dbMediaItems, logErr := database.GetAllMediaItemsWithFlags(ctx)
	if logErr.Message != "" {
		logAction.SetError("Failed to retrieve MediaItems from database", "", map[string]any{"error": logErr.Message})
		return *logAction.Error
	}

	// Compare each DB MediaItem's RatingKey with the Cache's RatingKey
	// If different, update the DB MediaItem's RatingKey
	// If a DB MediaItem is not found in the Cache AND it doesn't have Saved Sets AND it is not Temp Ignored, this is an item we can remove from the DB
	// If a DB MediaItem is not found in the Cache AND it has Saved Sets, this is an item we want to keep in the DB but we should send a notification about it so the user can investigate
	// If a DB MediaItem is not found in the Cache AND it is Temp Ignored, this is an item we can remove from the DB
	for _, dbItem := range dbMediaItems {
		logging.LOGGER.Trace().Timestamp().Str("tmdb_id", dbItem.TMDB_ID).Str("library_title", dbItem.LibraryTitle).Str("title", dbItem.Title).Msg("Checking Media Item for changes")
		cachedItem, found := cache.LibraryStore.GetMediaItemFromSectionByTMDBID(dbItem.LibraryTitle, dbItem.TMDB_ID)
		if !found {
			mediaItem := models.MediaItem{
				TMDB_ID:      dbItem.TMDB_ID,
				LibraryTitle: dbItem.LibraryTitle,
				Title:        dbItem.Title,
				Year:         dbItem.Year,
			}
			message := ""
			if !dbItem.HasSavedSet && !dbItem.IsIgnored {
				// If the item is not found in the cache AND it doesn't have Saved Sets AND it is not Temp Ignored, this is an item we can remove from the DB
				logging.LOGGER.Warn().Timestamp().Str("tmdb_id", dbItem.TMDB_ID).Str("library_title", dbItem.LibraryTitle).Msg("MediaItem not found in cache and has no Saved Sets or Temp Ignore - removing from database")
				message = fmt.Sprintf("The media item '%s' (TMDB ID: %s) in library '%s' could not be found in the cache and has no Saved Sets and is not set to be ignored temporarily. This may indicate that the Media Item was removed or there is an issue with the media server. This item will be removed from the database, but please double check if this item exists. If it does exist and/or is something you want to keep, please add it back to the Saved Sets/Temp Ignore as appropriate.", dbItem.Title, dbItem.TMDB_ID, dbItem.LibraryTitle)
				database.DeleteMediaItemAndIgnoredStatus(ctx, dbItem.TMDB_ID, dbItem.LibraryTitle)
			} else if dbItem.HasSavedSet {
				// If the item is not found in the cache AND it has Saved Sets
				logging.LOGGER.Warn().Timestamp().Str("tmdb_id", dbItem.TMDB_ID).Str("library_title", dbItem.LibraryTitle).Msg("MediaItem not found in cache but has Saved Sets")
				message = fmt.Sprintf("The media item '%s' (TMDB ID: %s) in library '%s' could not be found in the cache but has Saved Sets. This may indicate that the Media Item was removed or there is an issue with the media server. This item will be kept in the database for now due to the existing Saved Sets, but please double check if this item exists. If it doesn't remove it from the Saved Sets.", dbItem.Title, dbItem.TMDB_ID, dbItem.LibraryTitle)
			} else if dbItem.IsIgnored {
				// If the item is not found in the cache AND it is Temp Ignored, we can just remove it from the DB
				logging.LOGGER.Warn().Timestamp().Str("tmdb_id", dbItem.TMDB_ID).Str("library_title", dbItem.LibraryTitle).Msg("MediaItem not found in cache but is Temp Ignored - removing from database")
				message = fmt.Sprintf("The media item '%s' (TMDB ID: %s) in library '%s' could not be found in the cache but is set to be ignored temporarily. This may indicate that the Media Item was removed or there is an issue with the media server. This item will be removed from the database since it is set to be ignored temporarily, but please double check if this item exists. If it does exist and you want to keep it as ignored temporarily, please ignore it again.", dbItem.Title, dbItem.TMDB_ID, dbItem.LibraryTitle)
				database.DeleteMediaItemAndIgnoredStatus(ctx, dbItem.TMDB_ID, dbItem.LibraryTitle)
			}

			sendNotFoundNotification(mediaItem, message)
			continue
		} else if dbItem.RatingKey != cachedItem.RatingKey {
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

func sendNotFoundNotification(mediaItem models.MediaItem, message string) {
	if len(config.Current.Notifications.Providers) == 0 || !config.Current.Notifications.Enabled {
		return
	}

	title := fmt.Sprintf("Media Item Not Found: %s", mediaItem.Title)
	imageURL := ""

	ctx, ld := logging.CreateLoggingContext(context.Background(), "Notification - Send Not Found Alert")
	action := ld.AddAction("Sending Not Found Notification", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, action)
	defer ld.Log()
	defer action.Complete()

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
