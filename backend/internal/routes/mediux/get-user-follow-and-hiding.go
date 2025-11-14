package routes_mediux

import (
	"aura/internal/api"
	"aura/internal/logging"
	"net/http"
)

func GetUserFollowingAndHiding(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Get User Following And Hiding", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	// Fetch user following and hiding data from the MediUX API
	userFollowHide, Err := api.Mediux_FetchUserFollowingAndHiding(ctx)
	if Err.Message != "" {
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}

	api.Util_Response_SendJSON(w, ld, userFollowHide)
}
