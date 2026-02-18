package routes_base

import (
	"aura/logging"
	"aura/utils/httpx"
	"net/http"
)

// MethodNotAllowed godoc
// @Summary      Method Not Allowed Handler
// @Description  Handle requests with HTTP methods that are not allowed for the endpoint
// @Tags         Base
// @Produce      json
// @Failure      500  {object}  httpx.JSONResponse "Internal Server Error"
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
