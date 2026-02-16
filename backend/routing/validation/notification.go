package routes_validation

import (
	"aura/config"
	"aura/logging"
	"aura/notification"
	"aura/utils"
	"aura/utils/httpx"
	"fmt"
	"net/http"
)

func SendTestNotification(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Send Test Notification", logging.LevelDebug)
	ctx = logging.WithCurrentAction(ctx, logAction)

	// Get the Notification Provider from the request body
	var nProvider config.Config_Notification_Provider
	Err := httpx.DecodeRequestBodyToJSON(ctx, r.Body, &nProvider, "Notification Provider")
	if Err.Message != "" {
		httpx.SendResponse(w, ld, nil)
		return
	}

	// If the provider is not enabled, return early
	if !nProvider.Enabled {
		logAction.SetError("Notification provider is not enabled", fmt.Sprintf("%s is not enabled, cannot send test notification", nProvider.Provider), nil)
		httpx.SendResponse(w, ld, nil)
		return
	}

	// Validate the provider settings
	validProvider := config.ValidateNotificationsProvider(ctx, &nProvider)
	if !validProvider {
		httpx.SendResponse(w, ld, nil)
		return
	}

	vars := utils.TemplateVars_TestNotification()
	title := utils.RenderTemplate(config.Current.Notifications.NotificationTemplate.TestNotification.Title, vars)
	message := utils.RenderTemplate(config.Current.Notifications.NotificationTemplate.TestNotification.Message, vars)
	imageURL := ""

	// If the Template is disabled, don't send a message
	if !config.Current.Notifications.NotificationTemplate.TestNotification.Enabled {
		httpx.SendResponse(w, ld, "ok")
		return
	}

	switch nProvider.Provider {
	case "Discord":
		webhook := nProvider.Discord.Webhook
		if config.IsMaskedWebhook(webhook) {
			unmasked := getUnmaskedDiscordWebhook(webhook)
			if unmasked == "" {
				logAction.SetError("Unable to unmask Discord webhook", "Please provide the full Discord webhook URL", nil)
				httpx.SendResponse(w, ld, nil)
				return
			}
			nProvider.Discord.Webhook = unmasked
		}
		Err := notification.SendDiscordMessage(ctx, nProvider.Discord, message, imageURL, title)
		if Err.Message != "" {
			httpx.SendResponse(w, ld, nil)
			return
		}
	case "Pushover":
		userKey := nProvider.Pushover.UserKey
		apiToken := nProvider.Pushover.ApiToken
		if config.IsMaskedField(userKey) {
			userKey = getUnmaskedPushoverField("UserKey", userKey)
		}
		if config.IsMaskedField(apiToken) {
			apiToken = getUnmaskedPushoverField("Token", apiToken)
		}
		if userKey == "" || apiToken == "" {
			logAction.SetError("Unable to unmask Pushover credentials", "Please provide the full Pushover UserKey and ApiToken", nil)
			httpx.SendResponse(w, ld, nil)
			return
		}
		nProvider.Pushover.UserKey = userKey
		nProvider.Pushover.ApiToken = apiToken
		Err := notification.SendPushoverMessage(ctx, nProvider.Pushover, message, imageURL, title)
		if Err.Message != "" {
			httpx.SendResponse(w, ld, nil)
			return
		}
	case "Gotify":
		url := nProvider.Gotify.URL
		apiToken := nProvider.Gotify.ApiToken
		if config.IsMaskedField(url) {
			url = getUnmaskedGotifyField("URL", url)
		}
		if config.IsMaskedField(apiToken) {
			apiToken = getUnmaskedGotifyField("Token", apiToken)
		}
		if url == "" || apiToken == "" {
			logAction.SetError("Unable to unmask Gotify credentials", "Please provide the full Gotify URL and ApiToken", nil)
			httpx.SendResponse(w, ld, nil)
			return
		}
		nProvider.Gotify.URL = url
		nProvider.Gotify.ApiToken = apiToken
		Err := notification.SendGotifyMessage(ctx, nProvider.Gotify, message, imageURL, title)
		if Err.Message != "" {
			httpx.SendResponse(w, ld, nil)
			return
		}
	case "Webhook":
		Err := notification.SendWebhookMessage(ctx, nProvider.Webhook, message, imageURL, title)
		if Err.Message != "" {
			httpx.SendResponse(w, ld, nil)
			return
		}
	default:
		logAction.SetError("Unsupported notification provider", fmt.Sprintf("The notification provider '%s' is not supported for test messages", nProvider.Provider), nil)
		httpx.SendResponse(w, ld, nil)
		return
	}

	httpx.SendResponse(w, ld, "ok")
}

func getUnmaskedPushoverField(field, currentValue string) string {
	for _, existingProvider := range config.Current.Notifications.Providers {
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
				if existingProvider.Pushover.ApiToken != "" {
					// Make sure that the last few characters match the masked value
					if len(currentValue) > 3 && len(existingProvider.Pushover.ApiToken) >= 3 {
						if currentValue[len(currentValue)-3:] == existingProvider.Pushover.ApiToken[len(existingProvider.Pushover.ApiToken)-3:] {
							return existingProvider.Pushover.ApiToken
						}
					}
				}
			}
		}
	}
	return ""
}

func getUnmaskedDiscordWebhook(currentValue string) string {
	for _, existingProvider := range config.Current.Notifications.Providers {
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
	for _, existingProvider := range config.Current.Notifications.Providers {
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
				if existingProvider.Gotify.ApiToken != "" {
					// Make sure that the last few characters match the masked value
					if len(currentValue) > 3 && len(existingProvider.Gotify.ApiToken) >= 3 {
						if currentValue[len(currentValue)-3:] == existingProvider.Gotify.ApiToken[len(existingProvider.Gotify.ApiToken)-3:] {
							return existingProvider.Gotify.ApiToken
						}
					}
				}
			}
		}
	}
	return ""
}
