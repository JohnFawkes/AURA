package routes_config

import (
	"aura/internal/api"
	"aura/internal/logging"
	"net/http"
)

func GetSanitizedConfig(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Get Sanitized Config", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	safeConfigData := api.Global_Config.Sanitize(ctx)

	api.Util_Response_SendJSON(w, ld, safeConfigData)
}
