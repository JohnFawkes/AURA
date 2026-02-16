package routes_download

import (
	downloadqueue "aura/download/queue"
	"aura/logging"
	"aura/utils/httpx"
	"net/http"
)

func QueueGetItems(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Download Queue - Get Items", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	inProgressItems, warningItems, errorItems, Err := downloadqueue.GetQueueItems(ctx)
	if Err.Message != "" {
		httpx.SendResponse(w, ld, nil)
		return
	}

	responseData := map[string]any{
		"in_progress_entries": inProgressItems,
		"warning_entries":     warningItems,
		"error_entries":       errorItems,
	}

	httpx.SendResponse(w, ld, responseData)
}
