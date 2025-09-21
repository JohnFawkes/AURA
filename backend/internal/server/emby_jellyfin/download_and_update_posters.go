package emby_jellyfin

import (
	"aura/internal/config"
	"aura/internal/logging"
	"aura/internal/mediux"
	"aura/internal/modals"
	"aura/internal/utils"
	"encoding/base64"
	"fmt"
	"os"
	"path"
)

func DownloadAndUpdatePosters(item modals.MediaItem, file modals.PosterFile) logging.StandardError {
	Err := logging.NewStandardError()

	itemRatingKey := getItemRatingKey(item, file)
	if itemRatingKey == "" {
		Err.Message = "Media not found"
		Err.HelpText = "Ensure the item exists in the Emby/Jellyfin library."
		Err.Details = fmt.Sprintf("Item Title: %s, Rating Key: %s, File ID: %s", item.Title, itemRatingKey, file.ID)
		return Err
	}

	// Check if the temporary folder has the image
	// If it does, we don't need to download it again
	// If it doesn't, we need to download it
	// The image is saved in the temp-images/mediux/full folder with the file ID as the name
	formatDate := file.Modified.Format("20060102150405")
	fileName := fmt.Sprintf("%s_%s.jpg", file.ID, formatDate)
	filePath := path.Join(mediux.MediuxFullTempImageFolder, fileName)
	exists := utils.CheckIfImageExists(filePath)
	var imageData []byte
	if !exists {
		// Check if the temporary folder exists
		Err := utils.CheckFolderExists(mediux.MediuxFullTempImageFolder)
		if Err.Message != "" {
			return Err
		}
		// Download the image from Mediux
		imageData, _, Err = mediux.FetchImage(file.ID, formatDate, "original")
		if Err.Message != "" {
			return Err
		}
		// Add the image to the temporary folder
		err := os.WriteFile(filePath, imageData, 0644)
		if err != nil {
			Err.Message = "Failed to write image to temporary folder"
			Err.HelpText = fmt.Sprintf("Ensure the path %s is accessible and writable.", mediux.MediuxFullTempImageFolder)
			Err.Details = fmt.Sprintf("Error writing image: %v", err)
			return Err
		}
		logging.LOG.Trace(fmt.Sprintf("Image %s downloaded and saved to temporary folder", file.ID))
	} else {
		// Read the contents of the file
		var err error
		imageData, err = os.ReadFile(filePath)
		if err != nil {
			Err.Message = "Failed to read image from temporary folder"
			Err.HelpText = fmt.Sprintf("Ensure the path %s is accessible and readable.", mediux.MediuxFullTempImageFolder)
			Err.Details = fmt.Sprintf("Error reading image: %v", err)
			return Err
		}
	}

	var posterType string
	if file.Type == "backdrop" {
		posterType = "Backdrop"
	} else {
		posterType = "Primary"
	}

	baseURL, Err := utils.MakeMediaServerAPIURL(fmt.Sprintf("/Items/%s/Images/%s", itemRatingKey, posterType), config.Global.MediaServer.URL)
	if Err.Message != "" {
		return Err
	}

	// Encode the image data as Base64
	base64ImageData := base64.StdEncoding.EncodeToString(imageData)

	// Make a POST request to the Emby/Jellyfin server
	headers := map[string]string{
		"Content-Type": "image/jpeg",
	}
	response, _, Err := utils.MakeHTTPRequest(baseURL.String(), "POST", headers, 60, []byte(base64ImageData), "MediaServer")
	if Err.Message != "" {
		return Err
	}
	defer response.Body.Close()

	return logging.StandardError{}
}
