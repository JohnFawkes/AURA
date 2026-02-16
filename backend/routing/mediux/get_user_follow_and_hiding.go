package routes_mediux

import (
	"aura/logging"
	"aura/mediux"
	"aura/utils/httpx"
	"net/http"
)

func GetUserFollowingAndHiding(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Get MediUX User Following and Hiding", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	users, Err := mediux.GetUserFollowingAndHiding(ctx)
	if Err.Message != "" {
		httpx.SendResponse(w, ld, nil)
		return
	}

	httpx.SendResponse(w, ld, users)
}
