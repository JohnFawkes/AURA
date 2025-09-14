package masking

import "strings"

func MaskWebhookURL(url string) string {
	if url == "" {
		return ""
	}

	// Split the URL by "/"
	parts := strings.Split(url, "/")
	if len(parts) < 6 {
		return url
	}

	// Get the webhook ID and token
	webhookID := parts[len(parts)-2]
	webhookToken := parts[len(parts)-1]

	// Mask the webhook ID - keep last 3 digits
	maskedID := "****" + webhookID[len(webhookID)-3:]

	// Mask the token - keep last 3 characters
	maskedToken := "***" + webhookToken[len(webhookToken)-3:]

	// Reconstruct the URL
	parts[len(parts)-2] = maskedID
	parts[len(parts)-1] = maskedToken

	return strings.Join(parts, "/")
}
