package routes_images

import (
	"aura/cache"
	"aura/logging"
	"aura/mediaserver"
	"aura/utils/httpx"
	"net/http"
)

func GetMediaItemImage(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Get Media Item Image From Media Server", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	actionGetQueryParams := logAction.AddSubAction("Get Query Params", logging.LevelDebug)
	ratingKey := r.URL.Query().Get("rating_key")
	imageRatingKey := r.URL.Query().Get("image_rating_key")
	imageType := r.URL.Query().Get("image_type")
	if ratingKey == "" || imageType == "" {
		actionGetQueryParams.SetError("Missing Query Parameters", "One or more required query parameters are missing",
			map[string]any{
				"rating_key":       ratingKey,
				"image_rating_key": imageRatingKey,
				"image_type":       imageType,
			})
		httpx.SendResponse(w, ld, nil)
		return
	} else if imageType != "poster" && imageType != "backdrop" && imageType != "thumb" {
		actionGetQueryParams.SetError("Invalid Query Parameters", "Image type must be either 'poster', 'backdrop', or 'thumb'",
			map[string]any{
				"rating_key":       ratingKey,
				"image_rating_key": imageRatingKey,
				"image_type":       imageType,
			})
		httpx.SendResponse(w, ld, nil)
		return
	}
	if imageRatingKey == "" {
		imageRatingKey = ratingKey
	}
	actionGetQueryParams.Complete()

	// Get the matching media item from the cache
	item, found := cache.LibraryStore.GetMediaItemByRatingKey(ratingKey)
	if !found {
		logAction.SetError("Media Item Not Found", "No media item found matching the provided rating key",
			map[string]any{
				"rating_key": ratingKey,
			})
		httpx.SendResponse(w, ld, nil)
		return
	}

	// If the image does not exist, then get it from the media server
	imageData, Err := mediaserver.GetMediaItemImage(ctx, item, imageRatingKey, imageType)
	if Err.Message != "" {
		httpx.SendResponse(w, ld, nil)
		return
	}

	// Set the content type for the response
	w.Header().Set("Content-Type", "image/jpeg")
	// Write the image data to the response
	w.WriteHeader(http.StatusOK)
	w.Write(imageData)
}

func GetCollectionItemImage(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Get Collection Item Image From Media Server", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	actionGetQueryParams := logAction.AddSubAction("Get Query Params", logging.LevelDebug)
	ratingKey := r.URL.Query().Get("rating_key")
	imageType := r.URL.Query().Get("image_type")
	if ratingKey == "" || imageType == "" {
		actionGetQueryParams.SetError("Missing Query Parameters", "One or more required query parameters are missing",
			map[string]any{
				"rating_key": ratingKey,
				"image_type": imageType,
			})
		httpx.SendResponse(w, ld, nil)
		return
	} else if imageType != "poster" && imageType != "backdrop" {
		actionGetQueryParams.SetError("Invalid Query Parameters", "Image type must be either 'poster' or 'backdrop'",
			map[string]any{
				"rating_key": ratingKey,
				"image_type": imageType,
			})
		httpx.SendResponse(w, ld, nil)
		return
	}
	actionGetQueryParams.Complete()

	// Get the matching collection item from the cache
	item, found := cache.CollectionsStore.GetCollectionByRatingKey(ratingKey)
	if !found {
		logAction.SetError("Collection Item Not Found", "No collection item found matching the provided rating key in the cache",
			map[string]any{
				"rating_key": ratingKey,
			})
		httpx.SendResponse(w, ld, nil)
		return
	}

	imageData, Err := mediaserver.GetCollectionItemImage(ctx, item, imageType)
	if Err.Message != "" {
		httpx.SendResponse(w, ld, nil)
		return
	}

	// Set the content type for the response
	w.Header().Set("Content-Type", "image/jpeg")
	// Write the image data to the response
	w.WriteHeader(http.StatusOK)
	w.Write(imageData)
}
