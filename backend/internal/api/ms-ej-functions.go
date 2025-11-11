package api

import (
	"aura/internal/logging"
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
)

type EmbyJellyItemImagesResponse struct {
	ImageType  string `json:"ImageType"`
	ImageIndex int    `json:"ImageIndex,omitempty"`
	ImageTag   string `json:"ImageTag"`
}

func EJ_UploadImage(ctx context.Context, title string, ratingKey string, file PosterFile, imageData []byte) logging.LogErrorInfo {
	var posterType string
	if file.Type == "backdrop" {
		posterType = "Backdrop"
	} else {
		posterType = "Primary"
	}

	ctx, logAction := logging.AddSubActionToContext(ctx, fmt.Sprintf("Uploading %s Image for %s to %s", posterType, title, Global_Config.MediaServer.Type), logging.LevelInfo)
	defer logAction.Complete()

	// Make the URL
	u, err := url.Parse(Global_Config.MediaServer.URL)
	if err != nil {
		logAction.SetError("Failed to parse Media Server URL",
			"Ensure the Media Server URL is valid",
			map[string]any{
				"error": err.Error(),
			})
		return *logAction.Error
	}
	u.Path = path.Join(u.Path, "Items", ratingKey, "Images", posterType)
	URL := u.String()

	// Make the Auth Headers for Request
	headers := MakeAuthHeader("X-Emby-Token", Global_Config.MediaServer.Token)

	// Encode the image data as Base64
	base64ImageData := base64.StdEncoding.EncodeToString(imageData)

	// Add Content-Type Header
	headers["Content-Type"] = "image/jpeg"

	// Make a POST request to upload the image
	httpResp, respBody, logErr := MakeHTTPRequest(ctx, URL, http.MethodPost, headers, 60, []byte(base64ImageData), Global_Config.MediaServer.Type)
	if logErr.Message != "" {
		logAction.SetError("Failed to upload image to Emby/Jellyfin",
			"Ensure the Emby/Jellyfin server is reachable and the API key is valid",
			map[string]any{
				"error":        logErr.Detail,
				"responseBody": respBody,
			})
		return *logAction.Error
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != 200 && httpResp.StatusCode != 204 {
		logAction.SetError("Failed to upload image to Emby/Jellyfin",
			"Ensure the Emby/Jellyfin server is reachable and the API key is valid",
			map[string]any{
				"error":        logErr.Detail,
				"responseBody": respBody,
			})
		return *logAction.Error
	}

	logAction.AppendResult("status_code", httpResp.StatusCode)
	logAction.AppendResult("message", fmt.Sprintf("%s uploaded to %s for item '%s'", strings.ToTitle(file.Type), Global_Config.MediaServer.Type, title))
	logAction.Complete()
	return logging.LogErrorInfo{}
}

func EJ_ChangeImageIndex(ctx context.Context, title string, ratingKey string, newImageitem EmbyJellyItemImagesResponse) logging.LogErrorInfo {
	ctx, logAction := logging.AddSubActionToContext(ctx, fmt.Sprintf("Changing Image Index for %s in %s", title, Global_Config.MediaServer.Type), logging.LevelInfo)
	defer logAction.Complete()

	// Make the URL
	u, err := url.Parse(Global_Config.MediaServer.URL)
	if err != nil {
		logAction.SetError("Failed to parse Media Server URL",
			"Ensure the Media Server URL is valid",
			map[string]any{
				"error": err.Error(),
			})
		return *logAction.Error
	}
	u.Path = path.Join(u.Path, "Items", ratingKey, "Images", "Backdrop",
		fmt.Sprintf("%d", newImageitem.ImageIndex), "Index")
	query := u.Query()
	query.Set("newIndex", "0")
	u.RawQuery = query.Encode()
	URL := u.String()

	// Make the Auth Headers for Request
	headers := MakeAuthHeader("X-Emby-Token", Global_Config.MediaServer.Token)

	// Make a POST request to set the new index
	httpResp, respBody, logErr := MakeHTTPRequest(ctx, URL, http.MethodPost, headers, 30, nil, Global_Config.MediaServer.Type)
	if logErr.Message != "" {
		logAction.SetError("Failed to set new image index in Emby/Jellyfin",
			"Ensure the Emby/Jellyfin server is reachable and the API key is valid",
			map[string]any{
				"error":        logErr.Detail,
				"responseBody": respBody,
			})
		return *logAction.Error
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != 200 && httpResp.StatusCode != 204 {
		logAction.SetError("Failed to set new image index in Emby/Jellyfin",
			"Ensure the Emby/Jellyfin server is reachable and the API key is valid",
			map[string]any{
				"error":        logErr.Detail,
				"responseBody": respBody,
			})
		return *logAction.Error
	}

	logAction.AppendResult("status_code", httpResp.StatusCode)
	logAction.AppendResult("message", fmt.Sprintf("Image index changed to 0 for item '%s'", title))
	logAction.Complete()
	return logging.LogErrorInfo{}
}

func EJ_GetCurrentImages(ctx context.Context, title string, ratingKey, status string) ([]EmbyJellyItemImagesResponse, logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, fmt.Sprintf("Fetching %s Images for %s from %s", status, title, Global_Config.MediaServer.Type), logging.LevelInfo)
	defer logAction.Complete()

	// Make the URL
	u, err := url.Parse(Global_Config.MediaServer.URL)
	if err != nil {
		logAction.SetError("Failed to parse Media Server URL",
			"Ensure the Media Server URL is valid",
			map[string]any{
				"error": err.Error(),
			})
		return nil, *logAction.Error
	}
	u.Path = path.Join(u.Path, "Items", ratingKey, "Images")
	URL := u.String()

	// Make the Auth Headers for Request
	headers := MakeAuthHeader("X-Emby-Token", Global_Config.MediaServer.Token)

	// Make a GET request to retrieve current images
	httpResp, respBody, logErr := MakeHTTPRequest(ctx, URL, http.MethodGet, headers, 30, nil, Global_Config.MediaServer.Type)
	if logErr.Message != "" {
		logAction.SetError("Failed to retrieve current images from Emby/Jellyfin",
			"Ensure the Emby/Jellyfin server is reachable and the API key is valid",
			map[string]any{
				"error":        logErr.Detail,
				"responseBody": respBody,
			})
		return nil, *logAction.Error
	}
	defer httpResp.Body.Close()

	var currentImages []EmbyJellyItemImagesResponse

	logErr = DecodeJSONBody(ctx, respBody, &currentImages, "EmbyJellyItemImagesResponse")
	if logErr.Message != "" {
		logAction.SetError("Failed to parse current images response from Emby/Jellyfin",
			"Ensure the Emby/Jellyfin server is reachable and the API key is valid",
			map[string]any{
				"error":        logErr.Detail,
				"responseBody": respBody,
			})
		return nil, *logAction.Error
	}

	return currentImages, logging.LogErrorInfo{}
}

func EJ_FindNewImage(currentImages, newImages []EmbyJellyItemImagesResponse, posterType string) EmbyJellyItemImagesResponse {
	for _, updatedImage := range newImages {
		if updatedImage.ImageType == "" || updatedImage.ImageType != "Backdrop" {
			continue
		}
		found := false
		for _, currentImage := range currentImages {
			if updatedImage.ImageType == currentImage.ImageType && updatedImage.ImageIndex == currentImage.ImageIndex && updatedImage.ImageTag == currentImage.ImageTag {
				found = true
				break
			}
		}
		if !found && updatedImage.ImageType == posterType {
			return updatedImage
		}
	}
	return EmbyJellyItemImagesResponse{}
}
