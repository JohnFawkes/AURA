package plex

import (
	"aura/internal/config"
	"aura/internal/logging"
	"aura/internal/modals"
	"aura/internal/utils"
	"encoding/xml"
	"net/http"
)

func GetMediaServerStatus() (string, logging.ErrorLog) {
	logging.LOG.Trace("Checking Plex server status")

	baseURL, logErr := utils.MakeMediaServerAPIURL("/identity", config.Global.MediaServer.URL)
	if logErr.Err != nil {
		return "", logErr
	}

	response, body, logErr := utils.MakeHTTPRequest(baseURL.String(), "GET", nil, 60, nil, "MediaServer")
	if logErr.Err != nil {
		return "", logErr
	}
	defer response.Body.Close()

	// Check if the response status is OK
	if response.StatusCode != http.StatusOK {
		return "", logging.ErrorLog{
			Err: logErr.Err,
			Log: logging.Log{
				Message: "Failed to get Plex server status",
			},
		}
	}

	// Check if the response body is empty
	if len(body) == 0 {
		return "", logging.ErrorLog{
			Err: logErr.Err,
			Log: logging.Log{
				Message: "Plex server returned an empty response body",
			},
		}
	}

	var plexResponse modals.PlexResponse
	err := xml.Unmarshal(body, &plexResponse)
	if err != nil {
		return "", logging.ErrorLog{
			Err: err,
			Log: logging.Log{
				Message: "Failed to unmarshal Plex server response",
			},
		}
	}

	// Get the server version
	serverVersion := plexResponse.Version

	if serverVersion == "" {
		return "", logging.ErrorLog{
			Err: logErr.Err,
			Log: logging.Log{
				Message: "Plex server version is empty",
			},
		}
	}

	logging.LOG.Trace("Plex server status retrieved successfully, version: " + serverVersion)
	return serverVersion, logging.ErrorLog{}
}
