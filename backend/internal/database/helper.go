package database

import (
	"aura/internal/logging"
	"aura/internal/modals"
	"encoding/json"
)

func UnmarshalPosterSet(posterSetJSON string, posterSet *modals.PosterSet) logging.ErrorLog {
	err := json.Unmarshal([]byte(posterSetJSON), posterSet)
	if err != nil {
		return logging.ErrorLog{Err: err, Log: logging.Log{
			Message: "Failed to unmarshal PosterSet JSON",
		}}
	}
	return logging.ErrorLog{}
}

func UnmarshalMediaItem(mediaItemJSON string, mediaItem *modals.MediaItem) logging.ErrorLog {
	err := json.Unmarshal([]byte(mediaItemJSON), mediaItem)
	if err != nil {
		return logging.ErrorLog{Err: err, Log: logging.Log{
			Message: "Failed to unmarshal MediaItem JSON",
		}}
	}
	return logging.ErrorLog{}
}
