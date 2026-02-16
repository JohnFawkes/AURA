package routes_base

import (
	"aura/logging"
	"aura/utils/httpx"
	"net/http"
)

func MethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Method Not Allowed", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	logAction.SetError("Method Not Allowed", "Make sure that the HTTP method is correct",
		map[string]any{
			"method": r.Method,
			"path":   r.URL.Path,
			"params": r.URL.Query(),
		})
	httpx.SendResponse(w, ld, nil)
}
