package database

import (
	"aura/internal/logging"
	"aura/internal/modals"
	"encoding/json"
)

func UnmarshalPosterSet(posterSetJSON string, posterSet *modals.PosterSet) logging.StandardError {

	err := json.Unmarshal([]byte(posterSetJSON), posterSet)
	if err != nil {
		Err := logging.NewStandardError()
		Err.Message = "Failed to unmarshal PosterSet JSON"
		Err.HelpText = "Ensure the JSON structure matches the PosterSet model."
		Err.Details = map[string]any{
			"error": err.Error(),
			"json":  posterSetJSON,
		}
		return Err
	}
	return logging.StandardError{}
}

func UnmarshalMediaItem(mediaItemJSON string, mediaItem *modals.MediaItem) logging.StandardError {
	err := json.Unmarshal([]byte(mediaItemJSON), mediaItem)
	if err != nil {
		Err := logging.NewStandardError()
		Err.Message = "Failed to unmarshal MediaItem JSON"
		Err.HelpText = "Ensure the JSON structure matches the MediaItem model."
		Err.Details = map[string]any{
			"error": err.Error(),
			"json":  mediaItemJSON,
		}
		return Err
	}
	return logging.StandardError{}
}
