package routes_ms

import (
	"aura/internal/api"
	"aura/internal/logging"
	"net/http"
)

func GetItemContent(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Get Item Content", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	actionGetQueryParams := logAction.AddSubAction("Get Query Params", logging.LevelDebug)
	ratingKey := r.URL.Query().Get("ratingKey")
	sectionTitle := r.URL.Query().Get("sectionTitle")
	returnType := r.URL.Query().Get("returnType")
	if returnType == "" {
		returnType = "full"
	}
	if ratingKey == "" || sectionTitle == "" || returnType == "" {
		actionGetQueryParams.SetError("Missing Query Parameters", "One or more required query parameters are missing",
			map[string]any{
				"ratingKey":    ratingKey,
				"sectionTitle": sectionTitle,
				"returnType":   returnType,
			})
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}
	actionGetQueryParams.Complete()

	var itemInfo api.MediaItem
	var posterSets []api.PosterSet
	var userFollowHide api.UserFollowHide

	itemInfo, Err := api.CallFetchItemContent(ctx, ratingKey, sectionTitle)
	if Err.Message != "" {
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}

	// Make sure that the Media Item has a Rating Key and TMDB ID
	actionValidateMediaItem := logAction.AddSubAction("Validate Media Item", logging.LevelDebug)
	if itemInfo.RatingKey == "" || itemInfo.TMDB_ID == "" {
		actionValidateMediaItem.SetError("Invalid Media Item", "The media item is missing a Rating Key or TMDB ID",
			map[string]any{
				"ratingKey": itemInfo.RatingKey,
				"tmdbID":    itemInfo.TMDB_ID,
				"title":     itemInfo.Title,
				"library":   itemInfo.LibraryTitle,
			})
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}
	actionValidateMediaItem.Complete()

	// Update the cache with the item content
	api.Global_Cache_LibraryStore.UpdateMediaItem(sectionTitle, &itemInfo)

	// If the return type is "mediaitem", return just the item info
	if returnType == "mediaitem" {
		api.Util_Response_SendJSON(w, ld, map[string]any{
			"serverType": api.Global_Config.MediaServer.Type,
			"mediaItem":  itemInfo,
		})
		return
	}

	// From here on out, we need to get more data (poster sets, user follow/hide)
	// If any of them fail, we just send what we have and make status warning
	// Get the all sets for this TMDB ID
	posterSets, Err = api.Mediux_FetchAllSets(ctx, itemInfo.TMDB_ID, itemInfo.Type, itemInfo.LibraryTitle)
	if Err.Message != "" {
		ld.Status = logging.StatusWarn
		api.Util_Response_SendJSON(w, ld,
			map[string]any{
				"serverType":     api.Global_Config.MediaServer.Type,
				"mediaItem":      itemInfo,
				"posterSets":     posterSets,
				"userFollowHide": userFollowHide,
				"error":          Err,
			})
		return
	}

	// Fetch user following and hiding data from the Mediux API
	userFollowHide, Err = api.Mediux_FetchUserFollowingAndHiding(ctx)
	if Err.Message != "" {
		ld.Status = logging.StatusWarn
		api.Util_Response_SendJSON(w, ld,
			map[string]any{
				"serverType":     api.Global_Config.MediaServer.Type,
				"mediaItem":      itemInfo,
				"posterSets":     posterSets,
				"userFollowHide": userFollowHide,
				"error":          Err,
			})
		return
	}

	// All done, send the full response
	api.Util_Response_SendJSON(w, ld,
		map[string]any{
			"serverType":     api.Global_Config.MediaServer.Type,
			"mediaItem":      itemInfo,
			"posterSets":     posterSets,
			"userFollowHide": userFollowHide,
			"error":          nil,
		})
}
