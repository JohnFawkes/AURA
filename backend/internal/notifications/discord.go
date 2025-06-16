package notifications

import (
	"aura/internal/config"
	"aura/internal/logging"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func SendDiscordNotification(message string, imageURL string, title string) logging.StandardError {
	Err := logging.NewStandardError()

	if !validNotificationProvider() || config.Global.Notification.Provider != "Discord" {

		Err.Message = fmt.Sprintf("Invalid notification provider: %s", config.Global.Notification.Provider)
		Err.HelpText = "Ensure the notification provider is set to 'Discord' in the configuration."
		return Err
	}

	webhookURL := config.Global.Notification.Webhook

	if webhookURL == "" {

		Err.Message = "Discord webhook URL is not configured"
		Err.HelpText = "Please set the Discord webhook URL in the configuration."
		return Err
	}

	embed := map[string]any{
		"author": map[string]any{
			"name":     "MediUX AURA Bot",
			"url":      "https://github.com/mediux-team/aura",
			"icon_url": "https://raw.githubusercontent.com/mediux-team/aura/master/frontend/public/aura_logo.png",
		},
		"title":       title,
		"description": message,
		"color":       0x9B59B6, // purple color
	}
	if imageURL != "" {
		embed["image"] = map[string]any{
			"url": imageURL,
		}
	}

	webhookBody := map[string]any{
		"username":   "MediUX AURA Bot",
		"avatar_url": "https://raw.githubusercontent.com/mediux-team/aura/master/frontend/public/aura_logo.png",
		"embeds":     []map[string]any{embed},
	}

	bodyBytes, err := json.Marshal(webhookBody)
	if err != nil {

		Err.Message = "Failed to marshal webhook body"
		Err.HelpText = "Ensure the webhook body is correctly formatted."
		Err.Details = fmt.Sprintf("Error: %s", err.Error())
		return Err
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {

		Err.Message = "Failed to send webhook request"
		Err.HelpText = "Ensure the Discord webhook URL is correct and accessible."
		Err.Details = fmt.Sprintf("Error: %s", err.Error())
		return Err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {

		Err.Message = fmt.Sprintf("Failed to send Discord notification, received status code: %d", resp.StatusCode)
		Err.HelpText = "Ensure the Discord webhook URL is correct and the bot has permission to send messages."
		Err.Details = fmt.Sprintf("Response status: %s", resp.Status)
		return Err
	}

	return logging.StandardError{}
}

func SendDiscordAppStartNotification() logging.StandardError {
	if !validNotificationProvider() || config.Global.Notification.Provider != "Discord" {
		Err := logging.NewStandardError()

		Err.Message = fmt.Sprintf("Invalid notification provider: %s", config.Global.Notification.Provider)
		Err.HelpText = "Ensure the notification provider is set to 'Discord' in the configuration."
		return Err
	}

	message := "MediUX AURA has started successfully!"
	return SendDiscordNotification(message, "", "MediUX AURA Notification")
}
