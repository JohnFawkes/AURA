package notification

import (
	"aura/config"
	"aura/logging"
	"aura/utils"
	"context"
)

func SendAppStartNotification(app_port int, app_name string, app_version string) {
	ctx, ld := logging.CreateLoggingContext(context.Background(), "Notification - Send App Start Message")
	logAction := ld.AddAction("Sending App Start Notification", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)
	defer ld.Log()
	defer logAction.Complete()

	// If notifications are disabled, skip
	if !config.Current.Notifications.Enabled {
		logging.LOGGER.Debug().Timestamp().Msg("Notifications are disabled, skipping app start notification")
		return
	}

	// If notification providers are not configured, skip
	if len(config.Current.Notifications.Providers) == 0 {
		logging.LOGGER.Debug().Timestamp().Msg("No notification providers configured, skipping app start notification")
		return
	}

	// If app startup notification is disabled, skip
	if !config.Current.Notifications.NotificationTemplate.AppStartup.Enabled {
		logging.LOGGER.Debug().Timestamp().Msg("App startup notification is disabled, skipping app start notification")
		return
	}

	vars := utils.TemplateVars_AppStartup(app_name, app_version, app_port)
	title := utils.RenderTemplate(config.Current.Notifications.NotificationTemplate.AppStartup.Title, vars)
	startMessage := utils.RenderTemplate(config.Current.Notifications.NotificationTemplate.AppStartup.Message, vars)
	imageURL := ""

	for _, provider := range config.Current.Notifications.Providers {
		if provider.Enabled {
			switch provider.Provider {
			case "Discord":
				SendDiscordMessage(ctx, provider.Discord, startMessage, imageURL, title)
			case "Pushover":
				SendPushoverMessage(ctx, provider.Pushover, startMessage, imageURL, title)
			case "Gotify":
				SendGotifyMessage(ctx, provider.Gotify, startMessage, imageURL, title)
			case "Webhook":
				SendWebhookMessage(ctx, provider.Webhook, startMessage, imageURL, title)
			}
		}
	}

}
