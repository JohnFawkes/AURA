package api

import (
	"aura/internal/logging"
	"aura/internal/masking"
	"fmt"
	"net/http"

	"github.com/gregdel/pushover"
)

func Notification_SendPushoverMessage(provider *Config_Notification_Pushover, message string, imageURL, title string) logging.StandardError {
	logging.LOG.Trace("Sending Pushover Notification")
	Err := logging.NewStandardError()

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
			Err.Message = fmt.Sprintf("Failed to download image for Pushover notification: %v", err)
			Err.HelpText = "Ensure the image URL is correct and accessible."
			Err.Details = map[string]any{
				"imageURL": imageURL,
				"error":    err.Error(),
			}
			return Err
		}
		defer resp.Body.Close()
		msg.AddAttachment(resp.Body)
	}

	// Send the notification
	_, err := app.SendMessage(msg, recipient)
	if err != nil {
		Err.Message = fmt.Sprintf("Failed to send Pushover notification: %v", err)
		Err.HelpText = "Ensure the Pushover token and user key are correct."
		Err.Details = map[string]any{
			"pushoverToken":   masking.Masking_Token(provider.Token),
			"pushoverUserKey": masking.Masking_Token(provider.UserKey),
		}
		return Err
	}

	return logging.StandardError{}
}
