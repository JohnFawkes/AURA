package plex

import (
	"fmt"
	"poster-setter/internal/logging"
	"poster-setter/internal/modals"
)

func UpdateSetOnly(item modals.MediaItem, file modals.PosterFile) logging.ErrorLog {

	// Determine the itemRatingKey
	itemRatingKey := getItemRatingKey(item, file)
	if itemRatingKey == "" {
		return logging.ErrorLog{Log: logging.Log{Message: "Item rating key is empty"}}
	}
	mediuxURL := fmt.Sprintf("%s/%s", "https://staged.mediux.io/assets", file.ID)
	refreshPlexItem(itemRatingKey)
	setPoster(itemRatingKey, mediuxURL, file.Type)

	return logging.ErrorLog{}
}
