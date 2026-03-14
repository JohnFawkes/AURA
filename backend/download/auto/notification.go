package autodownload

import (
	"aura/config"
	"aura/logging"
	"aura/models"
	"aura/notification"
	"aura/utils"
	"context"
	"fmt"
)

func sendFileDownloadNotification(mediaItem models.MediaItem, set models.DBPosterSetDetail, imageWithReason ImageFileWithReason) {
	if !config.Current.Notifications.Enabled {
		return
	} else if len(config.Current.Notifications.Providers) == 0 {
		return
	}

	vars := utils.TemplateVars_Autodownload(mediaItem, set, imageWithReason.ImageFile, imageWithReason.ReasonTitle, imageWithReason.Reason)
	title := utils.RenderTemplate(config.Current.Notifications.NotificationTemplate.Autodownload.Title, vars)
	message := utils.RenderTemplate(config.Current.Notifications.NotificationTemplate.Autodownload.Message, vars)
	imageURL := ""
	if config.Current.Notifications.NotificationTemplate.Autodownload.IncludeImage {
		imageURL = fmt.Sprintf("%s/%s?v=%s&key=jpg",
			"https://images.mediux.io/assets",
			imageWithReason.ID,
			imageWithReason.Modified.Format("20060102150405"),
		)
	}

	ctx, ld := logging.CreateLoggingContext(context.Background(), "Notification - Send File Download Message")
	logAction := ld.AddAction("Sending File Download Notification", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)
	defer ld.Log()
	defer logAction.Complete()

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
}
