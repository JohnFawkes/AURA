package api

import (
	"aura/internal/logging"
	"context"
	"net/http"

	"github.com/gregdel/pushover"
)

func Notification_SendPushoverMessage(ctx context.Context, provider *Config_Notification_Pushover, message string, imageURL string, title string) logging.LogErrorInfo {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Sending Pushover Notification", logging.LevelInfo)
	defer logAction.Complete()

	// Create a new pushover client
	app := pushover.New(provider.Token)

	// Create a new recipient
	recipient := pushover.NewRecipient(provider.UserKey)

	// Create a new message
	msg := pushover.NewMessageWithTitle(message, title)
	// If an image URL is provided, download it and add it as an attachment
	if imageURL != "" {
		resp, err := http.Get(imageURL)
		if err != nil {
			logAction.SetError("Failed to download image for Pushover message",
				"An error occurred while downloading the image",
				map[string]any{"error": err.Error()})
			return *logAction.Error
		}
		defer resp.Body.Close()
		msg.AddAttachment(resp.Body)
	}

	// Send the notification
	_, err := app.SendMessage(msg, recipient)
	if err != nil {
		logAction.SetError("Failed to send Pushover message",
			"An error occurred while sending the Pushover message",
			map[string]any{"error": err.Error()})
		return *logAction.Error
	}

	return logging.LogErrorInfo{}
}
