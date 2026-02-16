package routes_download

import (
	downloadqueue "aura/download/queue"
	"aura/logging"
	"aura/models"
	"aura/utils/httpx"
	"net/http"
)

func QueueAddItem(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Download Queue - Add Item", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	// Parse and validate request body
	var saveItem models.DBSavedItem
	Err := httpx.DecodeRequestBodyToJSON(ctx, r.Body, &saveItem, "Queue Add Item - Decode Request Body")
	if Err.Message != "" {
		httpx.SendResponse(w, ld, nil)
		return
	}

	// Validate the JSON structure
	validateAction := logAction.AddSubAction("Validate Save Item", logging.LevelDebug)
	if saveItem.MediaItem.Title == "" || saveItem.MediaItem.LibraryTitle == "" || saveItem.MediaItem.TMDB_ID == "" || saveItem.MediaItem.RatingKey == "" {
		validateAction.SetError("Invalid Save Item structure",
			"Ensure that the request body contains a valid Save Item with all required fields",
			map[string]any{
				"item": saveItem,
			})
		validateAction.Complete()
		httpx.SendResponse(w, ld, nil)
		return
	}

	// Make sure there is at least one Poster Set
	if len(saveItem.PosterSets) == 0 {
		validateAction.SetError("Invalid Save Item structure - No Poster Sets",
			"Ensure that the Media Item contains at least one Poster Set",
			map[string]any{
				"item": saveItem,
			})
		validateAction.Complete()
		httpx.SendResponse(w, ld, nil)
		return
	}

	// Make sure that each Poster Set has a Set ID
	for _, posterSet := range saveItem.PosterSets {
		if posterSet.ID == "" {
			validateAction.SetError("Invalid Save Item structure - Poster Set missing Set ID",
				"Ensure that each Poster Set in the Media Item contains a valid Set ID",
				map[string]any{
					"set": posterSet,
				})
			validateAction.Complete()
			httpx.SendResponse(w, ld, nil)
			return
		}
	}
	validateAction.Complete()

	// Add the item to the download queue
	addErr := downloadqueue.AddToQueue(ctx, saveItem)
	if addErr.Message != "" {
		httpx.SendResponse(w, ld, nil)
		return
	}

	httpx.SendResponse(w, ld, "ok")
}
