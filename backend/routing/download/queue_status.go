package routes_download

import (
	downloadqueue "aura/download/queue"
	"aura/logging"
	"aura/utils/httpx"
	"net/http"
)

func QueueGetStatus(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Download Queue - Get Status", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	httpx.SendResponse(w, ld, map[string]any{
		"time":     downloadqueue.LatestInfo.Time,
		"status":   downloadqueue.LatestInfo.Status,
		"message":  downloadqueue.LatestInfo.Message,
		"warnings": downloadqueue.LatestInfo.Warnings,
		"errors":   downloadqueue.LatestInfo.Errors,
	})
}
