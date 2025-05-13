package plex

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"poster-setter/internal/config"
	"poster-setter/internal/logging"
	"poster-setter/internal/utils"
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

	ratingKey := chi.URLParam(r, "ratingKey")
	imageType := chi.URLParam(r, "imageType")
	if ratingKey == "" || imageType == "" {
		utils.SendErrorJSONResponse(w, http.StatusInternalServerError, logging.ErrorLog{Err: errors.New("missing rating key or image type"),
			Log: logging.Log{
				Message: "Missing rating key or image type in URL",
				Elapsed: utils.ElapsedTime(startTime),
			},
		})
		return
	} else if imageType != "poster" && imageType != "backdrop" {
		utils.SendErrorJSONResponse(w, http.StatusInternalServerError, logging.ErrorLog{Err: errors.New("invalid image type"),
			Log: logging.Log{
				Message: "Invalid image type in URL",
				Elapsed: utils.ElapsedTime(startTime),
			},
		})
		return
	}

	// Check if the temporary folder has the image
	fileName := fmt.Sprintf("%s_%s.jpg", ratingKey, imageType)
	filePath := path.Join(PlexTempImageFolder, fileName)
	exists := utils.CheckIfImageExists(filePath)
	if exists {
		logging.LOG.Trace(fmt.Sprintf("Image %s already exists in temporary folder", fileName))
		// Serve the image from the temporary folder
		imagePath := path.Join(PlexTempImageFolder, fileName)
		http.ServeFile(w, r, imagePath)
		return
	}

	logging.LOG.Trace(fmt.Sprintf("Image %s does not exist in temporary folder", fileName))
	// If the image does not exist, then get it from Plex
	imageData, logErr := FetchImageFromMediaServer(ratingKey, imageType)
	if logErr.Err != nil {
		utils.SendErrorJSONResponse(w, http.StatusInternalServerError, logErr)
		return
	}

	if config.Global.CacheImages {
		// Add the image to the temporary folder
		logErr = utils.CheckFolderExists(PlexTempImageFolder)
		if logErr.Err != nil {
			utils.SendErrorJSONResponse(w, http.StatusInternalServerError, logErr)
			return
		}
		imagePath := fmt.Sprintf("%s/%s", PlexTempImageFolder, fileName)
		err := os.WriteFile(imagePath, imageData, 0644)
		if err != nil {
			utils.SendErrorJSONResponse(w, http.StatusInternalServerError, logging.ErrorLog{Err: err,
				Log: logging.Log{
					Message: "Failed to write image to temporary folder",
					Elapsed: utils.ElapsedTime(startTime),
				},
			})
			return
		}
	}

	// Set the content type for the response
	w.Header().Set("Content-Type", "image/jpeg")
	w.WriteHeader(http.StatusOK)
	w.Write(imageData)
}

func FetchImageFromMediaServer(ratingKey string, imageType string) ([]byte, logging.ErrorLog) {
	logging.LOG.Trace(fmt.Sprintf("Getting %s for rating key: %s", imageType, ratingKey))

	// Construct the URL for the Plex server API request
	//plexURL := fmt.Sprintf("%s/library/metadata/%s/%s/", config.Global.MediaServer.URL, ratingKey, imageType)
	photoUrl := fmt.Sprintf("/library/metadata/%s/%s/", ratingKey, imageType)
	// Encode the URL for the request
	encodedPhotoUrl := url.QueryEscape(photoUrl)
	plexURL := fmt.Sprintf("%s/photo/:/transcode?url=%s&width=300&height=450", config.Global.MediaServer.URL, encodedPhotoUrl)
	if imageType == "art" {
		plexURL = fmt.Sprintf("%s/photo/:/transcode?url=%s&width=1920&height=1080", config.Global.MediaServer.URL, encodedPhotoUrl)
	}

	response, body, logErr := utils.MakeHTTPRequest(plexURL, "GET", nil, 60, nil, "MediaServer")
	if logErr.Err != nil {
		return nil, logErr
	}
	defer response.Body.Close()

	// Check if the response status is OK
	if response.StatusCode != http.StatusOK {
		return nil, logging.ErrorLog{Err: errors.New("plex server error"),
			Log: logging.Log{Message: fmt.Sprintf("Received status code '%d' from Plex server", response.StatusCode)},
		}
	}
	// Check if the response body is empty
	if len(body) == 0 {
		return nil, logging.ErrorLog{Err: errors.New("empty response body"),
			Log: logging.Log{Message: "Plex returned an empty response body"},
		}
	}

	// Return the image data
	return body, logging.ErrorLog{}
}
