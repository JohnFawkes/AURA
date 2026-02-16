package routes_mediux

import (
	"aura/cache"
	"aura/logging"
	"aura/mediux"
	"aura/models"
	"aura/utils/httpx"
	"net/http"
)

func GetItemSets(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Get Mediux Item Sets", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	actionGetQueryParams := ld.AddAction("Get all query params", logging.LevelTrace)
	tmdbID := r.URL.Query().Get("tmdb_id")
	itemType := r.URL.Query().Get("item_type")
	itemLibraryTitle := r.URL.Query().Get("item_library_title")
	if tmdbID == "" || itemType == "" || itemLibraryTitle == "" {
		actionGetQueryParams.SetError("Missing Query Parameters", "One or more required query parameters are missing",
			map[string]any{
				"tmdb_id":            tmdbID,
				"item_type":          itemType,
				"item_library_title": itemLibraryTitle,
			})
		httpx.SendResponse(w, ld, nil)
		return
	}
	actionGetQueryParams.Complete()

	// If cache is empty, return false
	actionCheckCache := ld.AddAction("Check library cache", logging.LevelTrace)
	if cache.LibraryStore.IsEmpty() {
		actionCheckCache.SetError("Library Cache Empty", "The library cache is empty, cannot fetch sets", nil)
		httpx.SendResponse(w, ld, nil)
		return
	}
	actionCheckCache.Complete()

	var response models.PosterSetsResponse

	switch itemType {
	case "show":
		// For Shows, we get just Show Sets
		showSets, showItems, Err := mediux.GetShowItemSets(ctx, tmdbID, itemLibraryTitle)
		if Err.Message != "" {
			httpx.SendResponse(w, ld, nil)
			return
		}
		response.Sets = showSets
		response.IncludedItems = showItems
	case "movie":
		setItems := map[string]models.IncludedItem{}
		// For Movies, we get Movie Sets and Movie Collection Sets
		movieSets, Err := mediux.GetMovieItemSets(ctx, tmdbID, itemLibraryTitle, &setItems)
		if Err.Message != "" {
			httpx.SendResponse(w, ld, nil)
			return
		}
		response.Sets = movieSets
		collectionSets, Err := mediux.GetMovieItemCollectionSets(ctx, tmdbID, itemLibraryTitle, &setItems)
		if Err.Message != "" {
			httpx.SendResponse(w, ld, nil)
			return
		}
		response.Sets = append(response.Sets, collectionSets...)
		if len(movieSets) == 0 && len(collectionSets) == 0 && len(setItems) == 0 {
			logAction.SetError("No Sets Found", "No movie sets or collection sets found for the provided TMDB ID", map[string]any{
				"tmdb_id": tmdbID,
			})
			httpx.SendResponse(w, ld, nil)
			return
		}
		response.IncludedItems = setItems
	default:
		logAction.SetError("Invalid Item Type", "The provided item type is not valid", map[string]any{
			"item_type": itemType,
		})
		httpx.SendResponse(w, ld, nil)
		return
	}

	httpx.SendResponse(w, ld, response)
}
