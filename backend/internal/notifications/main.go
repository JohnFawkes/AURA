package notifications

import (
	"aura/internal/config"
	"aura/internal/logging"
)

func SendAppStartNotification() logging.StandardError {
	startMessage := "aura has started successfully!"
	imageURL := ""
	title := "Notification | aura"

	if !config.Global.Notifications.Enabled {
		logging.LOG.Debug("Notifications are disabled, not sending app start notification")
		return logging.StandardError{}
	}

	logging.LOG.Debug("Sending app start notification to all providers")

	for _, provider := range config.Global.Notifications.Providers {
		if provider.Enabled {
			switch provider.Provider {
			case "Discord":
				Err := SendDiscordNotification(provider.Discord, startMessage, imageURL, title)
				if Err.Message != "" {
					logging.LOG.ErrorWithLog(Err)
				}
			case "Pushover":
				Err := SendPushoverNotification(provider.Pushover, startMessage, imageURL, title)
				if Err.Message != "" {
					logging.LOG.ErrorWithLog(Err)
				}
			}
		}
	}

	return logging.StandardError{}
}
