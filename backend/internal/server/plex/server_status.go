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

	baseURL, Err := utils.MakeMediaServerAPIURL("/", config.Global.MediaServer.URL)
	if Err.Message != "" {
		return "", Err
	}

	httpResponse, body, Err := utils.MakeHTTPRequest(baseURL.String(), "GET", nil, 60, nil, "MediaServer")
	if Err.Message != "" {
		return "", Err
	}
	defer httpResponse.Body.Close()

	// Check to see if the Status is OK
	if httpResponse.StatusCode != 200 {
		Err.Message = "Failed to connect to Plex server"
		Err.HelpText = "Ensure the Plex server is running and accessible at the configured URL with the correct token."
		Err.Details = "HTTP Status: " + httpResponse.Status
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
