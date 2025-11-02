package api

import (
	"aura/internal/logging"
	"context"
)

func Plex_UpdatePosterViaMediuxURL(ctx context.Context, item MediaItem, file PosterFile) logging.LogErrorInfo {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Update Poster via MediUX URL", logging.LevelInfo)
	defer logAction.Complete()

	// Determine the itemRatingKey
	itemRatingKey := Plex_GetItemRatingKey(item, file)
	if itemRatingKey == "" {
		logAction.SetError("Failed to determine Rating Key for item",
			"Cannot update poster without a valid Rating Key",
			map[string]any{
				"Item": item,
				"File": file,
			})
		return *logAction.Error
	}
	logAction.AppendResult("item_rating_key", itemRatingKey)

	// Get the Image URL from MediUX
	mediuxURL, Err := Mediux_GetImageURL(ctx, file.ID, file.Modified.String(), "original")
	if Err.Message != "" {
		return Err
	}

	// Refresh the Plex Item
	Plex_RefreshItem(ctx, itemRatingKey)

	// Set the Poster using the MediUX URL
	Plex_SetPoster(ctx, itemRatingKey, mediuxURL, file.Type)

	return logging.LogErrorInfo{}
}
