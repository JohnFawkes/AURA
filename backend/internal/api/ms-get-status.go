package api

import (
	"aura/internal/logging"
	"encoding/json"
)

func (p *PlexServer) GetMediaServerStatus(msConfig Config_MediaServer) (string, logging.StandardError) {
	// Get the status of the Plex server
	version, Err := Plex_GetMediaServerStatus(msConfig)
	if Err.Message != "" {
		return "", Err
	}
	return version, logging.StandardError{}
}

func (e *EmbyJellyServer) GetMediaServerStatus(msConfig Config_MediaServer) (string, logging.StandardError) {
	//Get the status of the Emby/Jellyfin server
	version, Err := EJ_GetMediaServerStatus(msConfig)
	if Err.Message != "" {
		return "", Err
	}
	return version, logging.StandardError{}
}

//////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////

func Plex_GetMediaServerStatus(msConfig Config_MediaServer) (string, logging.StandardError) {
	logging.LOG.Trace("Checking Plex server status")
	Err := logging.NewStandardError()

	baseURL, Err := MakeMediaServerAPIURL("/", msConfig.URL)
	if Err.Message != "" {
		return "", Err
	}

	httpResponse, body, Err := MakeHTTPRequest(baseURL.String(), "GET", nil, 60, nil, "Plex")
	if Err.Message != "" {
		return "", Err
	}
	defer httpResponse.Body.Close()

	// Check to see if the Status is OK
	if httpResponse.StatusCode != 200 {
		Err.Message = "Failed to connect to Plex server"
		Err.HelpText = "Ensure the Plex server is running and accessible at the configured URL with the correct token."
		Err.Details = map[string]any{
			"statusCode": httpResponse.StatusCode,
			"request":    baseURL.String(),
		}
		return "", Err
	}

	var plexResponse PlexConnectionInfoWrapper
	err := json.Unmarshal(body, &plexResponse)
	if err != nil {
		Err.Message = "Failed to parse Plex server response"
		Err.HelpText = "Ensure the Plex server is returning a valid JSON response."
		Err.Details = map[string]any{
			"error":   err.Error(),
			"request": baseURL.String(),
		}
		return "", Err
	}

	// Get the server version
	serverVersion := plexResponse.MediaContainer.Version
	if serverVersion == "" {
		Err.Message = "Failed to retrieve Plex server version"
		Err.HelpText = "Ensure the Plex server is running and accessible at the configured URL."
		Err.Details = map[string]any{
			"request": baseURL.String(),
		}
		return "", Err
	}

	return serverVersion, logging.StandardError{}
}

//////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////

func EJ_GetMediaServerStatus(msConfig Config_MediaServer) (string, logging.StandardError) {
	logging.LOG.Trace("Checking Emby/Jellyfin server status")
	Err := logging.NewStandardError()

	baseURL, Err := MakeMediaServerAPIURL("/System/Info", msConfig.URL)
	if Err.Message != "" {
		return "", Err
	}

	response, body, Err := MakeHTTPRequest(baseURL.String(), "GET", nil, 60, nil, "MediaServer")
	if Err.Message != "" {
		return "", Err
	}
	defer response.Body.Close()

	var statusResponse struct {
		Version string `json:"Version"`
	}

	err := json.Unmarshal(body, &statusResponse)
	if err != nil {
		Err.Message = "Failed to parse JSON response from media server"
		Err.HelpText = "Ensure the Emby/Jellyfin server is returning a valid JSON response."
		Err.Details = map[string]any{
			"error": err.Error(),
		}
		return "", Err
	}

	status := statusResponse.Version

	if status == "" {
		Err.Message = "Received empty status from media server"
		Err.HelpText = "Ensure the media server is running and accessible."
		Err.Details = map[string]any{
			"statusCode": response.StatusCode,
		}
		return "", Err
	}

	return status, logging.StandardError{}
}
