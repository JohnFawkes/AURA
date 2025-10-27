package api

import (
	"aura/internal/logging"
	"encoding/json"
)

func UnmarshalPosterSet(posterSetJSON string, posterSet *PosterSet) logging.StandardError {

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

func UnmarshalMediaItem(mediaItemJSON string, mediaItem *MediaItem) logging.StandardError {
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
