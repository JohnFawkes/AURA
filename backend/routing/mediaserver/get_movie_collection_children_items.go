package routes_ms

import (
	"aura/cache"
	"aura/config"
	"aura/logging"
	"aura/mediaserver"
	"aura/mediux"
	"aura/models"
	"aura/utils/httpx"
	"net/http"
)

func GetAllCollectionChildrenItems(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Get All Movie Collection Children Items And Posters", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	actionGetQueryParams := logAction.AddSubAction("Get Query Params", logging.LevelTrace)
	ratingKey := r.URL.Query().Get("rating_key")
	if ratingKey == "" {
		actionGetQueryParams.SetError("Missing query parameter: rating_key", "Make sure to provide a valid rating_key", nil)
		httpx.SendResponse(w, ld, nil)
		return
	}

	var response struct {
		CollectionItem models.CollectionItem   `json:"collection_item"`
		Sets           []models.SetRef         `json:"sets"`
		UserFollowHide []models.MediuxUserInfo `json:"user_follow_hide"`
	}

	// Get the Collection Item from the cache
	collectionItem, found := cache.CollectionsStore.GetCollectionByRatingKey(ratingKey)
	if !found {
		actionGetQueryParams.SetError("Collection item not found in cache", "Make sure the rating_key is correct and the media server is connected", map[string]any{
			"rating_key": ratingKey,
		})
		httpx.SendResponse(w, ld, nil)
		return
	}

	// Get all children items for the collection
	Err := mediaserver.GetCollectionChildrenItems(ctx, collectionItem)
	if Err.Message != "" {
		httpx.SendResponse(w, ld, nil)
		return
	}
	response.CollectionItem = *collectionItem

	// Get MediUX user follow/hide info
	userFollowHide, Err := mediux.GetUserFollowingAndHiding(ctx)
	if Err.Message != "" {
		logAction.Status = logging.StatusWarn
	} else {
		response.UserFollowHide = userFollowHide
	}
	response.UserFollowHide = userFollowHide

	// Get Poster Sets for the collection items
	// For Plex Servers, we need to get an array of unique TMDB IDs from the collection items
	// For Emby/Jellyfin, we can use the existing TMDB IDs in the collection items
	switch config.Current.MediaServer.Type {
	case "Plex":
		uniqueMovieIDs := make([]string, 0)
		for _, child := range collectionItem.MediaItems {
			if child.Type == "movie" && child.TMDB_ID != "" {
				uniqueMovieIDs = append(uniqueMovieIDs, child.TMDB_ID)
			}
		}
		response.Sets, Err = mediux.GetCollectionImagesByMovieTMDBIDs(ctx, uniqueMovieIDs)
		if Err.Message != "" {
			logAction.Status = logging.StatusWarn
			break
		}

	case "Emby", "Jellyfin":
		response.Sets, Err = mediux.GetCollectionImagesByTMDBID(ctx, collectionItem.TMDB_ID)
		if Err.Message != "" {
			logAction.Status = logging.StatusWarn
			break
		}

	default:
		logAction.SetError("Unsupported Media Server Type", "The media server type is not supported for fetching collection items", map[string]any{
			"server_type": config.Current.MediaServer.Type,
		})
		httpx.SendResponse(w, ld, nil)
		return
	}

	// Send the response with the collection items
	httpx.SendResponse(w, ld, response)
}
