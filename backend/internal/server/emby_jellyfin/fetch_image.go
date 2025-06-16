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

func FetchImageFromMediaServer(ratingKey, imageType string) ([]byte, logging.StandardError) {
	Err := logging.NewStandardError()
	if imageType == "poster" {
		imageType = "Primary"
	} else if imageType == "backdrop" {
		imageType = "Backdrop"
	} else {

		Err.Message = "Invalid image type"
		Err.HelpText = "Image type must be either 'poster' or 'backdrop'."
		Err.Details = fmt.Sprintf("Received image type: %s", imageType)
		return nil, Err
	}

	baseURL, Err := utils.MakeMediaServerAPIURL(fmt.Sprintf("Items/%s/Images/%s", ratingKey, imageType), config.Global.MediaServer.URL)
	if Err.Message != "" {
		return nil, Err
	}

	response, body, Err := utils.MakeHTTPRequest(baseURL.String(), http.MethodGet, nil, 60, nil, "MediaServer")
	if Err.Message != "" {
		return nil, Err
	}
	defer response.Body.Close()

	// Check if the response is empty
	if len(body) == 0 {

		Err.Message = "Received empty response from media server"
		Err.HelpText = fmt.Sprintf("Ensure the media server is running and the item with rating key %s exists.", ratingKey)
		Err.Details = "The media server returned an empty response for the requested image."
		return nil, Err
	}

	// Return the image data
	return body, logging.StandardError{}
}
