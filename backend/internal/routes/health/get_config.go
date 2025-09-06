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
	safeConfigData.Mediux.Token = config.MaskToken(safeConfigData.Mediux.Token)
	safeConfigData.TMDB.ApiKey = config.MaskToken(safeConfigData.TMDB.ApiKey)
	safeConfigData.MediaServer.Token = config.MaskToken(safeConfigData.MediaServer.Token)

	// Mask the Notification Webhook URL
	for _, provider := range safeConfigData.Notifications.Providers {
		switch provider.Provider {
		case "Discord":
			provider.Discord.Webhook = config.MaskWebhookURL(provider.Discord.Webhook)
		case "Pushover":
			provider.Pushover.Token = config.MaskToken(provider.Pushover.Token)
			provider.Pushover.UserKey = config.MaskToken(provider.Pushover.UserKey)
		}
	}

	safeConfigData.Logging.File = logging.GetTodayLogFile()

	utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
		Status:  "success",
		Elapsed: utils.ElapsedTime(startTime),
		Data:    safeConfigData,
	})
}
