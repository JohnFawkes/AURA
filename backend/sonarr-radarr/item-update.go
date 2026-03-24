package sonarr_radarr

import (
	"aura/config"
	"aura/logging"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"path"
)

func UpdateItemInfo(ctx context.Context, app config.Config_SonarrRadarrApp, item any) (Err logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, fmt.Sprintf("Updating Item Info in %s | %s", app.Type, app.Library), logging.LevelInfo)
	defer logAction.Complete()

	// Determine the endpoint and the item ID
	urlEndpoint := ""
	var itemID int64
	switch app.Type {
	case "Sonarr":
		srItem, ok := item.(SR_SonarrItem)
		if !ok {
			logAction.SetError("Invalid Item Type for Sonarr",
				"Ensure the item provided is of type SonarrItem",
				map[string]any{
					"item_type": fmt.Sprintf("%T", item),
				})
			return *logAction.Error
		}
		itemID = srItem.ID
		urlEndpoint = "series"
	case "Radarr":
		srItem, ok := item.(SR_RadarrItem)
		if !ok {
			logAction.SetError("Invalid Item Type for Radarr",
				"Ensure the item provided is of type RadarrItem",
				map[string]any{
					"item_type": fmt.Sprintf("%T", item),
				})
			return *logAction.Error
		}
		itemID = srItem.ID
		urlEndpoint = "movie"
	default:
		logAction.SetError("Unsupported Sonarr/Radarr Type",
			"Ensure the Sonarr/Radarr Type is either 'Sonarr' or 'Radarr'",
			map[string]any{
				"type": app.Type,
			})
		return *logAction.Error
	}

	// Construct the URL
	u, err := url.Parse(app.URL)
	if err != nil {
		logAction.SetError(fmt.Sprintf("Invalid %s URL", app.Type),
			"Make sure that the URL is properly formatted",
			map[string]any{
				"url":   app.URL,
				"error": err.Error(),
			})
		return *logAction.Error
	}
	u.Path = path.Join(u.Path, "api", "v3", urlEndpoint, fmt.Sprintf("%d", itemID))
	URL := u.String()

	// Marshal the item to JSON
	body, err := json.Marshal(item)
	if err != nil {
		logAction.SetError("Failed to Marshal Item to JSON",
			"Ensure the item can be properly converted to JSON format",
			map[string]any{
				"item_id": itemID,
				"error":   err.Error(),
			})
		return *logAction.Error
	}

	// Make the request to Sonarr/Radarr
	_, _, Err = makeRequest(ctx, app, URL, "PUT", body)
	if Err.Message != "" {
		logAction.SetErrorFromInfo(Err)
		return *logAction.Error
	}

	return logging.LogErrorInfo{}
}
