package route_config

import (
	"aura/internal/config"
	"aura/internal/utils"
	"net/http"
	"time"
)

func GetMediaServerType(w http.ResponseWriter, r *http.Request) {
	// Get the start time
	startTime := time.Now()

	utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
		Status:  "success",
		Elapsed: utils.ElapsedTime(startTime),
		Data:    map[string]any{"serverType": config.Global.MediaServer.Type},
	})
}
