package routes_download

import (
	downloadqueue "aura/download/queue"
	"aura/logging"
	"aura/models"
	"aura/utils/httpx"
	"fmt"
	"net/http"
)

func QueueRemoveItem(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Download Queue - Remove Item", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	// Parse and validate request body
	var deleteItem models.DBSavedItem
	Err := httpx.DecodeRequestBodyToJSON(ctx, r.Body, &deleteItem, "Queue Remove Item - Decode Request Body")
	if Err.Message != "" {
		httpx.SendResponse(w, ld, nil)
		return
	}

	// Validate the JSON structure
	validateAction := logAction.AddSubAction("Validate Delete Item", logging.LevelDebug)
	if deleteItem.MediaItem.Title == "" || deleteItem.MediaItem.LibraryTitle == "" || deleteItem.MediaItem.TMDB_ID == "" || deleteItem.MediaItem.RatingKey == "" {
		validateAction.SetError("Invalid Delete Item structure",
			"Ensure that the request body contains a valid Delete Item with all required fields",
			map[string]any{
				"item": deleteItem,
			})
		validateAction.Complete()
		httpx.SendResponse(w, ld, nil)
		return
	}

	deleted, Err := downloadqueue.RemoveFromQueue(ctx, deleteItem)
	if Err.Message != "" {
		httpx.SendResponse(w, ld, nil)
		return
	}

	if deleted > 0 {
		logAction.AppendResult("total_deleted", deleted)
	} else {
		logAction.AppendResult("total_deleted", 0)
		logAction.AppendResult("message", "No matching items found in the download queue")
	}

	httpx.SendResponse(w, ld, fmt.Sprintf("Removed %d item(s) from the download queue", deleted))
}
