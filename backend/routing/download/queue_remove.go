package routes_download

import (
	downloadqueue "aura/download/queue"
	"aura/logging"
	"aura/models"
	"aura/utils/httpx"
	"fmt"
	"net/http"
)

type RemoveItemFromDownloadQueue_Request struct {
	Item models.DBSavedItem `json:"item"`
}

type RemoveItemFromDownloadQueue_Response struct {
	Result string `json:"result"`
}

// RemoveItemFromQueue godoc
// @Summary      Download Queue - Remove Item
// @Description  Remove a specific Media Item from the download queue. This can be used to cancel pending download tasks or clean up items that are no longer needed in the queue.
// @Tags         Download
// @Accept       json
// @Produce      json
// @Param        req  body      RemoveItemFromDownloadQueue_Request  true  "Queue Remove Item Request"
// @Security 	 BearerAuth
// @Failure      401  {object}  httpx.UnauthorizedResponse "Unauthorized (only when Auth.Enabled=true)"
// @Success      200           {object}  httpx.JSONResponse{data=RemoveItemFromDownloadQueue_Response}
// @Failure      500           {object}  httpx.JSONResponse "Internal Server Error"
// @Router       /api/download/queue/item [delete]
func RemoveItemFromDownloadQueue(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Download Queue - Remove Item", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)
	var req RemoveItemFromDownloadQueue_Request
	var response RemoveItemFromDownloadQueue_Response

	// Parse and validate request body
	Err := httpx.DecodeRequestBodyToJSON(ctx, r.Body, &req, "Queue Remove Item - Decode Request Body")
	if Err.Message != "" {
		httpx.SendResponse(w, ld, response)
		return
	}

	// Validate the JSON structure
	validateAction := logAction.AddSubAction("Validate Delete Item", logging.LevelDebug)
	if req.Item.MediaItem.Title == "" || req.Item.MediaItem.LibraryTitle == "" || req.Item.MediaItem.TMDB_ID == "" || req.Item.MediaItem.RatingKey == "" {
		validateAction.SetError("Invalid Delete Item structure",
			"Ensure that the request body contains a valid Delete Item with all required fields",
			map[string]any{
				"item": req.Item,
			})
		validateAction.Complete()
		httpx.SendResponse(w, ld, response)
		return
	}

	deleted, Err := downloadqueue.RemoveFromQueue(ctx, req.Item)
	if Err.Message != "" {
		httpx.SendResponse(w, ld, response)
		return
	}

	if deleted > 0 {
		logAction.AppendResult("total_deleted", deleted)
	} else {
		logAction.AppendResult("total_deleted", 0)
		logAction.AppendResult("message", "No matching items found in the download queue")
	}

	response.Result = fmt.Sprintf("Removed %d item(s) from the download queue", deleted)
	httpx.SendResponse(w, ld, response)
}
