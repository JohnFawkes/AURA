package api

import (
	"aura/internal/logging"
)

func Plex_UpdatePosterViaMediuxURL(item MediaItem, file PosterFile) logging.StandardError {

	// Determine the itemRatingKey
	itemRatingKey := Plex_GetItemRatingKey(item, file)
	if itemRatingKey == "" {
		Err := logging.NewStandardError()
		Err.Message = "Failed to determine item rating key"
		Err.HelpText = "Ensure the item and file types are correctly set."
		return Err
	}

	mediuxImageUrl, Err := Mediux_GetImageURL(file.ID, file.Modified.String(), "original")
	if Err.Message != "" {
		return Err
	}

	Plex_RefreshItem(itemRatingKey)
	Plex_SetPoster(itemRatingKey, mediuxImageUrl, file.Type)

	// Handle Labels and Tags in a separate go routine for quicker finish
	go func() {
		Err = Plex_HandleLabels(item)
		if Err.Message != "" {
			logging.LOG.Warn(Err.Message)
		}
		SR_CallHandleTags(item)
	}()
	return logging.StandardError{}
}
