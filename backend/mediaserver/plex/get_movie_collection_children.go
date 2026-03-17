package plex

import (
	"aura/cache"
	"aura/config"
	"aura/logging"
	"aura/models"
	"aura/utils/httpx"
	"context"
	"fmt"
	"net/url"
	"path"
)

func (p *Plex) GetMovieCollectionChildrenItems(ctx context.Context, collection *models.CollectionItem) (Err logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, fmt.Sprintf(
		"Plex: Fetching Collection Children for '%s' | %s [%s | %s]",
		collection.Title, collection.LibraryTitle, collection.TMDB_ID, collection.RatingKey,
	), logging.LevelDebug)
	defer logAction.Complete()

	Err = logging.LogErrorInfo{}

	// Construct the URL for the Plex API request
	// Construct the URL for the Plex API request
	u, err := url.Parse(config.Current.MediaServer.URL)
	if err != nil {
		logAction.SetError(logging.Error_BaseUrlParsing(err))
		return *logAction.Error
	}
	u.Path = path.Join(u.Path, "library", "collections", collection.RatingKey, "children")
	query := u.Query()
	query.Set("includeGuids", "1")
	u.RawQuery = query.Encode()
	URL := u.String()

	// Make the HTTP Request to Plex
	resp, respBody, Err := makeRequest(ctx, config.Current.MediaServer, URL, "GET", nil)
	if Err.Message != "" {
		logAction.SetErrorFromInfo(Err)
		return *logAction.Error
	}
	defer resp.Body.Close()

	// Decode the Response
	var plexResp PlexLibraryItemsWrapper
	Err = httpx.DecodeResponseToJSON(ctx, respBody, &plexResp, "Plex Collection Children Response")
	if Err.Message != "" {
		return *logAction.Error
	}

	var items []models.MediaItem
	for _, item := range plexResp.MediaContainer.Metadata {
		if item.Type != "movie" {
			continue
		}

		isInSelectedLibraryTitles := false
		for _, library := range config.Current.MediaServer.Libraries {
			if library.Title == item.LibrarySectionTitle {
				isInSelectedLibraryTitles = true
				break
			}
		}
		if !isInSelectedLibraryTitles {
			continue
		}

		itemInfo, Err := extractMediaItemFromResponse(ctx, item)
		if Err.Message != "" {
			continue
		}

		// Update the collection cache with the item
		cache.CollectionsStore.UpdateMediaItemInCollectionByIndex(collection.Index, itemInfo)

		items = append(items, *itemInfo)
	}

	collection.MediaItems = items
	return logging.LogErrorInfo{}
}
