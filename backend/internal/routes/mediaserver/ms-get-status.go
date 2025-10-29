package routes_ms

import (
	"aura/internal/api"
	"aura/internal/logging"
	"net/http"
)

func GetStatus(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Get Status", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	status, Err := api.CallGetMediaServerStatus(ctx, api.Config_MediaServer{})
	if Err.Message != "" {
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}

	api.Util_Response_SendJSON(w, ld, status)
}
