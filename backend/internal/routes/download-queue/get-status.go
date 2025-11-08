package routes_download_queue

import (
	"aura/internal/api"
	"aura/internal/logging"
	"net/http"
)

func GetDownloadQueueLastStatus(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Get Download Queue Status", logging.LevelTrace)
	ctx = logging.WithCurrentAction(ctx, logAction)

	api.Util_Response_SendJSON(w, ld,
		map[string]any{
			"time":     api.DOWNLOAD_QUEUE_LATEST_INFO.Time,
			"status":   api.DOWNLOAD_QUEUE_LATEST_INFO.Status,
			"message":  api.DOWNLOAD_QUEUE_LATEST_INFO.Message,
			"warnings": api.DOWNLOAD_QUEUE_LATEST_INFO.Warnings,
			"errors":   api.DOWNLOAD_QUEUE_LATEST_INFO.Errors,
		})
}
