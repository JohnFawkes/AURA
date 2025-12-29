package routes_ms

import (
	"aura/internal/api"
	"aura/internal/logging"
	"net/http"
)

func RateMediaItem(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Rate Media Item", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	type RateMediaItemRequest struct {
		RatingKey  string  `json:"ratingKey"`
		UserRating float64 `json:"userRating"`
	}

	var req RateMediaItemRequest
	if err := api.DecodeRequestBodyJSON(ctx, r.Body, &req, "rateMediaItemRequest"); err.Message != "" {
		logAction.SetError("Failed to decode request body", "", nil)
		ld.Status = logging.StatusError
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}

	if req.RatingKey == "" {
		logAction.SetError("Item ID is required", "", nil)
		ld.Status = logging.StatusError
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}

	if req.UserRating < 0 || req.UserRating > 5 {
		logAction.SetError("Rating must be between 0 and 5", "", nil)
		ld.Status = logging.StatusError
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}

	// Convert rating to scale of 10 for Plex
	req.UserRating = req.UserRating * 2

	err := api.Plex_RateMediaItem(ctx, req.RatingKey, req.UserRating)
	if err.Message != "" {
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}

	api.Util_Response_SendJSON(w, ld, map[string]any{
		"ratingKey":  req.RatingKey,
		"userRating": req.UserRating,
	})
}
