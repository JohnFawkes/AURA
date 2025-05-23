package emby_jellyfin

import (
	"aura/internal/config"
	"aura/internal/logging"
	"aura/internal/mediux"
	"aura/internal/modals"
	"aura/internal/utils"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"path"
)

func DownloadAndUpdatePosters(item modals.MediaItem, file modals.PosterFile) logging.ErrorLog {

	itemRatingKey := getItemRatingKey(item, file)
	if itemRatingKey == "" {
		return logging.ErrorLog{Err: fmt.Errorf("media not found"),
			Log: logging.Log{Message: "Media not found"}}
	}

	// Check if the temporary folder has the image
	// If it does, we don't need to download it again
	// If it doesn't, we need to download it
	// The image is saved in the temp-images/mediux/full folder with the file ID as the name
	formatDate := file.Modified.Format("20060102")
	fileName := fmt.Sprintf("%s_%s.jpg", file.ID, formatDate)
	filePath := path.Join(mediux.MediuxFullTempImageFolder, fileName)
	exists := utils.CheckIfImageExists(filePath)
	var imageData []byte
	if !exists {
		// Check if the temporary folder exists
		logErr := utils.CheckFolderExists(mediux.MediuxFullTempImageFolder)
		if logErr.Err != nil {
			return logErr
		}
		// Download the image from Mediux
		imageData, _, logErr = mediux.FetchImage(file.ID, formatDate, true)
		if logErr.Err != nil {
			return logErr
		}
		// Add the image to the temporary folder
		err := os.WriteFile(filePath, imageData, 0644)
		if err != nil {
			return logging.ErrorLog{Err: err, Log: logging.Log{Message: fmt.Sprintf("Failed to write image to %s: %v", filePath, err)}}
		}
		logging.LOG.Trace(fmt.Sprintf("Image %s downloaded and saved to temporary folder", file.ID))
	} else {
		// Read the contents of the file
		var err error
		imageData, err = os.ReadFile(filePath)
		if err != nil {
			return logging.ErrorLog{Err: err, Log: logging.Log{Message: fmt.Sprintf("Failed to read image from %s: %v", filePath, err)}}
		}
	}

	var posterType string
	if file.Type == "backdrop" {
		posterType = "Backdrop"
	} else {
		posterType = "Primary"
	}

	baseURL, logErr := utils.MakeMediaServerAPIURL(fmt.Sprintf("/Items/%s/Images/%s", itemRatingKey, posterType), config.Global.MediaServer.URL)
	if logErr.Err != nil {
		return logErr
	}

	// Encode the image data as Base64
	base64ImageData := base64.StdEncoding.EncodeToString(imageData)

	// Make a POST request to the Emby/Jellyfin server
	headers := map[string]string{
		"Content-Type": "image/jpeg",
	}
	response, _, logErr := utils.MakeHTTPRequest(baseURL.String(), "POST", headers, 60, []byte(base64ImageData), "MediaServer")
	if logErr.Err != nil {
		return logErr
	}
	defer response.Body.Close()

	// Check if the response status is OK
	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusNoContent {
		return logging.ErrorLog{Err: fmt.Errorf("received status code '%d' from %s server", response.StatusCode, config.Global.MediaServer.Type),
			Log: logging.Log{Message: fmt.Sprintf("Received status code '%d' from %s server", response.StatusCode, config.Global.MediaServer.Type)}}
	}

	return logging.ErrorLog{}
}
