package route_health

import (
	"aura/internal/config"
	"aura/internal/logging"
	"aura/internal/notifications"
	"aura/internal/utils"
	"net/http"
	"time"
)

func SendTestNotification(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	logging.LOG.Trace(r.URL.Path)
	Err := logging.NewStandardError()

	startMessage := "This is a test notification from aura"
	imageURL := ""
	title := "Notification | aura"

	if !config.Global.Notifications.Enabled {
		return
	}

	errorSending := false
	for _, provider := range config.Global.Notifications.Providers {
		if provider.Enabled {
			switch provider.Provider {
			case "Discord":
				Err := notifications.SendDiscordNotification(provider.Discord, startMessage, imageURL, title)
				if Err.Message != "" {
					logging.LOG.Warn(Err.Message)
					errorSending = true
				}
			case "Pushover":
				Err := notifications.SendPushoverNotification(provider.Pushover, startMessage, imageURL, title)
				if Err.Message != "" {
					logging.LOG.Warn(Err.Message)
					errorSending = true
				}
			case "Gotify":
				Err := notifications.SendGotifyNotification(provider.Gotify, startMessage, imageURL, title)
				if Err.Message != "" {
					logging.LOG.Warn(Err.Message)
					errorSending = true
				}
			}
		}
	}

	// If there was an error sending notifications, respond with the error
	if errorSending {
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}

	// Respond with a success message
	utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
		Status:  "success",
		Elapsed: utils.ElapsedTime(startTime),
		Data:    "success",
	})
}
