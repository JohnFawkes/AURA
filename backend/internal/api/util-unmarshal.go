package api

import (
	"aura/internal/logging"
	"encoding/json"
)

func UnmarshalPosterSet(posterSetJSON string, posterSet *PosterSet) logging.LogErrorInfo {
	err := json.Unmarshal([]byte(posterSetJSON), posterSet)
	if err != nil {
		return logging.LogErrorInfo{
			Message: "Failed to unmarshal PosterSet JSON",
			Help:    "The PosterSet JSON data could not be parsed.",
			Detail: map[string]any{
				"error": err.Error(),
			},
		}
	}
	return logging.LogErrorInfo{}
}

func UnmarshalMediaItem(mediaItemJSON string, mediaItem *MediaItem) logging.LogErrorInfo {
	err := json.Unmarshal([]byte(mediaItemJSON), mediaItem)
	if err != nil {
		return logging.LogErrorInfo{
			Message: "Failed to unmarshal MediaItem JSON",
			Help:    "The MediaItem JSON data could not be parsed.",
			Detail: map[string]any{
				"error": err.Error(),
			},
		}
	}
	return logging.LogErrorInfo{}
}
