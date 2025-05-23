package health

import (
	"aura/internal/logging"
	"aura/internal/utils"
	"net/http"
	"time"
)

func NotFound(w http.ResponseWriter, r *http.Request) {
	logging.LOG.Debug(r.URL.Path)
	startTime := time.Now()

	utils.SendJsonResponse(w, http.StatusInternalServerError, utils.JSONResponse{
		Status:  "error",
		Message: "Route not found",
		Elapsed: utils.ElapsedTime(startTime),
	})
}
