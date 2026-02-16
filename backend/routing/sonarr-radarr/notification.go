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

func sendFileDownloadNotification(itemInfo string, setID string, image models.ImageFile, isUpgrade bool, result string) {
	if !config.Current.Notifications.Enabled {
		return
	} else if len(config.Current.Notifications.Providers) == 0 {
		return
	}

	title := "Sonarr - Upgrade"
	if !isUpgrade {
		title = "Sonarr - New Download"
	}
	message := fmt.Sprintf(
		"%s\n%s\nSet ID: %s",
		itemInfo,
		utils.GetFileDownloadName(itemInfo, image),
		setID,
	)
	if result != "Success" {
		message = fmt.Sprintf("%s\n\n%s", message, result)
	}
	imageURL := fmt.Sprintf("%s/%s?v=%s&key=jpg",
		"https://images.mediux.io/assets",
		image.ID,
		image.Modified.Format("20060102150405"),
	)

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
