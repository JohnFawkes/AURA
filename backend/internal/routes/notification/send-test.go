package routes_notification

import (
	"aura/internal/api"
	"aura/internal/logging"
	"aura/internal/masking"
	routes_config "aura/internal/routes/config"
	"fmt"
	"net/http"
	"strings"
)

func SendTest(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Send Test Notification", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	// Get the Notification Provider information from the request
	var nProvider api.Config_Notification_Provider
	Err := api.DecodeRequestBodyJSON(ctx, r.Body, &nProvider, "Config_Notification_Provider")
	if Err.Message != "" {
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}

	// If the provider is not enabled, return early
	if !nProvider.Enabled {
		logAction.SetError("Notification Provider Disabled",
			"The specified notification provider is not enabled, cannot send test notification", map[string]any{
				"provider": nProvider.Provider,
			})
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}

	startMessage := "This is a test notification from aura"
	imageURL := ""
	title := "Notification | aura"

	switch nProvider.Provider {
	case "Discord":
		// Validate Discord provider
		if nProvider.Discord == nil || nProvider.Discord.Webhook == "" {
			logAction.SetError("Invalid Discord Provider",
				"Ensure the Discord provider is properly configured",
				map[string]any{
					"provider": nProvider.Provider,
					"webhook":  masking.Masking_WebhookURL(nProvider.Discord.Webhook),
				})
			api.Util_Response_SendJSON(w, ld, nil)
			return
		}
		webhook := nProvider.Discord.Webhook
		if routes_config.IsMaskedWebhook(webhook) {
			unmasked := getUnmaskedDiscordWebhook(webhook)
			if unmasked == "" {
				logAction.SetError("Invalid Discord Provider",
					"Discord provider has masked webhook but no existing configuration to unmask from",
					map[string]any{
						"provider": nProvider.Provider,
						"webhook":  masking.Masking_WebhookURL(webhook),
					})
				api.Util_Response_SendJSON(w, ld, nil)
				return
			}
			nProvider.Discord.Webhook = unmasked
		}
		Err := api.Notification_SendDiscordMessage(ctx, nProvider.Discord, startMessage, imageURL, title)
		if Err.Message != "" {
			api.Util_Response_SendJSON(w, ld, nil)
			return
		}

	case "Pushover":
		// Validate Pushover provider
		if nProvider.Pushover == nil || nProvider.Pushover.UserKey == "" || nProvider.Pushover.Token == "" {
			logAction.SetError("Invalid Pushover Provider",
				"Ensure the Pushover provider is properly configured",
				map[string]any{
					"provider": nProvider.Provider,
					"userKey":  masking.Masking_Token(nProvider.Pushover.UserKey),
					"token":    masking.Masking_Token(nProvider.Pushover.Token),
				})
			api.Util_Response_SendJSON(w, ld, nil)
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
			logAction.SetError("Invalid Pushover Provider",
				"The Pushover notification provider appears to have masked fields but no existing values could be found",
				map[string]any{
					"provider": nProvider.Provider,
					"userKey":  masking.Masking_Token(nProvider.Pushover.UserKey),
					"token":    masking.Masking_Token(nProvider.Pushover.Token),
				})
			api.Util_Response_SendJSON(w, ld, nil)
			return
		}
		nProvider.Pushover.UserKey = userKey
		nProvider.Pushover.Token = token
		Err := api.Notification_SendPushoverMessage(ctx, nProvider.Pushover, startMessage, imageURL, title)
		if Err.Message != "" {
			api.Util_Response_SendJSON(w, ld, nil)
			return
		}

	case "Gotify":
		// Validate Gotify provider
		if nProvider.Gotify == nil || nProvider.Gotify.URL == "" || nProvider.Gotify.Token == "" {
			logAction.SetError("Invalid Gotify Provider",
				"The Gotify notification provider is missing required fields",
				map[string]any{
					"provider": nProvider.Provider,
					"url":      nProvider.Gotify.URL,
					"token":    masking.Masking_Token(nProvider.Gotify.Token),
				})
			api.Util_Response_SendJSON(w, ld, nil)
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
			logAction.SetError("Invalid Gotify Provider",
				"The Gotify notification provider appears to have masked fields but no existing values could be found",
				map[string]any{
					"provider": nProvider.Provider,
					"url":      nProvider.Gotify.URL,
					"token":    masking.Masking_Token(nProvider.Gotify.Token),
				})
			api.Util_Response_SendJSON(w, ld, nil)
			return
		}
		nProvider.Gotify.URL = url
		nProvider.Gotify.Token = token
		Err := api.Notification_SendGotifyMessage(ctx, nProvider.Gotify, startMessage, imageURL, title)
		if Err.Message != "" {
			api.Util_Response_SendJSON(w, ld, nil)
			return
		}

	case "Webhook":
		// Validate Webhook provider
		if nProvider.Webhook == nil || nProvider.Webhook.URL == "" {
			logAction.SetError("Invalid Webhook Provider",
				"Ensure the Webhook provider is properly configured",
				map[string]any{
					"provider": nProvider.Provider,
					"url":      masking.Masking_WebhookURL(nProvider.Webhook.URL),
				})
			api.Util_Response_SendJSON(w, ld, nil)
			return
		}
		Err := api.Notification_SendWebhookMessage(ctx, nProvider.Webhook, startMessage, imageURL, title)
		if Err.Message != "" {
			api.Util_Response_SendJSON(w, ld, nil)
			return
		}

	default:
		logAction.SetError("Unsupported Notification Provider",
			"The specified notification provider is not supported for test notifications",
			map[string]any{
				"provider": nProvider.Provider,
			})
		api.Util_Response_SendJSON(w, ld, nil)
		return

	}

	api.Util_Response_SendJSON(w, ld, fmt.Sprintf("Test notification sent via %s", nProvider.Provider))
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
