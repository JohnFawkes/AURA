package routes_config

import (
	"aura/internal/api"
	"aura/internal/logging"
	"net/http"
)

func ReloadConfig(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Reload Config", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	api.Config_LoadYamlConfig(ctx)

	api.Global_Config.PrintDetails()

	safeConfigData := api.Global_Config.Sanitize(ctx)

	api.Util_Response_SendJSON(w, ld, safeConfigData)
}
