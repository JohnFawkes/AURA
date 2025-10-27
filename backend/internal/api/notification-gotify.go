package api

import (
	"aura/internal/logging"
	"aura/internal/masking"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func Notification_SendGotifyMessage(provider *Config_Notification_Gotify, message string, imageURL string, title string) logging.StandardError {
	logging.LOG.Trace("Sending Gotify Notification")
	Err := logging.NewStandardError()

	if provider.URL == "" || provider.Token == "" {
		Err.Message = "Gotify provider is not properly configured"
		logging.LOG.ErrorWithLog(Err)
		return Err
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
		Err.Message = "Failed to send Gotify notification: " + httpErr.Error()
		Err.HelpText = "Ensure the Gotify URL and token are correct and the server is reachable."
		Err.Details = map[string]any{
			"gotifyURL":   provider.URL,
			"gotifyToken": masking.Masking_Token(provider.Token),
			"error":       httpErr.Error(),
		}
		logging.LOG.ErrorWithLog(Err)
		return Err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		body, _ := io.ReadAll(resp.Body)
		Err.Message = "Failed to send Gotify notification"
		Err.HelpText = "Ensure the Gotify URL and token are correct and the server is reachable."
		Err.Details = map[string]any{
			"gotifyURL":    provider.URL,
			"gotifyToken":  masking.Masking_Token(provider.Token),
			"statusCode":   resp.StatusCode,
			"responseBody": strings.TrimSpace(string(body)),
		}
		logging.LOG.ErrorWithLog(Err)
		return Err
	}

	return Err
}
