package api

import (
	"aura/internal/logging"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
)

func Plex_RateMediaItem(ctx context.Context, ratingKey string, rating float64) logging.LogErrorInfo {
	ctx, logAction := logging.AddSubActionToContext(ctx, fmt.Sprintf("Rating Item with Rating Key '%s' on Plex", ratingKey), logging.LevelDebug)
	defer logAction.Complete()

	// Make the URL
	u, err := url.Parse(Global_Config.MediaServer.URL)
	if err != nil {
		logAction.SetError("Failed to parse Plex base URL", err.Error(), nil)
		return *logAction.Error
	}
	u.Path = path.Join(u.Path, ":", "rate")
	query := u.Query()
	query.Add("identifier", "com.plexapp.plugins.library")
	query.Add("key", ratingKey)
	query.Add("rating", fmt.Sprintf("%.1f", rating))
	u.RawQuery = query.Encode()
	URL := u.String()

	// Make the Auth Headers for Request
	headers := MakeAuthHeader("X-Plex-Token", Global_Config.MediaServer.Token)

	// Make the API request to Plex
	httpResp, _, logErr := MakeHTTPRequest(ctx, URL, http.MethodPut, headers, 30, nil, "Plex")
	if logErr.Message != "" {
		return logErr
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		logAction.SetError("Plex returned non-OK status when rating item", fmt.Sprintf("Status Code: %d", httpResp.StatusCode), nil)
		return *logAction.Error
	}

	return logging.LogErrorInfo{}
}
