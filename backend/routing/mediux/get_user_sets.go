package routes_mediux

import (
	"aura/logging"
	"aura/mediux"
	"aura/utils/httpx"
	"net/http"
)

func GetUserSets(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Get Mediux User Sets", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	username := r.URL.Query().Get("username")
	if username == "" {
		logAction.SetError("Username query parameter is required", "", nil)
		httpx.SendResponse(w, ld, nil)
		return
	}

	userSets, Err := mediux.GetAllUserSets(ctx, username)
	if Err.Message != "" {
		httpx.SendResponse(w, ld, nil)
		return
	}

	httpx.SendResponse(w, ld, userSets)
}
