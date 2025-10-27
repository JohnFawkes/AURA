package routes_ms

import (
	"aura/internal/api"
	"aura/internal/logging"
	"net/http"
	"net/url"
	"time"

	"github.com/go-chi/chi/v5"
)

func GetItemContent(w http.ResponseWriter, r *http.Request) {
	logging.LOG.Trace(r.URL.Path)
	startTime := time.Now()
	Err := logging.NewStandardError()

	// Get the ratingKey from the URL parameters
	ratingKey := chi.URLParam(r, "ratingKey")
	// Get the sectionTitle from the query parameters
	sectionTitle := r.URL.Query().Get("sectionTitle")
	sectionTitle, _ = url.QueryUnescape(sectionTitle)
	// Get the return type from the query parameters (default to "full")
	returnType := r.URL.Query().Get("returnType")
	if returnType == "" {
		returnType = "full"
	}

	var itemInfo api.MediaItem
	var posterSets []api.PosterSet
	var userFollowHide api.UserFollowHide

	// Validate the rating key
	if ratingKey == "" {
		Err.Message = "Missing rating key in URL"
		Err.HelpText = "Ensure the URL contains a valid rating key."
		Err.Details = map[string]any{
			"error":     "Rating key is empty",
			"ratingKey": ratingKey,
			"request":   r.URL.Path,
		}
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		return
	}

	// Validate the section title
	if sectionTitle == "" {
		Err.Message = "Missing section title in query parameters"
		Err.HelpText = "Ensure the URL contains a valid sectionTitle query parameter."
		Err.Details = map[string]any{
			"error":        "Section title is empty",
			"sectionTitle": sectionTitle,
			"request":      r.URL.Path,
		}
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		return
	}

	itemInfo, Err = api.CallFetchItemContent(ratingKey, sectionTitle)
	if Err.Message != "" {
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		return
	}

	if itemInfo.RatingKey == "" {
		Err.Message = "Item content not found"
		Err.HelpText = "Ensure the rating key is valid and the item exists in the media server."
		Err.Details = map[string]any{
			"error":        "No content found",
			"ratingKey":    ratingKey,
			"sectionTitle": sectionTitle,
			"request":      r.URL.Path,
		}
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		return
	}

	if itemInfo.TMDB_ID == "" {
		Err.Message = "Item missing TMDB ID"
		Err.HelpText = "Ensure the item has a valid TMDB ID in the media server."
		Err.Details = map[string]any{
			"error":        "TMDB ID is empty",
			"ratingKey":    ratingKey,
			"sectionTitle": sectionTitle,
			"request":      r.URL.Path,
		}
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		return
	}

	// Update the cache with the item content
	api.Global_Cache_LibraryStore.UpdateMediaItem(sectionTitle, &itemInfo)

	// If the return type is "mediaitem", return just the item info
	if returnType == "mediaitem" {
		api.Util_Response_SendJson(w, http.StatusOK, api.JSONResponse{
			Status:  "success",
			Elapsed: api.Util_ElapsedTime(startTime),
			Data: map[string]any{
				"serverType": api.Global_Config.MediaServer.Type,
				"mediaItem":  itemInfo,
			},
		})
		return
	}

	// Get the all sets for this TMDB ID
	posterSets, Err = api.Mediux_FetchAllSets(itemInfo.TMDB_ID, itemInfo.Type, itemInfo.LibraryTitle)
	if Err.Message != "" {
		api.Util_Response_SendJson(w, http.StatusOK, api.JSONResponse{
			Status:  "warning",
			Elapsed: api.Util_ElapsedTime(startTime),
			Data: map[string]any{
				"serverType":     api.Global_Config.MediaServer.Type,
				"mediaItem":      itemInfo,
				"posterSets":     posterSets,
				"userFollowHide": userFollowHide,
				"error":          Err,
			},
		})
		return

	}

	if len(posterSets) == 0 {
		Err.Message = "No sets found for the provided TMDB ID and Item Type"
		Err.HelpText = "Ensure the TMDB ID and Item Type are correct and that sets exist for this item."
		Err.Details = map[string]any{
			"tmdbID":         itemInfo.TMDB_ID,
			"itemType":       itemInfo.Type,
			"librarySection": itemInfo.LibraryTitle,
		}
		api.Util_Response_SendJson(w, http.StatusOK, api.JSONResponse{
			Status:  "warning",
			Elapsed: api.Util_ElapsedTime(startTime),
			Data: map[string]any{
				"serverType":     api.Global_Config.MediaServer.Type,
				"mediaItem":      itemInfo,
				"posterSets":     posterSets,
				"userFollowHide": userFollowHide,
				"error":          Err,
			},
		})
		return
	}

	// Fetch user following and hiding data from the Mediux API
	userFollowHide, Err = api.Mediux_FetchUserFollowingAndHiding()
	if Err.Message != "" {
		api.Util_Response_SendJson(w, http.StatusOK, api.JSONResponse{
			Status:  "warning",
			Elapsed: api.Util_ElapsedTime(startTime),
			Data: map[string]any{
				"serverType":     api.Global_Config.MediaServer.Type,
				"mediaItem":      itemInfo,
				"posterSets":     posterSets,
				"userFollowHide": userFollowHide,
				"error":          Err,
			},
		})
		return
	}

	// Respond with a success message
	api.Util_Response_SendJson(w, http.StatusOK, api.JSONResponse{
		Status:  "success",
		Elapsed: api.Util_ElapsedTime(startTime),
		Data: map[string]any{
			"serverType":     api.Global_Config.MediaServer.Type,
			"mediaItem":      itemInfo,
			"posterSets":     posterSets,
			"userFollowHide": userFollowHide,
		},
	})
}
