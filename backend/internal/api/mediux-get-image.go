package api

import (
	"aura/internal/logging"
	"context"
	"fmt"
	"net/http"
	"os"
	"path"
)

func Mediux_GetImage(ctx context.Context, assetID string, formatDate string, qualityParam string) ([]byte, string, logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, fmt.Sprintf("Get Image '%s' from Mediux", assetID), logging.LevelTrace)
	defer logAction.Complete()

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
			logAction.SetError("Failed to read image from temporary folder",
				fmt.Sprintf("Ensure the path %s is accessible.", imagePath),
				map[string]any{
					"error":   fmt.Sprintf("Error reading image: %v", err),
					"request": filePath,
				})
			return nil, "", *logAction.Error
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

		return imageData, imageType, logging.LogErrorInfo{}
	}

	// If the image does not exist in the temp folder, fetch it from Mediux

	// Construct the URL for the Mediux API request
	mediuxURL, Err := Mediux_GetImageURL(ctx, assetID, formatDate, qualityParam)
	if Err.Message != "" {
		return nil, "", Err
	}

	// Make the API request to Mediux
	httpResp, respBody, logErr := MakeHTTPRequest(ctx, mediuxURL, http.MethodGet, nil, 60, nil, "Mediux")
	if logErr.Message != "" {
		return nil, "", logErr
	}
	defer httpResp.Body.Close()

	// Check if the response body is empty
	if len(respBody) == 0 {
		logAction.SetError("Mediux returned an empty image response",
			"Ensure the asset ID is correct and the image exists on the Mediux server.",
			map[string]any{
				"URL": mediuxURL,
			})
		return nil, "", *logAction.Error
	}

	// Get the image type from the response headers
	imageType := httpResp.Header.Get("Content-Type")
	if imageType == "" {
		logAction.SetError("Failed to determine image type from Mediux response",
			"Ensure the Mediux server is returning a valid image.",
			map[string]any{
				"URL": mediuxURL,
			})
		return nil, "", *logAction.Error
	}

	// Handle Cache Images Setting
	go func() {
		if Global_Config.Images.CacheImages.Enabled {
			// Create a new logging data for this goroutine
			ctx, ld := logging.CreateLoggingContext(context.Background(), "Caching - Mediux Image")
			logAction := ld.AddAction("Caching Mediux Image", logging.LevelDebug)
			ctx = logging.WithCurrentAction(ctx, logAction)

			// Add the image to the temporary folder
			imagePath := path.Join(MediuxThumbsTempImageFolder, fileName)
			Err = Util_File_CheckFolderExists(ctx, MediuxThumbsTempImageFolder)
			if Err.Message != "" {
				logAction.SetError("Failed to create Mediux thumbs temp image folder",
					"Ensure the application has permissions to create the temp image folder.",
					map[string]any{
						"folder_path": MediuxThumbsTempImageFolder,
						"error":       Err.Message,
					})
				return
			}
			writeToFileAction := logAction.AddSubAction("Write Image to Temp Folder", logging.LevelTrace)
			err := os.WriteFile(imagePath, respBody, 0644)
			if err != nil {
				logAction.SetError("Failed to write image to Mediux thumbs temp image folder",
					"Ensure the application has write permissions to the temp image folder.",
					map[string]any{
						"image_path": imagePath,
						"error":      err.Error(),
					})
				return
			}
			logAction.AppendResult("image_path", imagePath)
			writeToFileAction.Complete()
			logAction.Complete()
			ld.Log()
		}
	}()

	// Return the image data and type
	return respBody, imageType, logging.LogErrorInfo{}
}
