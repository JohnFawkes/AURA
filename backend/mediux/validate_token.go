package mediux

import (
	"aura/logging"
	"context"
	"fmt"
	"net/url"
	"path"
	"time"
)

func ValidateToken(ctx context.Context, token string) (isValid bool, Err logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Validating MediUX Token", logging.LevelDebug)
	defer logAction.Complete()

	isValid = false
	Err = logging.LogErrorInfo{}

	// If the token is empty, it's invalid
	if token == "" {
		logAction.SetError("MediUX token is empty", "Provide a valid token", nil)
		return isValid, *logAction.Error
	}

	// Construct the URL for the MediUX API request
	u, err := url.Parse(MediuxApiURL)
	if err != nil {
		logAction.SetError("Failed to parse MediUX base URL", "Ensure the URL is valid", map[string]any{"error": err.Error()})
		return isValid, *logAction.Error
	}
	u.Path = path.Join(u.Path, "users", "me")
	URL := u.String()

	var lastErrorMsg string
	var lastErrorDetail map[string]any

	for attempt := 1; attempt <= 3; attempt++ {
		attemptAction := logAction.AddSubAction(fmt.Sprintf("Attempt %d to validate token", attempt), logging.LevelTrace)
		resp, _, reqErr := makeRequest(ctx, URL, "GET", nil, token, false)
		if reqErr.Message != "" {
			attemptAction.AppendWarning(fmt.Sprintf("attempt_%d", attempt), map[string]any{"error": reqErr.Message})
			lastErrorMsg = reqErr.Message
			lastErrorDetail = reqErr.Detail
			if attempt < 3 {
				time.Sleep(1 * time.Second)
			}
			continue
		} else {
			defer resp.Body.Close()
			isValid = true
			return isValid, logging.LogErrorInfo{}
		}
	}

	// All attempts failed
	logAction.SetError(lastErrorMsg, "MediUX token validation failed after 3 attempts", lastErrorDetail)
	return isValid, *logAction.Error
}
