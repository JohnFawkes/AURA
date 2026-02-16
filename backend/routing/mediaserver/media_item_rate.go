package routes_ms

import (
	"aura/cache"
	"aura/config"
	"aura/logging"
	"aura/mediaserver"
	"aura/utils/httpx"
	"net/http"
	"strconv"
)

func RateMediaItem(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Rate Media Item", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	if config.Current.MediaServer.Type != "Plex" {
		httpx.SendResponse(w, ld, nil)
		return
	}

	// Get the query parameters
	actionGetQueryParams := logAction.AddSubAction("Get Query Parameters", logging.LevelTrace)
	ratingKey := r.URL.Query().Get("rating_key")
	userRatingStr := r.URL.Query().Get("rating")
	if ratingKey == "" || userRatingStr == "" {
		actionGetQueryParams.SetError("Missing query parameters: rating_key or rating", "Make sure to provide a valid rating_key and rating", map[string]any{
			"rating_key": ratingKey,
			"rating":     userRatingStr,
		})
		httpx.SendResponse(w, ld, nil)
		return
	}
	userRating, err := strconv.ParseFloat(userRatingStr, 64)
	if err != nil {
		actionGetQueryParams.SetError("Invalid rating value", "Rating must be a number between 0 and 5", map[string]any{
			"rating": userRatingStr,
		})
		httpx.SendResponse(w, ld, nil)
		return
	}

	if userRating < 0 || userRating > 5 {
		actionGetQueryParams.SetError("Invalid rating value", "Rating must be between 0 and 5", map[string]any{
			"rating": userRating,
		})
		httpx.SendResponse(w, ld, nil)
		return
	}

	// Get the Media Item from the cache
	mediaItem, found := cache.LibraryStore.GetMediaItemByRatingKey(ratingKey)
	if !found {
		actionGetQueryParams.SetError("Media item not found in cache", "Make sure the rating_key is correct and the media server is connected", map[string]any{
			"rating_key": ratingKey,
		})
		httpx.SendResponse(w, ld, nil)
		return
	}

	// Convert rating to scale of 10 for Plex
	userRating = userRating * 2

	Err := mediaserver.RateMediaItem(ctx, mediaItem, userRating)
	if Err.Message != "" {
		httpx.SendResponse(w, ld, nil)
		return
	}

	var response struct {
		Rating float64 `json:"rating"`
		Rated  bool    `json:"rated"`
	}
	response.Rated = true
	response.Rating = userRating

	httpx.SendResponse(w, ld, response)
}
