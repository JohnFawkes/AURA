package routes_base

import (
	"aura/config"
	"aura/logging"
	"aura/utils/httpx"
	"net/http"
)

type healthCheckResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

// HealthCheck godoc
// @Summary      Health Check
// @Description  Check the health status of the application
// @Tags         Health
// @Produce      json
// @Success      200  {object}  routes_base.healthCheckResponse
// @Router       /api/ [get]
// @Router       /api/health [get]
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	_, ld := logging.CreateLoggingContext(r.Context(), "Health Check")
	ld.Status = logging.StatusSuccess

	var response healthCheckResponse
	response.Status = "ok"
	response.Version = config.Version
	httpx.SendResponse(w, ld, response)
}
