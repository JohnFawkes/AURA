package api

import (
	"aura/internal/logging"
	"encoding/json"
	"fmt"
	"net/url"
	"path"
)

func SR_CallUpdateItemInfo(app Config_SonarrRadarrApp, item any) logging.StandardError {
	Err := logging.NewStandardError()

	// Determine endpoint and item ID
	urlEndPoint := ""
	var itemID int64
	switch app.Type {
	case "Sonarr":
		srItem, ok := item.(SR_SonarrItem)
		if !ok {
			Err.Message = "Item is not of type SR_SonarrItem"
			return Err
		}
		urlEndPoint = "series"
		itemID = srItem.ID
	case "Radarr":
		srItem, ok := item.(SR_RadarrItem)
		if !ok {
			Err.Message = "Item is not of type SR_RadarrItem"
			return Err
		}
		urlEndPoint = "movie"
		itemID = srItem.ID
	default:
		Err.Message = "Unknown service type"
		return Err
	}

	// Construct the URL for updating the item
	u, _ := url.Parse(app.URL)
	u.Path = path.Join(u.Path, "api/v3", urlEndPoint, fmt.Sprintf("%d", itemID))
	updateURL := u.String()

	apiHeader := map[string]string{
		"X-Api-Key": app.APIKey,
	}

	// Marshal the item to JSON
	payload, err := json.Marshal(item)
	if err != nil {
		Err.Message = "Failed to marshal item for update"
		Err.Details = map[string]any{"error": err.Error()}
		return Err
	}

	response, body, Err := MakeHTTPRequest(updateURL, "PUT", apiHeader, 60, payload, app.Type)
	if Err.Message != "" {
		return Err
	}
	defer response.Body.Close()

	// Check to see if Status is OK
	if response.StatusCode != 200 && response.StatusCode != 201 && response.StatusCode != 202 {
		Err.Message = fmt.Sprintf("%s returned non-200 status code: %d", app.Type, response.StatusCode)
		Err.HelpText = fmt.Sprintf("Check the %s URL and API key in the configuration.", app.Type)
		Err.Details = map[string]any{
			"status_code":   response.StatusCode,
			"response_body": string(body),
			"request_url":   updateURL,
			"request_body":  string(payload),
		}
		logging.LOG.Error(fmt.Sprintf("Error details: %+v", Err.Details))
		return Err
	}

	return Err
}
