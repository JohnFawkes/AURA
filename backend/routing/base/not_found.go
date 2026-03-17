package routes_base

import (
	"aura/logging"
	"aura/utils/httpx"
	"net/http"
)

// RouteNotFound godoc
// @Summary      Route Not Found Handler
// @Description  Handle requests to routes that are not defined in the application
// @Tags         Base
// @Produce      json
// @Failure      500  {object}  httpx.JSONResponse "Internal Server Error"
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
