package health

import (
	"aura/internal/config"
	"aura/internal/logging"
	"aura/internal/utils"
	"net/http"
	"strings"
	"time"
)

func GetConfig(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	logging.LOG.Trace(r.URL.Path)

	safeConfigData := *config.Global

	// Remove by masking sensitive information with "***"
	// Keep the last 4 characters of the data
	safeConfigData.Mediux.Token = "***" + safeConfigData.Mediux.Token[len(safeConfigData.Mediux.Token)-4:]
	safeConfigData.TMDB.ApiKey = "***" + safeConfigData.TMDB.ApiKey[len(safeConfigData.TMDB.ApiKey)-4:]
	safeConfigData.MediaServer.Token = "***" + safeConfigData.MediaServer.Token[len(safeConfigData.MediaServer.Token)-4:]

	// Mask the Notification Webhook URL
	safeConfigData.Notification.Webhook = maskWebhookURL(safeConfigData.Notification.Webhook)

	safeConfigData.Logging.File = logging.GetTodayLogFile()

	utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
		Status:  "success",
		Elapsed: utils.ElapsedTime(startTime),
		Data:    safeConfigData,
	})
}

func maskWebhookURL(url string) string {
	if url == "" {
		return ""
	}

	// Split the URL by "/"
	parts := strings.Split(url, "/")
	if len(parts) < 6 {
		return url
	}

	// Get the webhook ID and token
	webhookID := parts[len(parts)-2]
	webhookToken := parts[len(parts)-1]

	// Mask the webhook ID - keep last 3 digits
	maskedID := "****" + webhookID[len(webhookID)-3:]

	// Mask the token - keep last 3 characters
	maskedToken := "***" + webhookToken[len(webhookToken)-3:]

	// Reconstruct the URL
	parts[len(parts)-2] = maskedID
	parts[len(parts)-1] = maskedToken

	return strings.Join(parts, "/")
}
