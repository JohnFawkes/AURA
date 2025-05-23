package plex

import (
	"aura/internal/config"
	"aura/internal/logging"
	"aura/internal/utils"
	"errors"
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

	// If the image does not exist, then get it from Plex
	imageData, logErr := FetchImageFromMediaServer(ratingKey, imageType)
	if logErr.Err != nil {
		utils.SendErrorJSONResponse(w, http.StatusInternalServerError, logErr)
		return
	}

	// Set the content type for the response
	w.Header().Set("Content-Type", "image/jpeg")
	w.WriteHeader(http.StatusOK)
	w.Write(imageData)
}

func FetchImageFromMediaServer(ratingKey string, imageType string) ([]byte, logging.ErrorLog) {
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
