package routes_images

import (
	"aura/logging"
	"aura/utils/httpx"
	"net/http"
)

func ClearCachedImages(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Clear Cached Images", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	httpx.SendResponse(w, ld, nil)
}
