package api

import (
	"aura/internal/logging"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

func Notification_SendDiscordMessage(ctx context.Context, provider *Config_Notification_Discord, message string, imageURL string, title string) logging.LogErrorInfo {
	if Global_Config.Notifications.Enabled == false {
		return logging.LogErrorInfo{}
	}

	ctx, logAction := logging.AddSubActionToContext(ctx, "Sending Discord Notification", logging.LevelInfo)
	defer logAction.Complete()

	webhookURL := provider.Webhook
	if webhookURL == "" {
		logAction.SetError("Missing Webhook URL", "Please configure the Discord webhook URL", nil)
		return *logAction.Error
	}

	embed := map[string]any{
		"author": map[string]any{
			"name":     "MediUX Bot | aura",
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
		"username":   "MediUX Bot | aura",
		"avatar_url": "https://raw.githubusercontent.com/mediux-team/aura/master/frontend/public/aura_logo.png",
		"embeds":     []map[string]any{embed},
	}

	bodyBytes, err := json.Marshal(webhookBody)
	if err != nil {
		logAction.SetError("Failed to marshal webhook body",
			"An error occurred while preparing the Discord message",
			map[string]any{"error": err.Error()})
		return *logAction.Error
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		logAction.SetError("Failed to send Discord message",
			"An error occurred while sending the message to Discord",
			map[string]any{"error": err.Error()})
		return *logAction.Error
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		logAction.SetError("Failed to send Discord message",
			"Received non-204 response from Discord",
			map[string]any{
				"status_code": resp.StatusCode,
			})
		return *logAction.Error
	}

	return logging.LogErrorInfo{}
}
