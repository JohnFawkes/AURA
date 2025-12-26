package api

import (
	"aura/internal/logging"
	"context"
	"encoding/json"
	"net/http"
)

func Notification_SendWebhookMessage(ctx context.Context, provider *Config_Notification_Webhook, message string, imageURL string, title string) logging.LogErrorInfo {
	if Global_Config.Notifications.Enabled == false {
		return logging.LogErrorInfo{}
	}

	ctx, logAction := logging.AddSubActionToContext(ctx, "Sending Webhook Notification", logging.LevelInfo)
	defer logAction.Complete()

	if provider.URL == "" {
		logAction.SetError("Missing Webhook configuration", "Please configure the Webhook URL", nil)
		return *logAction.Error
	}

	// Prepare payload
	payload := map[string]any{
		"title":   title,
		"message": message,
	}
	if imageURL != "" {
		payload["image_url"] = imageURL
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		logAction.SetError("Failed to marshal webhook payload", "An error occurred while preparing the webhook payload", map[string]any{
			"error": err.Error(),
		})
		return *logAction.Error
	}

	httpResp, respBody, Err := MakeHTTPRequest(ctx, provider.URL, http.MethodPost, provider.Headers, 60, payloadBytes, provider.URL)
	if Err.Message != "" {
		return Err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode < 200 || httpResp.StatusCode > 299 {
		logAction.SetError("Failed to send Webhook message", "Received non-2xx response from Webhook URL", map[string]any{
			"status_code": httpResp.StatusCode,
			"response":    string(respBody),
		})
		return *logAction.Error
	}

	return logging.LogErrorInfo{}
}
