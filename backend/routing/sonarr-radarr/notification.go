package routes_sonarr_radarr

import (
	"aura/config"
	"aura/logging"
	"aura/models"
	"aura/notification"
	"aura/utils"
	"context"
	"fmt"
)

func sendFileDownloadNotification(mediaItem models.MediaItem, set models.DBPosterSetDetail, image models.ImageFile, isUpgrade bool, result string) {
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

	// If Sonarr/Radarr notification is disabled, skip
	if !config.Current.Notifications.NotificationTemplate.SonarrNotification.Enabled {
		logging.LOGGER.Debug().Timestamp().Msg("Sonarr/Radarr notification is disabled, skipping app start notification")
		return
	}

	reasonTitle := "New Download"
	reason := "A new episode was downloaded via Sonarr for this media item."
	if isUpgrade {
		reasonTitle = "Upgrade"
		reason = "An existing episode was upgraded via Sonarr for this media item."
	}

	vars := utils.TemplateVars_SonarrNotification(mediaItem, set, image, reasonTitle, reason, result)
	title := utils.RenderTemplate(config.Current.Notifications.NotificationTemplate.SonarrNotification.Title, vars)
	message := utils.RenderTemplate(config.Current.Notifications.NotificationTemplate.SonarrNotification.Message, vars)
	imageURL := ""
	if config.Current.Notifications.NotificationTemplate.SonarrNotification.IncludeImage {
		imageURL = fmt.Sprintf("%s/%s?v=%s&key=jpg",
			"https://images.mediux.io/assets",
			image.ID,
			image.Modified.Format("20060102150405"),
		)
	}

	ctx, ld := logging.CreateLoggingContext(context.Background(), "Notification - Send File Download Message")
	logAction := ld.AddAction("Sending File Download Notification", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	// Send a notification to all configured providers
	for _, provider := range config.Current.Notifications.Providers {
		if provider.Enabled {
			switch provider.Provider {
			case "Discord":
				notification.SendDiscordMessage(
					ctx,
					provider.Discord,
					message,
					imageURL,
					title,
				)
			case "Pushover":
				notification.SendPushoverMessage(
					ctx,
					provider.Pushover,
					message,
					imageURL,
					title,
				)
			case "Gotify":
				notification.SendGotifyMessage(
					ctx,
					provider.Gotify,
					message,
					imageURL,
					title,
				)
			case "Webhook":
				notification.SendWebhookMessage(
					ctx,
					provider.Webhook,
					message,
					imageURL,
					title,
				)
			}
		}
	}

	ld.Log()
	logAction.Complete()
}
