package api

import (
	"aura/internal/logging"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func Notification_SendGotifyMessage(ctx context.Context, provider *Config_Notification_Gotify, message string, imageURL string, title string) logging.LogErrorInfo {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Sending Gotify Notification", logging.LevelInfo)
	defer logAction.Complete()

	if provider.URL == "" || provider.Token == "" {
		logAction.SetError("Missing Gotify configuration", "Please configure the Gotify URL and Token", nil)
		return *logAction.Error
	}

	baseEndpoint := strings.TrimRight(provider.URL, "/")
	gotifyEndpoint := fmt.Sprintf("%s/message?token=%s", baseEndpoint, provider.Token)

	// Create form data for Gotify notification
	form := url.Values{}
	form.Set("message", message)
	form.Set("title", title)
	form.Set("priority", "5")

	// Optional extras for image
	if imageURL != "" {
		extras := map[string]any{
			"client::notification": map[string]any{
				"bigImageUrl": imageURL,
			},
		}
		if b, err := json.Marshal(extras); err == nil {
			form.Set("extras", string(b))
		}
	}

	resp, httpErr := http.PostForm(gotifyEndpoint, form)
	if httpErr != nil {
		logAction.SetError("Failed to send Gotify message", "An error occurred while sending the message to Gotify", map[string]any{
			"error": httpErr.Error(),
		})
		return *logAction.Error
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		body, _ := io.ReadAll(resp.Body)
		logAction.SetError("Failed to send Gotify message", "Received non-2xx response from Gotify", map[string]any{
			"status_code": resp.StatusCode,
			"response":    string(body),
		})
		return *logAction.Error
	}

	return logging.LogErrorInfo{}

}
