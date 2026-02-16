package routes_db

import (
	"aura/database"
	"aura/logging"
	"aura/models"
	"aura/utils/httpx"
	"net/http"
)

func UpdateItem(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Update Item In Database", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	type UpdateItemRequest struct {
		Complete   bool               `json:"complete"`
		UpdateItem models.DBSavedItem `json:"update_item"`
	}

	// Parse the request body
	var req UpdateItemRequest
	Err := httpx.DecodeRequestBodyToJSON(ctx, r.Body, &req, "Update Item - Decode Request Body")
	if Err.Message != "" {
		httpx.SendResponse(w, ld, nil)
		return
	}

	for _, ps := range req.UpdateItem.PosterSets {
		if ps.ToDelete {
			// Delete the poster set
			Err := database.DeletePosterSetForMediaItem(ctx, req.UpdateItem.MediaItem.TMDB_ID, req.UpdateItem.MediaItem.LibraryTitle, ps.ID)
			if Err.Message != "" {
				httpx.SendResponse(w, ld, nil)
				return
			}
		} else {
			// Upsert the poster set
			Err := database.UpsertSavedItem(ctx, req.UpdateItem)
			if Err.Message != "" {
				httpx.SendResponse(w, ld, nil)
				return
			}
		}
	}

	httpx.SendResponse(w, ld, "ok")
}
