package api

import (
	"aura/internal/logging"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"time"
)

var PlexTempImageFolder string
var EmbyJellyTempImageFolder string

func init() {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "/config"
	}
	EmbyJellyTempImageFolder = path.Join(configPath, "temp-images", "emby-jellyfin")
	PlexTempImageFolder = path.Join(configPath, "temp-images", "plex")
}

func (p *PlexServer) FetchImageFromMediaServer(ratingKey, imageType string) ([]byte, logging.StandardError) {
	// Fetch the image from Plex
	imageData, Err := Plex_FetchImageFromMediaServer(ratingKey, imageType)
	if Err.Message != "" {
		return nil, Err
	}
	return imageData, logging.StandardError{}
}

func (e *EmbyJellyServer) FetchImageFromMediaServer(ratingKey, imageType string) ([]byte, logging.StandardError) {
	// Fetch the image from Emby/Jellyfin
	imageData, Err := EJ_FetchImageFromMediaServer(ratingKey, imageType)
	if Err.Message != "" {
		return nil, Err
	}
	return imageData, logging.StandardError{}
}

///////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////

func Plex_FetchImageFromMediaServer(ratingKey string, imageType string) ([]byte, logging.StandardError) {
	Err := logging.NewStandardError()

	// Construct the URL for the Plex server API request
	if imageType == "backdrop" {
		imageType = "art"
	}
	photoUrl := fmt.Sprintf("/library/metadata/%s/%s", ratingKey, imageType)
	// Encode the URL for the request
	encodedPhotoUrl := url.QueryEscape(photoUrl + fmt.Sprintf("/%d", time.Now().Unix()))
	plexURL := fmt.Sprintf("%s/photo/:/transcode?width=300&height=450&url=%s", Global_Config.MediaServer.URL, encodedPhotoUrl)
	if imageType == "art" {
		plexURL = fmt.Sprintf("%s/photo/:/transcode?width=1920&height=1080&url=%s", Global_Config.MediaServer.URL, encodedPhotoUrl)
	}

	response, body, Err := MakeHTTPRequest(plexURL, "GET", nil, 60, nil, "MediaServer")
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

///////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////

func EJ_FetchImageFromMediaServer(ratingKey, imageType string) ([]byte, logging.StandardError) {
	Err := logging.NewStandardError()
	switch imageType {
	case "poster":
		imageType = "Primary"
	case "backdrop":
		imageType = "Backdrop"
	default:
		Err.Message = "Invalid image type"
		Err.HelpText = "Image type must be either 'poster' or 'backdrop'."
		Err.Details = map[string]any{
			"receivedImageType": imageType,
		}
		return nil, Err
	}

	baseURL, Err := MakeMediaServerAPIURL(fmt.Sprintf("Items/%s/Images/%s", ratingKey, imageType), Global_Config.MediaServer.URL)
	if Err.Message != "" {
		return nil, Err
	}

	response, body, Err := MakeHTTPRequest(baseURL.String(), http.MethodGet, nil, 60, nil, "MediaServer")
	if Err.Message != "" {
		return nil, Err
	}
	defer response.Body.Close()

	// Check if the response is empty
	if len(body) == 0 {
		Err.Message = "Received empty response from media server"
		Err.HelpText = fmt.Sprintf("Ensure the media server is running and the item with rating key %s exists.", ratingKey)
		Err.Details = map[string]any{
			"statusCode": response.StatusCode,
			"ratingKey":  ratingKey,
			"imageType":  imageType,
		}
		return nil, Err
	}

	// Return the image data
	return body, logging.StandardError{}
}
