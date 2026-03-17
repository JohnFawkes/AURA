package routes_download

import (
	downloadqueue "aura/download/queue"
	"aura/logging"
	"aura/models"
	"aura/utils/httpx"
	"net/http"
)

type GetAllDownloadQueueItems_Response struct {
	InProgressEntries []models.DBSavedItem `json:"in_progress_entries"`
	WarningEntries    []models.DBSavedItem `json:"warning_entries"`
	ErrorEntries      []models.DBSavedItem `json:"error_entries"`
}

// GetAllDownloadQueueItems godoc
// @Summary      Download Queue - Get Items
// @Description  Retrieve the current items in the download queue, categorized by their status (in-progress, warning, error). This endpoint allows clients to monitor the progress of queued download tasks and identify any issues that may have occurred during processing.
// @Tags         Download
// @Accept       json
// @Produce      json
// @Security 	 BearerAuth
// @Failure      401  {object}  httpx.UnauthorizedResponse "Unauthorized (only when Auth.Enabled=true)"
// @Success      200  {object}  httpx.JSONResponse{data=GetAllDownloadQueueItems_Response}
// @Failure      500           {object}  httpx.JSONResponse "Internal Server Error"
// @Router       /api/download/queue/item [get]
func GetAllDownloadQueueItems(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Download Queue - Get Items", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)
	var response GetAllDownloadQueueItems_Response

	inProgressItems, warningItems, errorItems, Err := downloadqueue.GetQueueItems(ctx)
	if Err.Message != "" {
		httpx.SendResponse(w, ld, response)
		return
	}

	response.InProgressEntries = inProgressItems
	response.WarningEntries = warningItems
	response.ErrorEntries = errorItems
	httpx.SendResponse(w, ld, response)
}
