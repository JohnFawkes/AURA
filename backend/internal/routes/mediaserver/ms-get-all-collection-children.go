package routes_ms

import (
	"aura/internal/api"
	"aura/internal/logging"
	"net/http"
)

func GetAllCollectionChildren(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Get All Collection Children", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	// Get the media server information from the request
	var collectionItem api.CollectionItem
	Err := api.DecodeRequestBodyJSON(ctx, r.Body, &collectionItem, "collectionItem")
	if Err.Message != "" {
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}

	// Validate the collection rating key
	if collectionItem.RatingKey == "" {
		logAction.SetError("Missing Query Parameters", "Collection Rating Key is missing",
			map[string]any{
				"collectionItem": collectionItem,
			})
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}

	// Fetch the collection children from the media server
	Err = api.CallFetchCollectionChildren(ctx, &collectionItem)
	if Err.Message != "" {
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}

	var allCollectionSets []api.CollectionSet
	var userFollowHide api.UserFollowHide

	// From here on out, we need to get more data (collections sets, user follow/hide)
	// If any of them fail, we just send what we have and make status warning

	// Fetch user following and hiding data from the Mediux API
	userFollowHide, Err = api.Mediux_FetchUserFollowingAndHiding(ctx)
	if Err.Message != "" {
		ld.Status = logging.StatusWarn
		api.Util_Response_SendJSON(w, ld,
			map[string]any{
				"collection_item":  collectionItem,
				"collection_sets":  allCollectionSets,
				"user_follow_hide": userFollowHide,
				"error":            Err,
			})
		return
	}

	uniqueMovieIDs := make([]string, 0)
	for _, child := range collectionItem.MediaItems {
		if child.Type == "movie" && child.TMDB_ID != "" {
			uniqueMovieIDs = append(uniqueMovieIDs, child.TMDB_ID)
		}
	}

	switch api.Global_Config.MediaServer.Type {
	case "Plex":

		// Fetch the images for this collection item
		allCollectionSets, Err = api.Mediux_FetchAllCollectionImagesByMovieIDs(ctx, uniqueMovieIDs)
		if Err.Message != "" {
			ld.Status = logging.StatusWarn
			api.Util_Response_SendJSON(w, ld,
				map[string]any{
					"collection_item":  collectionItem,
					"collection_sets":  []api.CollectionSet{},
					"user_follow_hide": userFollowHide,
					"error":            Err,
				})
			return
		}

	case "Emby", "Jellyfin":
		allCollectionSets, Err = api.Mediux_FetchAllCollectionImagesByTMDBID(ctx, collectionItem.TMDBID)
		if Err.Message != "" {
			ld.Status = logging.StatusWarn
			api.Util_Response_SendJSON(w, ld,
				map[string]any{
					"collection_item":  collectionItem,
					"collection_sets":  []api.CollectionSet{},
					"user_follow_hide": userFollowHide,
					"error":            Err,
				})
			return
		}
	}

	// Finally, send the response
	api.Util_Response_SendJSON(w, ld, map[string]any{
		"collection_item":  collectionItem,
		"collection_sets":  allCollectionSets,
		"user_follow_hide": userFollowHide,
		"error":            nil,
	})
}
