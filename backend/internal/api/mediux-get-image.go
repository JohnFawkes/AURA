package api

import (
	"aura/internal/logging"
	"fmt"
	"os"
	"path"
)

func Mediux_GetImage(assetID string, formatDate string, qualityParam string) ([]byte, string, logging.StandardError) {
	Err := logging.NewStandardError()

	// Check if the temporary folder has the image
	fileName := fmt.Sprintf("%s_%s.jpg", assetID, formatDate)
	filePath := path.Join(MediuxThumbsTempImageFolder, fileName)
	exists := Util_File_CheckIfFileExists(filePath)
	if exists {
		// Serve the image from the temporary folder
		imagePath := path.Join(MediuxThumbsTempImageFolder, fileName)

		// Convert the image to bytes
		imageData, err := os.ReadFile(imagePath)
		if err != nil {
			Err.Message = "Failed to read image from temporary folder"
			Err.HelpText = fmt.Sprintf("Ensure the path %s is accessible.", imagePath)
			Err.Details = map[string]any{
				"error":   fmt.Sprintf("Error reading image: %v", err),
				"request": filePath,
			}
			return nil, "", Err
		}

		// Get the image type based on the file extension
		imageType := "image/jpeg" // Assuming JPEG for .jpg files
		switch path.Ext(imagePath) {
		case ".png":
			imageType = "image/png"
		case ".gif":
			imageType = "image/gif"
		case ".webp":
			imageType = "image/webp"
		}

		return imageData, imageType, logging.StandardError{}
	}

	// If the image does not exist in the temp folder, fetch it from Mediux

	// Construct the URL for the Mediux API request
	mediuxURL, Err := Mediux_GetImageURL(assetID, formatDate, qualityParam)
	if Err.Message != "" {
		return nil, "", Err
	}

	response, body, Err := MakeHTTPRequest(mediuxURL, "GET", nil, 60, nil, "Mediux")
	if Err.Message != "" {
		return nil, "", Err
	}
	defer response.Body.Close()

	// Check if the response body is empty
	if len(body) == 0 {
		Err.Message = "Empty response body from Mediux"
		Err.HelpText = "Ensure the asset ID is valid and the Mediux service is operational."
		Err.Details = map[string]any{
			"assetID":    assetID,
			"formatDate": formatDate,
		}
		return nil, "", Err
	}

	// Get the image type from the response headers
	imageType := response.Header.Get("Content-Type")
	if imageType == "" {
		Err.Message = "Missing Content-Type header in Mediux response"
		Err.HelpText = "Ensure the Mediux service is returning a valid image type."
		Err.Details = map[string]any{
			"assetID":    assetID,
			"formatDate": formatDate,
		}
		return nil, "", Err
	}

	// Handle Cache Images Setting
	go func() {
		if Global_Config.Images.CacheImages.Enabled {
			// Add the image to the temporary folder
			imagePath := path.Join(MediuxThumbsTempImageFolder, fileName)
			Err = Util_File_CheckFolderExists(MediuxThumbsTempImageFolder)
			if Err.Message != "" {
				logging.LOG.ErrorWithLog(Err)
			}
			err := os.WriteFile(imagePath, body, 0644)
			if err != nil {
				Err.Message = "Failed to write image to temporary folder"
				Err.HelpText = fmt.Sprintf("Ensure the path %s is accessible and writable.", MediuxThumbsTempImageFolder)
				Err.Details = map[string]any{
					"error":   fmt.Sprintf("Error writing image: %v", err),
					"request": imagePath,
				}
				logging.LOG.ErrorWithLog(Err)
			}
			logging.LOG.Debug(fmt.Sprintf("Cached image %s to temporary folder", fileName))
		}
	}()

	// Return the image data
	return body, imageType, logging.StandardError{}
}
