package health

import (
	"aura/internal/config"
	"aura/internal/logging"
	"aura/internal/utils"
	"net/http"
	"time"
)

func GetConfig(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	logging.LOG.Trace(r.URL.Path)

	safeConfigData := *config.Global

	// Remove by masking sensitive information with "***"
	// Keep the last 4 characters of the data
	safeConfigData.Mediux.Token = maskToken(safeConfigData.Mediux.Token)
	safeConfigData.TMDB.ApiKey = maskToken(safeConfigData.TMDB.ApiKey)
	safeConfigData.MediaServer.Token = maskToken(safeConfigData.MediaServer.Token)

	// Mask the Notification Webhook URL
	safeConfigData.Notification.Webhook = config.MaskWebhookURL(safeConfigData.Notification.Webhook)

	safeConfigData.Logging.File = logging.GetTodayLogFile()

	utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
		Status:  "success",
		Elapsed: utils.ElapsedTime(startTime),
		Data:    safeConfigData,
	})
}

func maskToken(token string) string {
	if token == "" {
		return "N/A"
	}
	if len(token) < 4 {
		return "***" + token[len(token)-1:]
	}
	return "***" + token[len(token)-4:]
}
