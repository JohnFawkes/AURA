package plex

import (
	"fmt"
	"poster-setter/internal/config"
	"poster-setter/internal/logging"
	"poster-setter/internal/modals"
)

func UpdateSetOnly(item modals.MediaItem, file modals.PosterFile) logging.ErrorLog {

	// Determine the itemRatingKey
	itemRatingKey := getItemRatingKey(item, file)
	if itemRatingKey == "" {
		return logging.ErrorLog{Err: fmt.Errorf("item rating key is empty"),
			Log: logging.Log{Message: "Item rating key is empty"}}
	}
	mediuxURL := fmt.Sprintf("%s/%s", "https://staged.mediux.io/assets", file.ID)
	refreshPlexItem(itemRatingKey)
	setPoster(itemRatingKey, mediuxURL, file.Type)

	// If config.Global.Kometa.RemoveLabels is true, remove the labels specified in the config
	if config.Global.Kometa.RemoveLabels {
		for _, label := range config.Global.Kometa.Labels {
			logErr := removeLabel(itemRatingKey, label)
			if logErr.Err != nil {
				logging.LOG.Warn(fmt.Sprintf("Failed to remove label '%s': %v", label, logErr.Err))
			}
		}
	}

	return logging.ErrorLog{}
}
