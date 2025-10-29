package routes_api

import (
	"aura/internal/api"
	"aura/internal/logging"
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
	api.Util_Response_SendJSON(w, ld, nil)
}

// MethodNotAllowed handles requests with invalid HTTP methods.
func MethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Method Not Allowed", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	logAction.SetError("Method Not Allowed", "Make sure that the HTTP method is correct", map[string]any{
		"method": r.Method,
		"path":   r.URL.Path,
		"params": r.URL.Query(),
	})

	api.Util_Response_SendJSON(w, ld, nil)
}
