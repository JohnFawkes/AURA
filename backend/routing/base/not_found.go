package routes_base

import (
	"aura/logging"
	"aura/utils/httpx"
	"net/http"
)

// NotFound handles requests to undefined routes.
func NotFound(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Route Not Found", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	logAction.SetError("Route Not Found", "Make sure that the route is correct",
		map[string]any{
			"method": r.Method,
			"path":   r.URL.Path,
			"params": r.URL.Query(),
		})
	httpx.SendResponse(w, ld, nil)
}
