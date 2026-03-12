package mediaserver

import (
	"aura/config"
	"aura/database"
	"aura/logging"
	"aura/mediux"
	"aura/models"
	"aura/notification"
	"aura/utils"
	"context"
	"fmt"
)

func HandleTempIgnoredItems(ctx context.Context) (Err logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Handling Temp Ignored Items", logging.LevelInfo)
	defer logAction.Complete()

	Err = logging.LogErrorInfo{}

	// Get all temp ignored items from the database
	tempIgnoredItems, dbErr := database.GetTempIgnoredItems(ctx)
	if dbErr.Message != "" {
		return dbErr
	}

	for _, mediaItem := range tempIgnoredItems {
		numOfSets := 0
		var mainImage models.ImageFile
		switch mediaItem.Type {
		case "movie":
			setItems := map[string]models.IncludedItem{}
			// For Movies, we get Movie Sets and Movie Collection Sets
			movieSets, _ := mediux.GetMovieItemSets(ctx, mediaItem.TMDB_ID, mediaItem.LibraryTitle, &setItems)
			collectionSets, _ := mediux.GetMovieItemCollectionSets(ctx, mediaItem.TMDB_ID, mediaItem.LibraryTitle, &setItems)
			if len(movieSets) > 0 || len(collectionSets) > 0 {
				numOfSets += len(movieSets) + len(collectionSets)
				if len(movieSets) > 0 {
					mainImage = getMainImage(movieSets[0].Images)
				} else if len(collectionSets) > 0 {
					mainImage = getMainImage(collectionSets[0].Images)
				}
			}
		case "show":
			showSets, _, _ := mediux.GetShowItemSets(ctx, mediaItem.TMDB_ID, mediaItem.LibraryTitle)
			if len(showSets) > 0 {
				numOfSets += len(showSets)
				mainImage = getMainImage(showSets[0].Images)
			}
		}

		// If the item has sets, we can remove it from the temp ignored items
		if numOfSets > 0 {
			dbErr = database.StopIgnoringMediaItem(ctx, mediaItem.TMDB_ID, mediaItem.LibraryTitle)
			if dbErr.Message != "" {
				logging.LOGGER.Error().Timestamp().Str("tmdb_id", mediaItem.TMDB_ID).Str("library_title", mediaItem.LibraryTitle).Str("error", dbErr.Message).Msg("Failed to stop ignoring media item")
				continue
			}
			go sendNotification(mediaItem, numOfSets, mainImage)
		}
	}

	return Err
}

func sendNotification(mediaItem models.MediaItem, setCount int, mainImage models.ImageFile) {
	if len(config.Current.Notifications.Providers) == 0 || !config.Current.Notifications.Enabled {
		return
	}

	title := "New Set Detected for Ignored Item"
	message := fmt.Sprintf("A new set has been detected for the previously ignored item %s. It is now part of %d set(s) in MediUX, and will no longer be ignored in Aura.", utils.MediaItemInfo(mediaItem), setCount)
	imageURL := ""
	mediuxInfo, Err := mediux.GetBaseItemInfoByTMDB_ID(mediaItem.TMDB_ID, mediaItem.Type)
	if Err.Message != "" {
		imageURL = fmt.Sprintf("%s/%s?v=%s&key=jpg",
			"https://images.mediux.io/assets",
			mainImage.ID,
			mainImage.Modified.Format("20060102150405"),
		)
	} else {
		if mediuxInfo.TMDB_PosterPath == "" && mediuxInfo.TMDB_BackdropPath == "" {
			imageURL = fmt.Sprintf("%s/%s?v=%s&key=jpg",
				"https://images.mediux.io/assets",
				mainImage.ID,
				mainImage.Modified.Format("20060102150405"),
			)
		} else if mediuxInfo.TMDB_PosterPath != "" {
			imageURL = mediuxInfo.TMDB_PosterPath
		} else if mediuxInfo.TMDB_BackdropPath != "" {
			imageURL = mediuxInfo.TMDB_BackdropPath
		}
		if imageURL != "" {
			imageURL = fmt.Sprintf("https://image.tmdb.org/t/p/original%s", imageURL)
		}

	}

	ctx, ld := logging.CreateLoggingContext(context.Background(), "Notification - Send Temp Ignored Item Update")
	logAction := ld.AddAction("Sending Temp Ignored Item Update Notification", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)
	defer ld.Log()
	defer logAction.Complete()

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

func getMainImage(images []models.ImageFile) models.ImageFile {
	hasPosterImage := false
	hasBackdropImage := false
	var posterImage models.ImageFile
	var backdropImage models.ImageFile
	for _, image := range images {
		switch image.Type {
		case "poster":
			hasPosterImage = true
			posterImage = image
		case "backdrop":
			hasBackdropImage = true
			backdropImage = image
		}
	}

	if hasPosterImage {
		return posterImage
	} else if hasBackdropImage {
		return backdropImage
	}
	return images[0]
}
