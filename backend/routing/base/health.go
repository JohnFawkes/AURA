package routes_base

import (
	"aura/config"
	"aura/logging"
	"aura/utils/httpx"
	"net/http"
)

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	_, ld := logging.CreateLoggingContext(r.Context(), "Health Check")
	ld.Status = logging.StatusSuccess

	var response struct {
		Status  string `json:"status"`
		Version string `json:"version"`
	}
	response.Status = "ok"
	response.Version = config.Version
	httpx.SendResponse(w, ld, response)
}
