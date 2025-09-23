package plex

import (
	"aura/internal/logging"
	"aura/internal/mediux"
	"aura/internal/modals"
)

func UpdateSetOnly(item modals.MediaItem, file modals.PosterFile) logging.StandardError {

	// Determine the itemRatingKey
	itemRatingKey := getItemRatingKey(item, file)
	if itemRatingKey == "" {
		Err := logging.NewStandardError()
		Err.Message = "Failed to determine item rating key"
		Err.HelpText = "Ensure the item and file types are correctly set."
		return Err
	}

	mediuxImageUrl, Err := mediux.GetMediuxImageURL(file.ID, file.Modified.String(), "original")
	if Err.Message != "" {
		return Err
	}

	refreshPlexItem(itemRatingKey)
	setPoster(itemRatingKey, mediuxImageUrl, file.Type)

	Err = handleLabelsInPlex(item)
	if Err.Message != "" {
		logging.LOG.Warn(Err.Message)
	}

	return logging.StandardError{}
}
