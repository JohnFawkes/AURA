package api

import (
	"aura/internal/logging"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
)

type MediuxImageQuality string

const (
	MediuxImageQualityOriginal  MediuxImageQuality = "original"
	MediuxImageQualityOptimized MediuxImageQuality = "optimized"
	MediuxImageQualityThumb     MediuxImageQuality = "thumb"
)

func Mediux_GetImage(ctx context.Context, assetID string, formatDate string, imageQuality MediuxImageQuality) (imageData []byte, imageType string, Err logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, fmt.Sprintf("Get Image '%s' from MediUX (%s)", assetID, imageQuality), logging.LevelTrace)
	defer logAction.Complete()

	imageData = nil
	imageType = ""
	Err = logging.LogErrorInfo{}

	// Make the file name based on assetID and formatDate
	fileName := fmt.Sprintf("%s_%s.jpg", assetID, formatDate)

	// Determine the folder path based on qualityParam
	var folderPath string
	switch imageQuality {
	case MediuxImageQualityOriginal:
		folderPath = MediuxFullTempImageFolder
	case MediuxImageQualityOptimized, MediuxImageQualityThumb:
		folderPath = MediuxThumbsTempImageFolder
	}

	isNewDownload := false

	// Full file path
	filePath := path.Join(folderPath, fileName)

	// Run this after the function completes
	// Handle Cache Images Setting
	defer func() {
		// If there is an error or no image data, do not attempt to cache
		if Err.Message != "" || len(imageData) == 0 {
			return
		}

		if !isNewDownload {
			// If the image was not newly downloaded, no need to cache
			return
		}

		// Check if caching is enabled
		if Global_Config.Images.CacheImages.Enabled {
			// Create a new logging data for this goroutine
			ctx, ld := logging.CreateLoggingContext(context.Background(), "Caching - MediUX Image")
			logAction := ld.AddAction("Caching MediUX Image", logging.LevelDebug)
			ctx = logging.WithCurrentAction(ctx, logAction)
			// Check if the folder exists
			Err = Util_File_CheckFolderExists(ctx, folderPath)
			if Err.Message != "" {
				return
			}

			writeToFileAction := logAction.AddSubAction("Write Image to Temp Folder", logging.LevelTrace)
			err := os.WriteFile(filePath, imageData, 0644)
			if err != nil {
				logAction.SetError("Failed to write image to MediUX thumbs temp image folder",
					"Ensure the application has write permissions to the temp image folder.",
					map[string]any{
						"filePath": filePath,
						"error":    err.Error(),
					})

				return
			}
			logAction.AppendResult("filePath", filePath)
			writeToFileAction.Complete()
			logAction.Complete()
			ld.Log()

		}
	}()

	// Now we check to see if Global_Config.Images.CacheImages.Enabled is true
	// If it is, we check the temporary or full image folder for the image based on qualityParam
	// If the image exists there, we serve it from disk
	// If not, we fetch it from MediUX and save it to the temp/full folder based on qualityParam
	// If Global_Config.Images.CacheImages.Enabled is false, we always fetch from MediUX

	if Global_Config.Images.CacheImages.Enabled {
		// Check if the folder exists
		Err = Util_File_CheckFolderExists(ctx, folderPath)
		if Err.Message != "" {
			return imageData, imageType, Err
		}

		// Check if the file exists
		exists := Util_File_CheckIfFileExists(filePath)
		if exists { // If it exists, read and return it
			// Read the image from disk
			imageData, err := os.ReadFile(filePath)
			if err != nil {
				logAction.SetError("Failed to read cached image from disk",
					"Ensure the application has read permissions for the cached image file.",
					map[string]any{
						"file_path": filePath,
						"error":     err.Error(),
					})
				Err = *logAction.Error
				return imageData, imageType, *logAction.Error
			}

			// Get the image type based on the file extension
			imageType = "image/jpeg" // Assuming JPEG for .jpg files
			switch path.Ext(filePath) {
			case ".png":
				imageType = "image/png"
			case ".gif":
				imageType = "image/gif"
			case ".webp":
				imageType = "image/webp"
			}
			logAction.AppendResult("filePath", filePath)
			logAction.AppendResult("size", len(imageData))
			logAction.AppendResult("imageType", imageType)
			logAction.AppendResult("source", "Cache")
			return imageData, imageType, Err
		}
	}

	// If the image does not exist in the cache, fetch it from MediUX

	// Construct the URL for the MediUX API request
	mediuxURL, Err := Mediux_GetImageURL(ctx, assetID, formatDate, imageQuality)
	if Err.Message != "" {
		return imageData, imageType, Err
	}

	var headers map[string]string
	if strings.HasPrefix(mediuxURL, "https://images.mediux.io") {
		// Make the Auth Headers for Request
		headers = MakeAuthHeader("Authorization", Global_Config.Mediux.Token)
	}

	// Make the API request to MediUX
	httpResp, respBody, logErr := MakeHTTPRequest(ctx, mediuxURL, http.MethodGet, headers, 60, nil, "MediUX")
	if logErr.Message != "" {
		return imageData, imageType, Err
	}
	defer httpResp.Body.Close()

	// Check for HTTP errors
	if httpResp.StatusCode != http.StatusOK {
		logAction.SetError("MediUX returned non-200 status code",
			"Ensure the asset ID is correct and the MediUX service is operational.",
			map[string]any{
				"URL":        mediuxURL,
				"statusCode": httpResp.StatusCode,
			})
		Err = *logAction.Error
		return nil, "", *logAction.Error
	}

	// Check if the response body is empty
	if len(respBody) == 0 {
		logAction.SetError("MediUX returned an empty image response",
			"Ensure the asset ID is correct and the image exists on the MediUX server.",
			map[string]any{
				"URL": mediuxURL,
			})
		Err = *logAction.Error
		return nil, "", *logAction.Error
	}
	imageData = respBody

	// Get the image type from the response headers
	imageType = httpResp.Header.Get("Content-Type")
	if imageType == "" {
		logAction.SetError("Failed to determine image type from MediUX response",
			"Ensure the MediUX server is returning a valid image.",
			map[string]any{
				"URL": mediuxURL,
			})
		Err = *logAction.Error
		return nil, "", *logAction.Error
	}

	isNewDownload = true
	logAction.AppendResult("size", len(imageData))
	logAction.AppendResult("imageType", imageType)
	logAction.AppendResult("source", "MediUX")

	// Return the image data and type
	return imageData, imageType, Err
}

func Mediux_GetAvatarImage(ctx context.Context, avatarID string) (imageData []byte, imageType string, Err logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, fmt.Sprintf("Get Avatar Image '%s' from MediUX", avatarID), logging.LevelTrace)
	defer logAction.Complete()

	imageData = nil
	imageType = ""
	Err = logging.LogErrorInfo{}

	// Construct the MediUX URL
	u, err := url.Parse(MediuxBaseURL)
	if err != nil {
		logAction.SetError("Failed to parse MediUX base URL", err.Error(), nil)
		return imageData, imageType, *logAction.Error
	}
	u.Path = path.Join(u.Path, "assets", avatarID)

	// Make the Auth Headers for Request
	headers := MakeAuthHeader("Authorization", Global_Config.Mediux.Token)

	// Make the API request to MediUX
	httpResp, respBody, logErr := MakeHTTPRequest(ctx, u.String(), http.MethodGet, headers, 60, nil, "MediUX")
	if logErr.Message != "" {
		return imageData, imageType, Err
	}
	defer httpResp.Body.Close()

	// Check if the response body is empty
	if len(respBody) == 0 {
		logAction.SetError("MediUX returned an empty avatar image response",
			"Ensure the avatar ID is correct and the image exists on the MediUX server.",
			map[string]any{
				"avatarID": avatarID,
			})
		Err = *logAction.Error
		return nil, "", *logAction.Error
	}
	imageData = respBody

	// Get the image type from the response headers
	imageType = httpResp.Header.Get("Content-Type")
	if imageType == "" {
		logAction.SetError("Failed to determine avatar image type from MediUX response",
			"Ensure the MediUX server is returning a valid image.",
			map[string]any{
				"avatarID": avatarID,
			})
		Err = *logAction.Error
		return nil, "", *logAction.Error
	}

	logAction.AppendResult("size", len(imageData))
	logAction.AppendResult("imageType", imageType)

	// Return the image data and type
	return imageData, imageType, Err
}
