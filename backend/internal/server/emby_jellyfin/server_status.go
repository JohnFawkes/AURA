package emby_jellyfin

import (
	"aura/internal/config"
	"aura/internal/logging"
	"aura/internal/utils"
	"encoding/json"
	"net/http"
)

func GetMediaServerStatus() (string, logging.ErrorLog) {
	logging.LOG.Trace("Checking Emby/Jellyfin server status")

	baseURL, logErr := utils.MakeMediaServerAPIURL("/System/Info", config.Global.MediaServer.URL)
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
				Message: "Failed to get Emby/Jellyfin server status",
			},
		}
	}

	var statusResponse struct {
		Version string `json:"Version"`
	}

	err := json.Unmarshal(body, &statusResponse)
	if err != nil {
		return "", logging.ErrorLog{
			Err: err,
			Log: logging.Log{
				Message: "Failed to unmarshal Emby/Jellyfin server response",
			},
		}
	}

	status := statusResponse.Version

	if status == "" {
		return "", logging.ErrorLog{
			Err: logErr.Err,
			Log: logging.Log{
				Message: "Emby/Jellyfin server returned an empty version",
			},
		}
	}

	return status, logging.ErrorLog{}
}
