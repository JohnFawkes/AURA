package plex

import (
	"aura/config"
	"aura/logging"
	"aura/models"
	"aura/utils"
	"aura/utils/httpx"
	"context"
	"fmt"
	"net/url"
	"path"
)

func GetAllImages(ctx context.Context, item *models.MediaItem, imageType string) (plexImages []PlexGetAllImagesMetadata, Err logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, fmt.Sprintf(
		"Plex: Fetching All '%s' Images for %s",
		imageType,
		utils.MediaItemInfo(*item),
	), logging.LevelDebug)
	defer logAction.Complete()

	plexImages = []PlexGetAllImagesMetadata{}
	Err = logging.LogErrorInfo{}

	if imageType == "backdrop" {
		imageType = "arts"
	} else {
		imageType = "posters"
	}

	// Construct the URL for the Plex API request
	u, err := url.Parse(config.Current.MediaServer.URL)
	if err != nil {
		logAction.SetError("Failed to parse base URL", "Ensure the URL is valid", map[string]any{"error": err.Error()})
		return plexImages, *logAction.Error
	}
	u.Path = path.Join(u.Path, "library", "metadata", item.RatingKey, imageType)
	URL := u.String()

	// Make the HTTP Request to Plex
	resp, respBody, Err := makeRequest(ctx, config.Current.MediaServer, URL, "GET", nil)
	if Err.Message != "" {
		return plexImages, *logAction.Error
	}
	defer resp.Body.Close()

	// Decode the Response
	var plexResp PlexGetAllImagesWrapper
	Err = httpx.DecodeResponseToJSON(ctx, respBody, &plexResp, "Plex Get All Images Response")
	if Err.Message != "" {
		return plexImages, *logAction.Error
	}

	plexImages = plexResp.MediaContainer.Metadata
	return plexImages, Err
}
