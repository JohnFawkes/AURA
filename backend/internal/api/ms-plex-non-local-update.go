package api

import (
	"aura/internal/logging"
	"context"
	"fmt"
)

func Plex_UpdateImageViaMediuxURL(ctx context.Context, item MediaItem, file PosterFile) logging.LogErrorInfo {

	var posterImageType string
	switch file.Type {
	case "poster":
		posterImageType = "Poster"
	case "backdrop":
		posterImageType = "Backdrop"
	case "seasonPoster", "specialSeasonPoster":
		posterImageType = fmt.Sprintf("Season %d Poster", file.Season.Number)
	case "titlecard":
		posterImageType = fmt.Sprintf("S%dE%d Titlecard", file.Episode.SeasonNumber, file.Episode.EpisodeNumber)
	}

	ctx, logAction := logging.AddSubActionToContext(ctx, fmt.Sprintf("Update %s via MediUX URL", posterImageType), logging.LevelInfo)
	defer logAction.Complete()

	// Determine the itemRatingKey
	itemRatingKey := Plex_GetItemRatingKey(item, file)
	if itemRatingKey == "" {
		logAction.SetError("Failed to determine Rating Key for item",
			"Cannot update image without a valid Rating Key",
			map[string]any{
				"Item": item,
				"File": file,
			})
		return *logAction.Error
	}
	logAction.AppendResult("item_rating_key", itemRatingKey)

	// Get the Image URL from MediUX
	mediuxURL, Err := Mediux_GetImageURL(ctx, file.ID, file.Modified.String(), MediuxImageQualityOriginal)
	if Err.Message != "" {
		return Err
	}

	// Refresh the Plex Item
	Plex_RefreshItem(ctx, itemRatingKey)

	// Set the Poster using the MediUX URL
	Plex_SetPoster(ctx, itemRatingKey, mediuxURL, file.Type)

	return logging.LogErrorInfo{}
}

func Plex_UpdateCollectionImageViaMediuxURL(ctx context.Context, collectionItem CollectionItem, file PosterFile) logging.LogErrorInfo {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Update Collection Image via MediUX URL", logging.LevelInfo)
	defer logAction.Complete()

	itemRatingKey := collectionItem.RatingKey
	if itemRatingKey == "" {
		logAction.SetError("Collection Item Rating Key is required",
			"Cannot update collection image without a valid Rating Key",
			map[string]any{
				"CollectionItem": collectionItem,
			})
		return *logAction.Error
	}

	// Get the Image URL from MediUX
	mediuxURL, Err := Mediux_GetImageURL(ctx, file.ID, file.Modified.String(), MediuxImageQualityOriginal)
	if Err.Message != "" {
		return Err
	}

	// Set the Poster using the MediUX URL
	Plex_SetPoster(ctx, itemRatingKey, mediuxURL, file.Type)

	return logging.LogErrorInfo{}
}
