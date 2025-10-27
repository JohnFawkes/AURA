package routes_config

import (
	"aura/internal/api"
	"aura/internal/logging"
	"net/http"
	"time"
)

func ReloadConfig(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	logging.LOG.Trace(r.URL.Path)

	api.Config_LoadYamlConfig()

	api.Global_Config.PrintDetails()

	safeConfigData := api.Global_Config.Sanitize()

	// (If Logging.File is a pointer or needs refresh, set after clone)
	safeConfigData.Logging.File = logging.GetTodayLogFile()

	api.Util_Response_SendJson(w, http.StatusOK, api.JSONResponse{
		Status:  "success",
		Elapsed: api.Util_ElapsedTime(startTime),
		Data:    safeConfigData,
	})
}
