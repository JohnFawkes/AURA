package emby_jellyfin

import (
	"aura/internal/config"
	"aura/internal/logging"
	"aura/internal/utils"
	"fmt"
	"net/http"
	"os"
	"path"
)

var EmbyJellyTempImageFolder string

func init() {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "/config"
	}
	EmbyJellyTempImageFolder = path.Join(configPath, "temp-images", "emby-jellyfin")
}

func FetchImageFromMediaServer(ratingKey, imageType string) ([]byte, logging.ErrorLog) {
	if imageType == "poster" {
		imageType = "Primary"
	} else if imageType == "backdrop" {
		imageType = "Backdrop"
	}

	baseURL, logErr := utils.MakeMediaServerAPIURL(fmt.Sprintf("Items/%s/Images/%s", ratingKey, imageType), config.Global.MediaServer.URL)
	if logErr.Err != nil {
		return nil, logErr
	}

	response, body, logErr := utils.MakeHTTPRequest(baseURL.String(), http.MethodGet, nil, 60, nil, "MediaServer")
	if logErr.Err != nil {
		return nil, logErr
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, logging.ErrorLog{Err: fmt.Errorf("received status code '%d' from %s server", response.StatusCode, config.Global.MediaServer.Type),
			Log: logging.Log{Message: fmt.Sprintf("Received status code '%d' from %s server", response.StatusCode, config.Global.MediaServer.Type)}}
	}

	// Check if the response is empty
	if len(body) == 0 {
		return nil, logging.ErrorLog{Err: fmt.Errorf("received empty response from %s server", config.Global.MediaServer.Type),
			Log: logging.Log{Message: fmt.Sprintf("Received empty response from %s server", config.Global.MediaServer.Type)}}
	}

	// Return the image data
	return body, logging.ErrorLog{}
}
