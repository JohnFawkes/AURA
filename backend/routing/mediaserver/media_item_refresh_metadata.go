package routes_ms

import (
	"aura/cache"
	"aura/logging"
	"aura/mediaserver"
	"aura/utils/httpx"
	"net/http"
)

func RefreshMediaItemMetadata(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Refresh Media Item Metadata", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	// Get the query parameters
	actionGetQueryParams := logAction.AddSubAction("Get Query Parameters", logging.LevelTrace)
	ratingKey := r.URL.Query().Get("rating_key")
	refreshRatingKey := r.URL.Query().Get("refresh_rating_key")
	if ratingKey == "" || refreshRatingKey == "" {
		actionGetQueryParams.SetError("Query parameter 'rating_key' and 'refresh_rating_key' are required", "Make sure to provide valid rating_key and refresh_rating_key", map[string]any{
			"rating_key":         ratingKey,
			"refresh_rating_key": refreshRatingKey,
		})
		httpx.SendResponse(w, ld, nil)
		return
	}

	// Get the Media Item from the cache
	mediaItem, found := cache.LibraryStore.GetMediaItemByRatingKey(ratingKey)
	if !found {
		actionGetQueryParams.SetError("Media item not found in cache",
			"Make sure the rating_key is correct and the media server is connected",
			map[string]any{
				"rating_key": ratingKey,
			})
		httpx.SendResponse(w, ld, nil)
		return
	}

	Err := mediaserver.RefreshMediaItemMetadata(ctx, mediaItem, refreshRatingKey, true)
	if Err.Message != "" {
		httpx.SendResponse(w, ld, nil)
		return
	}

	var response struct {
		Refreshed bool `json:"refreshed"`
	}
	response.Refreshed = true

	httpx.SendResponse(w, ld, response)
}
