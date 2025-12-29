package api

import (
	"aura/internal/logging"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"time"
)

func (*PlexServer) RefreshMediaItem(ctx context.Context, ratingKey string) logging.LogErrorInfo {
	return Plex_RefreshItem(ctx, ratingKey)
}

func (e *EmbyJellyServer) RefreshMediaItem(ctx context.Context, ratingKey string) logging.LogErrorInfo {
	return EmbyJelly_RefreshItem(ctx, ratingKey)
}

//////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////

func CallRefreshMediaItem(ctx context.Context, ratingKey string) logging.LogErrorInfo {
	mediaServer, _, Err := NewMediaServerInterface(ctx, Config_MediaServer{})
	if Err.Message != "" {
		return Err
	}

	return mediaServer.RefreshMediaItem(ctx, ratingKey)
}

//////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////

// Plex_RefreshItem sends a request to the Plex server to refresh the specified item.
func Plex_RefreshItem(ctx context.Context, itemRatingKey string) logging.LogErrorInfo {
	ctx, logAction := logging.AddSubActionToContext(ctx, fmt.Sprintf("Refreshing Item with Rating Key '%s' on Plex", itemRatingKey), logging.LevelDebug)
	defer logAction.Complete()

	// Construct the URL for the Plex server API request
	u, err := url.Parse(Global_Config.MediaServer.URL)
	if err != nil {
		logAction.SetError("Failed to parse base URL", "Ensure the URL is valid", map[string]any{"error": err.Error()})
		return *logAction.Error
	}
	u.Path = path.Join(u.Path, "library", "metadata", itemRatingKey, "refresh")
	URL := u.String()

	// Make the Auth Headers for Request
	headers := MakeAuthHeader("X-Plex-Token", Global_Config.MediaServer.Token)

	// Make the API request to Plex
	httpResp, _, logErr := MakeHTTPRequest(ctx, URL, http.MethodPut, headers, 60, nil, "Plex")
	if logErr.Message != "" {
		return logErr
	}
	defer httpResp.Body.Close()

	// Check the response status code
	if httpResp.StatusCode != http.StatusOK && httpResp.StatusCode != http.StatusAccepted {
		logAction.SetError("Failed to refresh item on Plex",
			fmt.Sprintf("Plex server returned status code %d", httpResp.StatusCode),
			map[string]any{
				"URL":        URL,
				"StatusCode": httpResp.StatusCode,
			})
		return *logAction.Error
	}

	currentImages, logErr := Plex_GetAllImages(ctx, itemRatingKey, "poster")
	if logErr.Message != "" {
		return logErr
	}

	if len(currentImages) > 0 {
		hasSelectedImage := false
		for _, img := range currentImages {
			if img.Selected {
				hasSelectedImage = true
				break
			}
		}
		if !hasSelectedImage {
			hasLocal := false
			// If there is no selected image, select a local one if it exists, else select the first one
			for _, img := range currentImages {
				if img.Provider == "local" {
					logErr = Plex_SetPoster(ctx, itemRatingKey, img.RatingKey, "poster")
					if logErr.Message != "" {
						return logErr
					}
					hasLocal = true
					break
				}
			}
			if !hasLocal {
				logErr = Plex_SetPoster(ctx, itemRatingKey, currentImages[0].RatingKey, "poster")
				if logErr.Message != "" {
					return logErr
				}
			}
		}
	}
	time.Sleep(500 * time.Millisecond) // Give Plex a moment to process the refresh before any further actions
	return logging.LogErrorInfo{}
}

//////////////////////////////////////////////////////////////////////

func EmbyJelly_RefreshItem(ctx context.Context, itemRatingKey string) logging.LogErrorInfo {
	ctx, logAction := logging.AddSubActionToContext(ctx, fmt.Sprintf("Refresh Media Item (%s) in %s", itemRatingKey, Global_Config.MediaServer.Type), logging.LevelDebug)
	defer logAction.Complete()

	// Construct the URL for the Emby/Jellyfin server API request
	u, err := url.Parse(Global_Config.MediaServer.URL)
	if err != nil {
		logAction.SetError("Failed to parse URL", err.Error(), nil)
		return *logAction.Error
	}
	//Recursive=true&ImageRefreshMode=Default&MetadataRefreshMode=Default&ReplaceAllImages=false&RegenerateTrickplay=false&ReplaceAllMetadata=false
	u.Path = path.Join(u.Path, "Items", itemRatingKey, "Refresh")
	query := u.Query()
	query.Add("Recursive", "true")
	query.Add("ImageRefreshMode", "Default")
	query.Add("MetadataRefreshMode", "Default")
	query.Add("ReplaceAllImages", "false")
	query.Add("RegenerateTrickplay", "false")
	query.Add("ReplaceAllMetadata", "false")
	u.RawQuery = query.Encode()
	URL := u.String()

	// Make the Auth Headers for Request
	headers := MakeAuthHeader("X-Emby-Token", Global_Config.MediaServer.Token)

	// Make the API request to Emby/Jellyfin
	httpResp, _, logErr := MakeHTTPRequest(ctx, URL, http.MethodPost, headers, 60, nil, Global_Config.MediaServer.Type)
	if logErr.Message != "" {
		return logErr
	}
	defer httpResp.Body.Close()

	// Check the response status code
	if httpResp.StatusCode != http.StatusOK && httpResp.StatusCode != http.StatusAccepted && httpResp.StatusCode != http.StatusNoContent {
		logAction.SetError("Failed to refresh item on Emby/Jellyfin",
			fmt.Sprintf("%s server returned status code %d", Global_Config.MediaServer.Type, httpResp.StatusCode),
			map[string]any{
				"URL":        URL,
				"StatusCode": httpResp.StatusCode,
			})
		return *logAction.Error
	}

	return logging.LogErrorInfo{}
}
