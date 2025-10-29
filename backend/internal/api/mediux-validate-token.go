package api

import (
	"aura/internal/logging"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
)

func Mediux_ValidateToken(ctx context.Context, token string) logging.LogErrorInfo {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Validating Mediux Token", logging.LevelDebug)
	defer logAction.Complete()

	// If the token is empty, return an error
	if token == "" {
		logAction.SetError("MediUX token is empty", "Please provide a valid token", nil)
		return *logAction.Error
	}

	// Build the API URL
	u, err := url.Parse(MediuxBaseURL)
	if err != nil {
		logAction.SetError("Failed to parse Mediux base URL", err.Error(), nil)
		return *logAction.Error
	}
	u.Path = path.Join(u.Path, "users", "me")
	URL := u.String()

	// Add the authorization header
	headers := make(map[string]string)
	headers["Authorization"] = "Bearer " + token

	httpResp, respBody, logErr := MakeHTTPRequest(ctx, URL, http.MethodGet, headers, 30, nil, "")
	if logErr.Message != "" {
		return logErr
	}
	defer httpResp.Body.Close()

	// Check the response status code
	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		logAction.SetError("Mediux server returned non-200 status", fmt.Sprintf("Status Code: %d", httpResp.StatusCode), map[string]any{
			"url":         URL,
			"status_code": httpResp.StatusCode,
			"response":    string(respBody),
		})
		return *logAction.Error
	}

	return logging.LogErrorInfo{}
}
