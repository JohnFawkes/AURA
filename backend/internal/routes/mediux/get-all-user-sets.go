package routes_mediux

import (
	"aura/internal/api"
	"aura/internal/logging"
	"net/http"
)

func GetAllUserSets(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Get All User Sets", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	// Get user ID from URL
	actionGetQueryParams := ld.AddAction("Get all query params", logging.LevelTrace)
	username := r.URL.Query().Get("username")
	if username == "" {
		actionGetQueryParams.SetError("Missing Query Parameters", "One or more required query parameters are missing",
			map[string]any{
				"username": username,
			})
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}
	actionGetQueryParams.Complete()

	allSetsResponse, Err := api.Mediux_FetchAllUserSets(ctx, username)
	if Err.Message != "" {
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}

	api.Util_Response_SendJSON(w, ld, allSetsResponse.Data)
}
