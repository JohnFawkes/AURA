package api

import (
	"aura/internal/logging"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"time"
)

var PlexTempImageFolder string
var EmbyJellyTempImageFolder string

func init() {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "/config"
	}
	EmbyJellyTempImageFolder = path.Join(configPath, "temp-images", "emby-jellyfin")
	PlexTempImageFolder = path.Join(configPath, "temp-images", "plex")
}

///////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////

func (p *PlexServer) FetchImageFromMediaServer(ctx context.Context, ratingKey string, imageType string) ([]byte, logging.LogErrorInfo) {
	return Plex_FetchImageFromMediaServer(ctx, ratingKey, imageType)
}

func (e *EmbyJellyServer) FetchImageFromMediaServer(ctx context.Context, ratingKey string, imageType string) ([]byte, logging.LogErrorInfo) {
	return EJ_FetchImageFromMediaServer(ctx, ratingKey, imageType)
}

///////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////

func CallFetchImageFromMediaServer(ctx context.Context, ratingKey string, imageType string) ([]byte, logging.LogErrorInfo) {
	mediaServer, _, Err := NewMediaServerInterface(ctx, Config_MediaServer{})
	if Err.Message != "" {
		return nil, Err
	}

	return mediaServer.FetchImageFromMediaServer(ctx, ratingKey, imageType)
}

///////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////

func Plex_FetchImageFromMediaServer(ctx context.Context, ratingKey string, imageType string) ([]byte, logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Fetching Image from Plex Media Server", logging.LevelDebug)
	defer logAction.Complete()

	// Construct the URL for the Plex server API request
	if imageType == "backdrop" {
		imageType = "art"
	}

	// Make the URL
	u, err := url.Parse(Global_Config.MediaServer.URL)
	if err != nil {
		logAction.SetError("Failed to parse Plex base URL", err.Error(), nil)
		return nil, *logAction.Error
	}
	u.Path = path.Join(u.Path, "library", "metadata", ratingKey, imageType, fmt.Sprintf("%d", time.Now().Unix()))
	URL := u.String()

	// Make the Auth Headers for Request
	headers := MakeAuthHeader("X-Plex-Token", Global_Config.MediaServer.Token)

	// Make the API request to Plex
	httpResp, respBody, logErr := MakeHTTPRequest(ctx, URL, http.MethodGet, headers, 60, nil, "Plex")
	if logErr.Message != "" {
		return nil, logErr
	}
	defer httpResp.Body.Close()

	// Check the response status code
	if httpResp.StatusCode != http.StatusOK {
		logAction.SetError("Failed to fetch image from Plex",
			fmt.Sprintf("Plex server returned status code %d", httpResp.StatusCode),
			map[string]any{
				"URL":        URL,
				"StatusCode": httpResp.StatusCode,
			})
		return nil, *logAction.Error
	}

	// Check if the response body is empty
	if len(respBody) == 0 {
		logAction.SetError("No image data returned from Plex",
			"The Plex server returned an empty response body for the requested image.",
			map[string]any{
				"URL": URL,
			})
		return nil, *logAction.Error
	}

	// Return the image data
	return respBody, logging.LogErrorInfo{}
}

func EJ_FetchImageFromMediaServer(ctx context.Context, ratingKey string, imageType string) ([]byte, logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, fmt.Sprintf("Fetching Image from %s Media Server", Global_Config.MediaServer.Type), logging.LevelDebug)
	defer logAction.Complete()

	switch imageType {
	case "poster":
		imageType = "Primary"
	case "backdrop":
		imageType = "Backdrop"
	default:
		logAction.SetError("Unsupported image type requested",
			fmt.Sprintf("The image type '%s' is not supported. Use 'poster' or 'backdrop'.", imageType),
			nil)
		return nil, *logAction.Error
	}

	// Make the URL
	u, err := url.Parse(Global_Config.MediaServer.URL)
	if err != nil {
		logAction.SetError("Failed to parse Emby/Jellyfin base URL", err.Error(), nil)
		return nil, *logAction.Error
	}
	u.Path = path.Join("Items", ratingKey, "Images", imageType)
	URL := u.String()

	// Make the Auth Headers for Request
	headers := MakeAuthHeader("X-Emby-Token", Global_Config.MediaServer.Token)

	// Make the API request to Emby/Jellyfin
	httpResp, respBody, logErr := MakeHTTPRequest(ctx, URL, http.MethodGet, headers, 60, nil, Global_Config.MediaServer.Type)
	if logErr.Message != "" {
		return nil, logErr
	}
	defer httpResp.Body.Close()

	// Check if the response body is empty
	if len(respBody) == 0 {
		logAction.SetError("No image data returned from Media Server",
			"The Media Server returned an empty response body for the requested image.",
			map[string]any{
				"URL": URL,
			})
		return nil, *logAction.Error
	}

	// Return the image data
	return respBody, logging.LogErrorInfo{}
}
