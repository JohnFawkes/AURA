package routes_ms

import (
	"aura/internal/api"
	"net/http"
	"time"
)

func GetType(w http.ResponseWriter, r *http.Request) {
	// Get the start time
	startTime := time.Now()

	api.Util_Response_SendJson(w, http.StatusOK, api.JSONResponse{
		Status:  "success",
		Elapsed: api.Util_ElapsedTime(startTime),
		Data:    map[string]any{"serverType": api.Global_Config.MediaServer.Type},
	})
}
