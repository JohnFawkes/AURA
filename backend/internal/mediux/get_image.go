package mediux

import (
	"aura/internal/config"
	"aura/internal/logging"
	"aura/internal/utils"
	"fmt"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/go-chi/chi/v5"
)

var MediuxThumbsTempImageFolder string
var MediuxFullTempImageFolder string

func init() {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "/config"
	}
	MediuxThumbsTempImageFolder = path.Join(configPath, "temp-images", "mediux", "thumbs")
	MediuxFullTempImageFolder = path.Join(configPath, "temp-images", "mediux", "full")
}

func GetMediuxImage(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	logging.LOG.Trace(r.URL.Path)
	Err := logging.NewStandardError()

	// Get the asset ID from the URL
	assetID := chi.URLParam(r, "assetID")
	if assetID == "" {
		Err.Message = "Missing asset ID in URL"
		Err.HelpText = "Ensure the asset ID is provided in the URL path."
		Err.Details = map[string]any{
			"error":   "Asset ID is empty",
			"request": r.URL.Path,
		}
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}

	// Get the modified date from the URL query parameters
	modifiedDate := r.URL.Query().Get("modifiedDate")
	var modifiedDateTime time.Time
	var err error
	if modifiedDate == "" || modifiedDate == "0" || modifiedDate == "undefined" {
		// Use today's date if the modified date is not provided
		modifiedDateTime = time.Now()
	} else {
		// Try to parse the modified date as an ISO 8601 timestamp
		modifiedDateTime, err = time.Parse(time.RFC3339, modifiedDate)
		if err != nil {
			Err.Message = "Invalid modified date format"
			Err.HelpText = "Ensure the modified date is in ISO 8601 format (e.g., 2023-10-01T12:00:00Z)."
			Err.Details = map[string]any{
				"modifiedDate": modifiedDate,
			}
			utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
			return
		}
	}
	// Format the date to be YYYYMMDDHHMMSS
	// Example: 2025-06-20T10:20:30Z -> 20250620102030
	formatDate := modifiedDateTime.Format("20060102150405")

	// Get Quality from the URL query parameters
	qualityParam := r.URL.Query().Get("quality")
	if qualityParam == "" {
		// Default to "thumb" if quality is not provided
		qualityParam = "thumb"
	}
	// Check if the quality is valid
	if qualityParam != "thumb" && qualityParam != "original" && qualityParam != "optimized" {
		Err.Message = "Invalid quality parameter"
		Err.HelpText = "Ensure the quality parameter is either 'thumb', 'original', or 'optimized'."
		Err.Details = map[string]any{
			"quality": qualityParam,
		}
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}

	// Check if the temporary folder has the image
	fileName := fmt.Sprintf("%s_%s.jpg", assetID, formatDate)
	filePath := path.Join(MediuxThumbsTempImageFolder, fileName)
	exists := utils.CheckIfImageExists(filePath)
	if exists {
		// Serve the image from the temporary folder
		imagePath := path.Join(MediuxThumbsTempImageFolder, fileName)
		http.ServeFile(w, r, imagePath)
		return
	}

	// If the image does not exist, then get it from Mediux
	imageData, imageType, Err := FetchImage(assetID, formatDate, qualityParam)
	if Err.Message != "" {
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}

	if config.Global.Images.CacheImages.Enabled {
		// Add the image to the temporary folder
		imagePath := path.Join(MediuxThumbsTempImageFolder, fileName)
		Err = utils.CheckFolderExists(MediuxThumbsTempImageFolder)
		if Err.Message != "" {
			utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
			return
		}
		err := os.WriteFile(imagePath, imageData, 0644)
		if err != nil {
			Err.Message = "Failed to write image to temporary folder"
			Err.HelpText = fmt.Sprintf("Ensure the path %s is accessible and writable.", MediuxThumbsTempImageFolder)
			Err.Details = map[string]any{
				"error":   fmt.Sprintf("Error writing image: %v", err),
				"request": r.URL.Path,
			}
			utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
			return
		}
	}

	w.Header().Set("Content-Type", imageType)
	w.WriteHeader(http.StatusOK)
	w.Write(imageData)
}

func FetchImage(assetID string, formatDate string, qualityParam string) ([]byte, string, logging.StandardError) {
	Err := logging.NewStandardError()

	// Construct the URL for the Mediux API request
	mediuxURL, Err := GetMediuxImageURL(assetID, formatDate, qualityParam)
	if Err.Message != "" {
		return nil, "", Err
	}

	response, body, Err := utils.MakeHTTPRequest(mediuxURL, "GET", nil, 60, nil, "Mediux")
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

	// Return the image data
	return body, imageType, logging.StandardError{}
}
