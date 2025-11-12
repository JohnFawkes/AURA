package api

import (
	"aura/internal/logging"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

func (p *PlexServer) FetchMovieCollectionItemChildren(ctx context.Context, collectionItem *CollectionItem) logging.LogErrorInfo {
	return Plex_FetchMovieCollectionItemChildren(ctx, collectionItem)
}

func (e *EmbyJellyServer) FetchMovieCollectionItemChildren(ctx context.Context, collectionItem *CollectionItem) logging.LogErrorInfo {
	return EJ_FetchMovieCollectionItemChildren(ctx, collectionItem)
}

//////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////

func CallFetchCollectionChildren(ctx context.Context, collectionItem *CollectionItem) logging.LogErrorInfo {
	mediaServer, _, Err := NewMediaServerInterface(ctx, Config_MediaServer{})
	if Err.Message != "" {
		return Err
	}
	return mediaServer.FetchMovieCollectionItemChildren(ctx, collectionItem)
}

//////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////

func Plex_FetchMovieCollectionItemChildren(ctx context.Context, collectionItem *CollectionItem) logging.LogErrorInfo {
	ctx, logAction := logging.AddSubActionToContext(ctx,
		fmt.Sprintf("Getting all content for collection '%s' (ID: %s)", collectionItem.Title, collectionItem.RatingKey), logging.LevelDebug)
	defer logAction.Complete()

	// Construct the URL for the Plex server API request
	u, err := url.Parse(Global_Config.MediaServer.URL)
	if err != nil {
		logAction.SetError("Failed to parse URL", err.Error(), nil)
		return *logAction.Error
	}
	u.Path = path.Join(u.Path, "library", "collections", collectionItem.RatingKey, "children")
	query := u.Query()
	query.Set("includeGuids", "1")
	u.RawQuery = query.Encode()
	URL := u.String()

	// Make the Auth Headers for Request
	headers := MakeAuthHeader("X-Plex-Token", Global_Config.MediaServer.Token)

	// Make the HTTP request to Plex
	httpResp, respBody, logErr := MakeHTTPRequest(ctx, URL, http.MethodGet, headers, 60, nil, "Plex")
	if logErr.Message != "" {
		return logErr
	}
	defer httpResp.Body.Close()

	// Check the response status code
	if httpResp.StatusCode != 200 {
		logAction.SetError("Plex server returned non-200 status", fmt.Sprintf("Status Code: %d", httpResp.StatusCode), nil)
		return *logAction.Error
	}

	// Decode the response body
	var plexResponse PlexLibraryItemsWrapper
	logErr = DecodeJSONBody(ctx, respBody, &plexResponse, "PlexLibraryItemsWrapper")
	if logErr.Message != "" {
		return logErr
	}

	var items []MediaItem
	for _, item := range plexResponse.MediaContainer.Metadata {
		if item.Type != "movie" {
			continue
		}

		var itemInfo MediaItem
		itemInfo.RatingKey = item.RatingKey
		itemInfo.Type = item.Type
		itemInfo.Title = item.Title
		itemInfo.Year = item.Year
		itemInfo.LibraryTitle = item.LibrarySectionTitle
		itemInfo.UpdatedAt = item.UpdatedAt
		itemInfo.AddedAt = item.AddedAt
		itemInfo.Thumb = item.Thumb
		itemInfo.ContentRating = item.ContentRating
		itemInfo.Summary = item.Summary

		if t, err := time.Parse("2006-01-02", item.OriginallyAvailableAt); err == nil {
			itemInfo.ReleasedAt = t.Unix()
		} else {
			itemInfo.ReleasedAt = 0
		}

		itemInfo.Movie = &MediaItemMovie{
			File: MediaItemFile{
				Path:     item.Media[0].Part[0].File,
				Size:     item.Media[0].Part[0].Size,
				Duration: item.Media[0].Part[0].Duration,
			},
		}

		if len(item.Guid) > 0 {
			for _, guid := range item.Guids {
				if guid.ID != "" {
					// Sample guid.id : tmdb://######
					// Split into provider and id
					parts := strings.Split(guid.ID, "://")
					if len(parts) == 2 {
						provider := parts[0]
						id := parts[1]
						itemInfo.Guids = append(itemInfo.Guids, Guid{
							Provider: provider,
							ID:       id,
						})
						if provider == "tmdb" {
							itemInfo.TMDB_ID = id
						}
					}
				}
			}
		}

		existsInDB, posterSets, Err := DB_CheckIfMediaItemExists(ctx, itemInfo.TMDB_ID, itemInfo.LibraryTitle)
		if Err.Message != "" {
			logAction.Status = logging.LevelWarn
			logAction.AppendWarning("message", "Failed to check if media item exists in database")
			logAction.AppendWarning("error", Err)
		}
		if existsInDB {
			itemInfo.ExistInDatabase = true
			itemInfo.DBSavedSets = posterSets
		} else {
			itemInfo.ExistInDatabase = false
		}

		items = append(items, itemInfo)
	}

	collectionItem.MediaItems = items

	return logging.LogErrorInfo{}
}

//////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////

func EJ_FetchMovieCollectionItemChildren(ctx context.Context, collectionItem *CollectionItem) logging.LogErrorInfo {
	ctx, logAction := logging.AddSubActionToContext(ctx,
		fmt.Sprintf("Getting all content for collection '%s' (ID: %s)", collectionItem.Title, collectionItem.RatingKey), logging.LevelDebug)
	defer logAction.Complete()

	// Construct the URL for the Emby/Jellyfin server API request
	u, err := url.Parse(Global_Config.MediaServer.URL)
	if err != nil {
		logAction.SetError("Failed to parse Emby/Jellyfin base URL", err.Error(), nil)
		return *logAction.Error
	}
	u.Path = path.Join(u.Path, "Users", Global_Config.MediaServer.UserID, "Items")
	query := u.Query()
	query.Set("ParentId", collectionItem.RatingKey)
	query.Set("IncludeItemTypes", "Movie")
	query.Set("Fields", "DateLastContentAdded,PremiereDate,DateCreated,ProviderIds,BasicSyncInfo,CanDelete,CanDownload,PrimaryImageAspectRatio,ProductionYear,Status,EndDate,ProviderIds,Overview")
	u.RawQuery = query.Encode()
	URL := u.String()

	// Make the Auth Headers for Request
	headers := MakeAuthHeader("X-Emby-Token", Global_Config.MediaServer.Token)

	// Make the API request to Emby/Jellyfin
	httpResp, respBody, logErr := MakeHTTPRequest(ctx, URL, http.MethodGet, headers, 60, nil, Global_Config.MediaServer.Type)
	if logErr.Message != "" {
		return logErr
	}
	defer httpResp.Body.Close()

	// Parse the response body into an EmbyJellyLibraryItemsResponse struct
	var ejResponse EmbyJellyLibraryItemsResponse
	logErr = DecodeJSONBody(ctx, respBody, &ejResponse, "EmbyJellyLibraryItemsResponse")
	if logErr.Message != "" {
		return logErr
	}

	var items []MediaItem
	for _, item := range ejResponse.Items {
		if item.Type != "Movie" {
			continue
		}
		if item.ProviderIds.Tmdb == "" {
			continue
		}
		var itemInfo MediaItem
		itemInfo.RatingKey = item.ID
		itemInfo.Type = map[string]string{
			"Movie":  "movie",
			"Series": "show",
		}[item.Type]
		itemInfo.Title = item.Name
		itemInfo.Year = item.ProductionYear
		itemInfo.Thumb = item.ImageTags.Thumb
		itemInfo.LibraryTitle = collectionItem.LibraryTitle

		itemInfo.AddedAt = item.DateCreated.UnixMilli()
		itemInfo.ReleasedAt = item.PremiereDate.UnixMilli()

		existsInDB, posterSets, Err := DB_CheckIfMediaItemExists(ctx, itemInfo.TMDB_ID, itemInfo.LibraryTitle)
		if Err.Message != "" {
			logAction.Status = logging.LevelWarn
			logAction.AppendWarning("message", "Failed to check if media item exists in database")
			logAction.AppendWarning("error", Err)
		}
		if existsInDB {
			itemInfo.ExistInDatabase = true
			itemInfo.DBSavedSets = posterSets
		} else {
			itemInfo.ExistInDatabase = false
		}

		items = append(items, itemInfo)
		Global_Cache_LibraryStore.UpdateMediaItem(collectionItem.LibraryTitle, &itemInfo)
	}

	collectionItem.MediaItems = items

	return logging.LogErrorInfo{}
}
