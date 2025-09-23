package plex

import (
	"aura/internal/config"
	"aura/internal/logging"
	"aura/internal/modals"
	"aura/internal/utils"
	"encoding/json"
)

func GetMediaServerStatus() (string, logging.StandardError) {
	logging.LOG.Trace("Checking Plex server status")
	Err := logging.NewStandardError()

	baseURL, Err := utils.MakeMediaServerAPIURL("/identity", config.Global.MediaServer.URL)
	if Err.Message != "" {
		return "", Err
	}

	httpResponse, body, Err := utils.MakeHTTPRequest(baseURL.String(), "GET", nil, 60, nil, "MediaServer")
	if Err.Message != "" {
		return "", Err
	}
	defer httpResponse.Body.Close()

	// Check if the httpResponse body is empty
	if len(body) == 0 {
		Err.Message = "Received empty response body from Plex server"
		Err.HelpText = "Ensure the Plex server is running and accessible at the configured URL."
		return "", Err
	}

	var plexResponse modals.PlexConnectionInfoWrapper
	err := json.Unmarshal(body, &plexResponse)
	if err != nil {
		Err.Message = "Failed to parse Plex server response"
		Err.HelpText = "Ensure the Plex server is returning a valid JSON response."
		Err.Details = "Error: " + err.Error()
		return "", Err
	}

	// Get the server version
	serverVersion := plexResponse.MediaContainer.Version
	if serverVersion == "" {
		Err.Message = "Failed to retrieve Plex server version"
		Err.HelpText = "Ensure the Plex server is running and accessible at the configured URL."
		Err.Details = "Response: " + string(body)
		return "", Err
	}

	return serverVersion, logging.StandardError{}
}
