package routes_api

import (
	"aura/internal/api"
	"aura/internal/logging"
	"net/http"
)

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	_, ld := logging.CreateLoggingContext(r.Context(), "Health Check")

	ld.Status = logging.StatusSuccess
	api.Util_Response_SendJSON(w, ld, "Server is online")
}
