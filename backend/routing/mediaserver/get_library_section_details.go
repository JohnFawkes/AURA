package routes_ms

import (
	"aura/logging"
	"aura/utils/httpx"
	"net/http"
)

func GetLibrarySectionDetails(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Get Library Section Details", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	httpx.SendResponse(w, ld, nil)
}
