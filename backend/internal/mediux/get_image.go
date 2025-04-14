package mediux

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"poster-setter/internal/config"
	"poster-setter/internal/logging"
	"poster-setter/internal/utils"
	"time"

	"github.com/go-chi/chi/v5"
)

var MediuxTempImageFolder string

func init() {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "/config"
	}
	MediuxTempImageFolder = path.Join(configPath, "temp-images", "mediux")
}

func GetMediuxImage(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	logging.LOG.Trace(r.URL.Path)

	// Get the asset ID from the URL
	assetID := chi.URLParam(r, "assetID")
	if assetID == "" {
		utils.SendErrorJSONResponse(w, http.StatusInternalServerError, logging.ErrorLog{
			Log: logging.Log{
				Message: "Missing asset ID in URL",
				Elapsed: utils.ElapsedTime(startTime),
			},
		})
		return
	}

	// Check if the temporary folder has the image
	fileName := fmt.Sprintf("%s.jpg", assetID)
	filePath := path.Join(MediuxTempImageFolder, fileName)
	exists := utils.CheckIfImageExists(filePath)
	if exists {
		logging.LOG.Trace(fmt.Sprintf("Image %s already exists in temporary folder", fileName))
		// Serve the image from the temporary folder
		imagePath := path.Join(MediuxTempImageFolder, fileName)
		http.ServeFile(w, r, imagePath)
		return
	}

	// If the image does not exist, then get it from Mediux
	imageData, imageType, logErr := FetchImage(assetID)
	if logErr.Err != nil {
		utils.SendErrorJSONResponse(w, http.StatusInternalServerError, logErr)
		return
	}

	if config.Global.CacheImages {
		// Add the image to the temporary folder
		imagePath := path.Join(MediuxTempImageFolder, fileName)
		logErr = utils.CheckFolderExists(MediuxTempImageFolder)
		if logErr.Err != nil {
			utils.SendErrorJSONResponse(w, http.StatusInternalServerError, logErr)
			return
		}
		err := os.WriteFile(imagePath, imageData, 0644)
		if err != nil {
			utils.SendErrorJSONResponse(w, http.StatusInternalServerError, logging.ErrorLog{
				Log: logging.Log{
					Message: "Failed to write image to temporary folder",
					Elapsed: utils.ElapsedTime(startTime),
				},
			})
			return
		}
	}

	w.Header().Set("Content-Type", imageType)
	w.WriteHeader(http.StatusOK)
	w.Write(imageData)
}

func FetchImage(assetID string) ([]byte, string, logging.ErrorLog) {
	logging.LOG.Trace(fmt.Sprintf("Getting image for asset ID: %s", assetID))

	// Construct the URL for the Mediux API request
	url := fmt.Sprintf("%s/%s", "https://staged.mediux.io/assets", assetID)

	response, body, logErr := utils.MakeHTTPRequest(url, "GET", nil, 30, nil, "Mediux")
	if logErr.Err != nil {
		return nil, "", logErr
	}
	defer response.Body.Close()

	// Check if the response body is empty
	if len(body) == 0 {
		return nil, "", logging.ErrorLog{
			Err: errors.New("empty response body"),
			Log: logging.Log{Message: "Mediux returned an empty response body"},
		}
	}

	// Get the image type from the response headers
	imageType := response.Header.Get("Content-Type")
	if imageType == "" {
		return nil, "", logging.ErrorLog{
			Err: errors.New("missing content type in response headers"),
			Log: logging.Log{Message: "Mediux did not return a content type in the response headers"},
		}
	}

	if response.StatusCode != http.StatusOK {
		return nil, "", logging.ErrorLog{
			Err: errors.New("failed to fetch image from Mediux"),
			Log: logging.Log{Message: fmt.Sprintf("Mediux returned status code %d", response.StatusCode)},
		}
	}

	// Return the image data
	return body, imageType, logging.ErrorLog{}
}
