package emby_jellyfin

import (
	"aura/internal/config"
	"aura/internal/logging"
	"aura/internal/utils"
	"encoding/json"
)

func GetMediaServerStatus() (string, logging.StandardError) {
	logging.LOG.Trace("Checking Emby/Jellyfin server status")
	Err := logging.NewStandardError()

	baseURL, Err := utils.MakeMediaServerAPIURL("/System/Info", config.Global.MediaServer.URL)
	if Err.Message != "" {
		return "", Err
	}

	response, body, Err := utils.MakeHTTPRequest(baseURL.String(), "GET", nil, 60, nil, "MediaServer")
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
