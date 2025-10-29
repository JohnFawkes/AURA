package api

import (
	"aura/internal/logging"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
)

func SR_CallUpdateItemInfo(ctx context.Context, app Config_SonarrRadarrApp, item any) logging.LogErrorInfo {
	ctx, logAction := logging.AddSubActionToContext(ctx, fmt.Sprintf("Updating %s Item Info in %s (%s)", app.Type, app.Type, app.Library), logging.LevelInfo)
	defer logAction.Complete()

	// Determine endpoint and item ID
	urlEndPoint := ""
	var itemID int64
	switch app.Type {
	case "Sonarr":
		srItem, ok := item.(SR_SonarrItem)
		if !ok {
			logAction.SetError("Invalid item type for Sonarr", "Ensure the item is of type SR_SonarrItem", nil)
			return *logAction.Error
		}
		urlEndPoint = "series"
		itemID = srItem.ID
	case "Radarr":
		srItem, ok := item.(SR_RadarrItem)
		if !ok {
			logAction.SetError("Invalid item type for Radarr", "Ensure the item is of type SR_RadarrItem", nil)
			return *logAction.Error
		}
		urlEndPoint = "movie"
		itemID = srItem.ID
	default:
		logAction.SetError("Unknown application type", "Ensure the application type is either 'Sonarr' or 'Radarr'", map[string]any{"app_type": app.Type})
		return *logAction.Error
	}

	// Construct URL
	u, err := url.Parse(app.URL)
	if err != nil {
		logAction.SetError("Failed to parse base URL", "Ensure the URL is valid", map[string]any{"error": err.Error()})
		return *logAction.Error
	}
	u.Path = path.Join(u.Path, "api/v3", urlEndPoint, fmt.Sprintf("%d", itemID))
	URL := u.String()

	// Set headers
	apiHeader := map[string]string{
		"X-Api-Key":    app.APIKey,
		"Content-Type": "application/json",
	}

	// Marshal the item to JSON
	payload, err := json.Marshal(item)
	if err != nil {
		logAction.SetError("Failed to marshal item to JSON", "Ensure the item structure is valid", map[string]any{"error": err.Error()})
		return *logAction.Error
	}

	// Make the request
	httpResp, _, Err := MakeHTTPRequest(ctx, URL, http.MethodPut, apiHeader, 60, payload, app.Type)
	if Err.Message != "" {
		return Err
	}

	// Check for valid status code
	if httpResp.StatusCode != http.StatusOK && httpResp.StatusCode != http.StatusCreated && httpResp.StatusCode != http.StatusAccepted {
		logAction.SetError("Unexpected HTTP status code", "Ensure the API is working correctly", map[string]any{"status_code": httpResp.StatusCode})
		return *logAction.Error
	}

	return logging.LogErrorInfo{}
}
