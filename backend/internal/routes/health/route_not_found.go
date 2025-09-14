package route_health

import (
	"aura/internal/logging"
	"aura/internal/utils"
	"net/http"
	"time"
)

func RouteNotFound(w http.ResponseWriter, r *http.Request) {
	logging.LOG.Debug(r.URL.Path)
	startTime := time.Now()

	utils.SendJsonResponse(w, http.StatusInternalServerError, utils.JSONResponse{
		Status:  "error",
		Elapsed: utils.ElapsedTime(startTime),
		Data:    "Route not found",
	})
}
