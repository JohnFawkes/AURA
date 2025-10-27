package api

import (
	"aura/internal/logging"
)

func SendAppStartNotification() logging.StandardError {
	startMessage := "aura has started successfully!"
	imageURL := ""
	title := "Notification | aura"

	if !Global_Config.Notifications.Enabled {
		logging.LOG.Debug("Notifications are disabled, not sending app start notification")
		return logging.StandardError{}
	}

	logging.LOG.Debug("Sending app start notification to all providers")

	for _, provider := range Global_Config.Notifications.Providers {
		if provider.Enabled {
			switch provider.Provider {
			case "Discord":
				Err := Notification_SendDiscordMessage(provider.Discord, startMessage, imageURL, title)
				if Err.Message != "" {
					logging.LOG.ErrorWithLog(Err)
				}
			case "Pushover":
				Err := Notification_SendPushoverMessage(provider.Pushover, startMessage, imageURL, title)
				if Err.Message != "" {
					logging.LOG.ErrorWithLog(Err)
				}
			case "Gotify":
				Err := Notification_SendGotifyMessage(provider.Gotify, startMessage, imageURL, title)
				if Err.Message != "" {
					logging.LOG.ErrorWithLog(Err)
				}
			}
		}
	}

	return logging.StandardError{}
}
