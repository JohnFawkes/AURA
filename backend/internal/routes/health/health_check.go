package route_health

import (
	"aura/internal/utils"
	"net/http"
	"time"
)

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	// Get the start time
	startTime := time.Now()

	utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
		Status:  "success",
		Elapsed: utils.ElapsedTime(startTime),
		Data:    "Server is online",
	})
}
