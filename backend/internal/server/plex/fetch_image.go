package plex

import (
	"aura/internal/config"
	"aura/internal/logging"
	"aura/internal/utils"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"time"
)

var PlexTempImageFolder string

func init() {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "/config"
	}
	PlexTempImageFolder = path.Join(configPath, "temp-images", "plex")
}

func FetchImageFromMediaServer(ratingKey string, imageType string) ([]byte, logging.StandardError) {
	Err := logging.NewStandardError()

	// Construct the URL for the Plex server API request
	if imageType == "backdrop" {
		imageType = "art"
	}
	photoUrl := fmt.Sprintf("/library/metadata/%s/%s", ratingKey, imageType)
	// Encode the URL for the request
	encodedPhotoUrl := url.QueryEscape(photoUrl + fmt.Sprintf("/%d", time.Now().Unix()))
	plexURL := fmt.Sprintf("%s/photo/:/transcode?width=300&height=450&url=%s", config.Global.MediaServer.URL, encodedPhotoUrl)
	if imageType == "art" {
		plexURL = fmt.Sprintf("%s/photo/:/transcode?width=1920&height=1080&url=%s", config.Global.MediaServer.URL, encodedPhotoUrl)
	}

	response, body, Err := utils.MakeHTTPRequest(plexURL, "GET", nil, 60, nil, "MediaServer")
	if Err.Message != "" {
		return nil, Err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		Err.Message = "Failed to fetch image from Plex server"
		Err.HelpText = fmt.Sprintf("Ensure the item with rating key %s exists. If it does, check the Plex server logs for more information. If it doesn't, please try refreshing aura from the Home page.", ratingKey)
		Err.Details = map[string]any{
			"statusCode": response.StatusCode,
			"ratingKey":  ratingKey,
			"imageType":  imageType,
			"request":    plexURL,
		}
		return nil, Err
	}

	// Check if the response body is empty
	if len(body) == 0 {
		Err.Message = "Received empty response from Plex server"
		Err.HelpText = fmt.Sprintf("Ensure the item with rating key %s exists. If it does, check the Plex server logs for more information. If it doesn't, please try refreshing aura from the Home page.", ratingKey)
		Err.Details = map[string]any{
			"statusCode": response.StatusCode,
			"ratingKey":  ratingKey,
			"imageType":  imageType,
			"request":    plexURL,
		}
		return nil, Err
	}

	// Return the image data
	return body, logging.StandardError{}
}
