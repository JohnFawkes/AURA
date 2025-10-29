package routes_ms

import (
	"aura/internal/api"
	"aura/internal/logging"
	"net/http"
)

func GetType(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Get Media Server Type", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)
	logAction.Complete()
	api.Util_Response_SendJSON(w, ld, map[string]any{"serverType": api.Global_Config.MediaServer.Type})
}
