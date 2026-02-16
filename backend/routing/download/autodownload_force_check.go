package routes_download

import (
	autodownload "aura/download/auto"
	"aura/logging"
	"aura/mediaserver"
	"aura/mediux"
	"aura/models"
	"aura/utils/httpx"
	"net/http"
)

func AutoDownloadForceCheck(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Auto Download - Force Check", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	var req struct {
		Item     models.DBSavedItem
		Complete bool
	}
	Err := httpx.DecodeRequestBodyToJSON(ctx, r.Body, &req, "Autodownload Force Check - Decode Request Body")
	if Err.Message != "" {
		httpx.SendResponse(w, ld, nil)
		return
	}

	// Validate the Save Item Media Item
	if req.Item.MediaItem.TMDB_ID == "" || req.Item.MediaItem.LibraryTitle == "" {
		logAction.SetError("Invalid Media Item Data", "TMDB ID and Library Title are required", map[string]any{
			"tmdb_id":       req.Item.MediaItem.TMDB_ID,
			"library_title": req.Item.MediaItem.LibraryTitle,
		})
		httpx.SendResponse(w, ld, nil)
		return
	}

	// Make sure each Poster Set has an ID and Type
	for _, posterSet := range req.Item.PosterSets {
		if posterSet.ID == "" || posterSet.Type == "" {
			logAction.SetError("Invalid Poster Set Data", "Each Poster Set must have an ID and Type", map[string]any{
				"tmdb_id":       req.Item.MediaItem.TMDB_ID,
				"library_title": req.Item.MediaItem.LibraryTitle,
				"set_id":        posterSet.ID,
				"set_type":      posterSet.Type,
			})
			httpx.SendResponse(w, ld, nil)
			return
		}
	}

	// If the request body is not complete, we need to get the information again
	// This is to prevent incomplete data from being used in the check
	// In the frontend, we omit MediaItem details (besides important fields)
	// and ImageFiles in PosterSets to reduce payload size
	fullSets := []models.DBPosterSetDetail{}
	if !req.Complete {
		// Get the MediaItem details from the Media Server
		found, Err := mediaserver.GetMediaItemDetails(ctx, &req.Item.MediaItem)
		if Err.Message != "" {
			httpx.SendResponse(w, ld, nil)
			return
		}
		if !found {
			logAction.SetError("Media Item Not Found On Media Server", "The provided media item could not be found on the media server. You can try and run the checking for Rating Key Changes Job and then try again.", map[string]any{
				"tmdb_id":       req.Item.MediaItem.TMDB_ID,
				"library_title": req.Item.MediaItem.LibraryTitle,
			})
			httpx.SendResponse(w, ld, nil)
			return
		}

		for _, posterSet := range req.Item.PosterSets {
			switch posterSet.Type {
			case "show":
				showSet, _, Err := mediux.GetShowSetByID(ctx, posterSet.ID, req.Item.MediaItem.LibraryTitle)
				if Err.Message != "" {
					httpx.SendResponse(w, ld, nil)
					return
				}
				posterSet.PosterSet = showSet.PosterSet
				fullSets = append(fullSets, posterSet)
			case "movie":
				movieSet, _, Err := mediux.GetMovieSetByID(ctx, posterSet.ID, req.Item.MediaItem.LibraryTitle)
				if Err.Message != "" {
					httpx.SendResponse(w, ld, nil)
					return
				}
				posterSet.PosterSet = movieSet.PosterSet
				fullSets = append(fullSets, posterSet)
			case "collection":
				collectionSet, _, Err := mediux.GetMovieCollectionSetByID(ctx, posterSet.ID, req.Item.MediaItem.TMDB_ID, req.Item.MediaItem.LibraryTitle)
				if Err.Message != "" {
					httpx.SendResponse(w, ld, nil)
					return
				}
				posterSet.PosterSet = collectionSet.PosterSet
				fullSets = append(fullSets, posterSet)
			default:
				logAction.SetError("Invalid Poster Set Type", "Poster Set type must be either 'show' or 'movie'", map[string]any{
					"set_type": posterSet.Type,
				})
				httpx.SendResponse(w, ld, nil)
				return
			}
		}
	} else {
		fullSets = req.Item.PosterSets
	}

	saveItem := models.DBSavedItem{
		MediaItem:  req.Item.MediaItem,
		PosterSets: fullSets,
	}

	// Perform the Force Check
	result := autodownload.CheckItem(ctx, saveItem)

	httpx.SendResponse(w, ld, result)
}
