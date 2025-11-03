package api

import (
	"aura/internal/logging"
	"context"
)

func SendAppStartNotification() {
	ctx, ld := logging.CreateLoggingContext(context.Background(), "Notification - Send App Start Message")
	logAction := ld.AddAction("Sending App Start Notification", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)
	defer ld.Log()
	defer logAction.Complete()

	if !Global_Config.Notifications.Enabled {
		logging.LOGGER.Debug().Timestamp().Msg("Notifications are disabled, skipping app start notification")
		return
	}

	if len(Global_Config.Notifications.Providers) == 0 {
		logging.LOGGER.Warn().Timestamp().Msg("No notification providers configured, skipping app start notification")
		return
	}

	startMessage := "aura has started successfully!"
	imageURL := ""
	title := "Notification | aura"

	for _, provider := range Global_Config.Notifications.Providers {
		if provider.Enabled {
			switch provider.Provider {
			case "Discord":
				Notification_SendDiscordMessage(ctx, provider.Discord, startMessage, imageURL, title)
			case "Pushover":
				Notification_SendPushoverMessage(ctx, provider.Pushover, startMessage, imageURL, title)
			case "Gotify":
				Notification_SendGotifyMessage(ctx, provider.Gotify, startMessage, imageURL, title)
			case "Webhook":
				Notification_SendWebhookMessage(ctx, provider.Webhook, startMessage, imageURL, title)
			}
		}
	}

}
