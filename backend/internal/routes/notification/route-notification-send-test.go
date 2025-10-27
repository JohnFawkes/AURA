package routes_notification

import (
	"aura/internal/api"
	"aura/internal/logging"
	routes_config "aura/internal/routes/config"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

func SendTest(w http.ResponseWriter, r *http.Request) {
	logging.LOG.Trace(r.URL.Path)
	startTime := time.Now()
	Err := logging.NewStandardError()

	// Get the Notification Provider information from the request
	var nProvider api.Config_Notification_Provider
	if err := json.NewDecoder(r.Body).Decode(&nProvider); err != nil {
		Err.Message = "Failed to decode request body"
		Err.HelpText = "Ensure the request body is valid JSON"
		Err.Details = map[string]any{
			"error": err.Error(),
		}
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		return
	}

	// If the provider is not enabled, return early
	if !nProvider.Enabled {
		return
	}

	startMessage := "This is a test notification from aura"
	imageURL := ""
	title := "Notification | aura"

	switch nProvider.Provider {
	case "Discord":
		// Validate Discord provider
		if nProvider.Discord == nil || nProvider.Discord.Webhook == "" {
			Err.Message = "Discord provider is missing or has no webhook URL"
			Err.HelpText = "Ensure the Discord provider is properly configured"
			api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
			return
		}
		webhook := nProvider.Discord.Webhook
		if routes_config.IsMaskedWebhook(webhook) {
			unmasked := getUnmaskedDiscordWebhook(webhook)
			if unmasked == "" {
				Err.Message = "Discord provider has masked webhook but no existing configuration to unmask from"
				Err.HelpText = "Try re-entering the full webhook URL"
				api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
				return
			}
			nProvider.Discord.Webhook = unmasked
		}
		Err = api.Notification_SendDiscordMessage(nProvider.Discord, startMessage, imageURL, title)
		if Err.Message != "" {
			api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
			return
		}
	case "Pushover":
		// Validate Pushover provider
		if nProvider.Pushover == nil || nProvider.Pushover.UserKey == "" || nProvider.Pushover.Token == "" {
			Err.Message = "Pushover provider is missing or has no user key/app token"
			Err.HelpText = "Ensure the Pushover provider is properly configured"
			api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
			return
		}
		userKey := nProvider.Pushover.UserKey
		token := nProvider.Pushover.Token
		if strings.HasPrefix(userKey, "***") {
			userKey = getUnmaskedPushoverField("UserKey", userKey)
		}
		if strings.HasPrefix(token, "***") {
			token = getUnmaskedPushoverField("Token", token)
		}
		if userKey == "" || token == "" {
			Err.Message = "Pushover provider has masked fields but no existing configuration to unmask from"
			Err.HelpText = "Try re-entering the full User Key and Token"
			api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
			return
		}
		nProvider.Pushover.UserKey = userKey
		nProvider.Pushover.Token = token
		Err := api.Notification_SendPushoverMessage(nProvider.Pushover, startMessage, imageURL, title)
		if Err.Message != "" {
			api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
			return
		}
	case "Gotify":
		// Validate Gotify provider
		if nProvider.Gotify == nil || nProvider.Gotify.URL == "" || nProvider.Gotify.Token == "" {
			Err.Message = "Gotify provider is missing or has no server URL/app token"
			Err.HelpText = "Ensure the Gotify provider is properly configured"
			api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
			return
		}
		url := nProvider.Gotify.URL
		token := nProvider.Gotify.Token
		if strings.HasPrefix(url, "***") {
			url = getUnmaskedGotifyField("URL", url)
		}
		if strings.HasPrefix(token, "***") {
			token = getUnmaskedGotifyField("Token", token)
		}
		if url == "" || token == "" {
			Err.Message = "Gotify provider has masked fields but no existing configuration to unmask from"
			Err.HelpText = "Try re-entering the full Server URL and Token"
			api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
			return
		}
		nProvider.Gotify.URL = url
		nProvider.Gotify.Token = token
		Err = api.Notification_SendGotifyMessage(nProvider.Gotify, startMessage, imageURL, title)
		if Err.Message != "" {
			api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
			return
		}
	default:
		Err.Message = "Unknown notification provider"
		Err.HelpText = "Ensure the notification provider is one of: Discord, Pushover, Gotify"
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		return
	}

	// Respond with a success message
	api.Util_Response_SendJson(w, http.StatusOK, api.JSONResponse{
		Status:  "success",
		Elapsed: api.Util_ElapsedTime(startTime),
		Data:    "success",
	})
}

func getUnmaskedPushoverField(field, currentValue string) string {
	for _, existingProvider := range api.Global_Config.Notifications.Providers {
		if existingProvider.Provider == "Pushover" && existingProvider.Pushover != nil {
			switch field {
			case "UserKey":
				if existingProvider.Pushover.UserKey != "" {
					// Make sure that the last few characters match the masked value
					if len(currentValue) > 3 && len(existingProvider.Pushover.UserKey) >= 3 {
						if currentValue[len(currentValue)-3:] == existingProvider.Pushover.UserKey[len(existingProvider.Pushover.UserKey)-3:] {
							return existingProvider.Pushover.UserKey
						}
					}
				}
			case "Token":
				if existingProvider.Pushover.Token != "" {
					// Make sure that the last few characters match the masked value
					if len(currentValue) > 3 && len(existingProvider.Pushover.Token) >= 3 {
						if currentValue[len(currentValue)-3:] == existingProvider.Pushover.Token[len(existingProvider.Pushover.Token)-3:] {
							return existingProvider.Pushover.Token
						}
					}
				}
			}
		}
	}
	return ""
}

func getUnmaskedDiscordWebhook(currentValue string) string {
	for _, existingProvider := range api.Global_Config.Notifications.Providers {
		if existingProvider.Provider == "Discord" && existingProvider.Discord != nil {
			if existingProvider.Discord.Webhook != "" {
				// Make sure that the last few characters match the masked value
				if len(currentValue) > 3 && len(existingProvider.Discord.Webhook) >= 3 {
					if currentValue[len(currentValue)-3:] == existingProvider.Discord.Webhook[len(existingProvider.Discord.Webhook)-3:] {
						return existingProvider.Discord.Webhook
					}
				}
			}
		}
	}
	return ""
}

func getUnmaskedGotifyField(field, currentValue string) string {
	for _, existingProvider := range api.Global_Config.Notifications.Providers {
		if existingProvider.Provider == "Gotify" && existingProvider.Gotify != nil {
			switch field {
			case "URL":
				if existingProvider.Gotify.URL != "" {
					// Make sure that the last few characters match the masked value
					if len(currentValue) > 3 && len(existingProvider.Gotify.URL) >= 3 {
						if currentValue[len(currentValue)-3:] == existingProvider.Gotify.URL[len(existingProvider.Gotify.URL)-3:] {
							return existingProvider.Gotify.URL
						}
					}
				}
			case "Token":
				if existingProvider.Gotify.Token != "" {
					// Make sure that the last few characters match the masked value
					if len(currentValue) > 3 && len(existingProvider.Gotify.Token) >= 3 {
						if currentValue[len(currentValue)-3:] == existingProvider.Gotify.Token[len(existingProvider.Gotify.Token)-3:] {
							return existingProvider.Gotify.Token
						}
					}
				}
			}
		}
	}
	return ""
}
