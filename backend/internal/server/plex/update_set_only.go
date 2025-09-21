package plex

import (
	"aura/internal/config"
	"aura/internal/logging"
	"aura/internal/mediux"
	"aura/internal/modals"
	"fmt"
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

	// If config.Global.Kometa.RemoveLabels is true, remove the labels specified in the config
	if config.Global.Kometa.RemoveLabels {
		for _, label := range config.Global.Kometa.Labels {
			Err := removeLabel(itemRatingKey, label)
			if Err.Message != "" {
				logging.LOG.Warn(fmt.Sprintf("Failed to remove label '%s': %v", label, Err.Message))
			}
		}
	}

	return logging.StandardError{}
}
