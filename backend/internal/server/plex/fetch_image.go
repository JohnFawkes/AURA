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

	"github.com/go-chi/chi/v5"
)

var PlexTempImageFolder string

func init() {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "/config"
	}
	PlexTempImageFolder = path.Join(configPath, "temp-images", "plex")
}

func GetPlexImage(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	logging.LOG.Trace(r.URL.Path)
	Err := logging.NewStandardError()

	ratingKey := chi.URLParam(r, "ratingKey")
	imageType := chi.URLParam(r, "imageType")
	if ratingKey == "" || imageType == "" {

		Err.Message = "Missing rating key or image type"
		Err.HelpText = "Ensure the URL contains both rating key and image type parameters."
		Err.Details = fmt.Sprintf("Received ratingKey: %s, imageType: %s", ratingKey, imageType)
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	} else if imageType != "poster" && imageType != "backdrop" {

		Err.Message = "Invalid image type"
		Err.HelpText = "Image type must be either 'poster' or 'backdrop'."
		Err.Details = fmt.Sprintf("Received image type: %s", imageType)
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}

	// If the image does not exist, then get it from Plex
	imageData, Err := FetchImageFromMediaServer(ratingKey, imageType)
	if Err.Message != "" {
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}

	// Set the content type for the response
	w.Header().Set("Content-Type", "image/jpeg")
	w.WriteHeader(http.StatusOK)
	w.Write(imageData)
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

	// Check if the response body is empty
	if len(body) == 0 {

		Err.Message = "Received empty response from Plex server"
		Err.HelpText = fmt.Sprintf("Ensure the Plex server is running and the item with rating key %s exists.", ratingKey)
		Err.Details = "The Plex server returned an empty response for the requested image."
		return nil, Err
	}

	// Return the image data
	return body, logging.StandardError{}
}
