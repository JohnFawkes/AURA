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

func (p *PlexServer) FetchLibrarySectionItems(ctx context.Context, section LibrarySection, sectionStartIndex string) ([]MediaItem, int, logging.LogErrorInfo) {
	return Plex_FetchLibrarySectionItems(ctx, section, sectionStartIndex, "500")
}

func (e *EmbyJellyServer) FetchLibrarySectionItems(ctx context.Context, section LibrarySection, sectionStartIndex string) ([]MediaItem, int, logging.LogErrorInfo) {
	return EJ_FetchLibrarySectionItems(ctx, section, sectionStartIndex, "500")
}

//////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////

func CallFetchLibrarySectionItems(ctx context.Context, sectionID, sectionTitle, sectionType string, sectionStartIndex string) (LibrarySection, logging.LogErrorInfo) {
	var section LibrarySection
	section.ID = sectionID
	section.Title = sectionTitle
	section.Type = sectionType

	mediaServer, _, Err := NewMediaServerInterface(ctx, Config_MediaServer{})
	if Err.Message != "" {
		return section, Err
	}

	// Fetch the section items from the media server
	mediaItems, totalSize, Err := mediaServer.FetchLibrarySectionItems(ctx, section, sectionStartIndex)
	if Err.Message != "" {
		return section, Err
	}

	section.MediaItems = mediaItems
	section.TotalSize = totalSize
	Global_Cache_LibraryStore.UpdateSection(&section)
	return section, logging.LogErrorInfo{}
}

//////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////

func Plex_FetchLibrarySectionItems(ctx context.Context, section LibrarySection, sectionStartIndex string, limit string) ([]MediaItem, int, logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx,
		fmt.Sprintf("Getting all content for '%s' (ID: %s | Start Index: %s) from Plex",
			section.Title, section.ID, sectionStartIndex), logging.LevelDebug)
	defer logAction.Complete()

	// If limit is not provided, set it to 500
	if limit == "" {
		limit = "500" // Default limit if not provided
	}

	// Construct the URL for the Plex server API request
	u, err := url.Parse(Global_Config.MediaServer.URL)
	if err != nil {
		logAction.SetError("Failed to parse URL", err.Error(), nil)
		return nil, 0, *logAction.Error
	}
	u.Path = path.Join(u.Path, "library", "sections", section.ID, "all")
	query := u.Query()
	query.Set("X-Plex-Container-Start", sectionStartIndex)
	query.Set("X-Plex-Container-Size", limit)
	query.Set("includeGuids", "1")
	u.RawQuery = query.Encode()
	URL := u.String()

	// Make the HTTP request to Plex
	httpResp, respBody, logErr := MakeHTTPRequest(ctx, URL, http.MethodGet, nil, 60, nil, "Plex")
	if logErr.Message != "" {
		return nil, 0, logErr
	}
	defer httpResp.Body.Close()

	// Check the response status code
	if httpResp.StatusCode != 200 {
		logAction.SetError("Plex server returned non-200 status", fmt.Sprintf("Status Code: %d", httpResp.StatusCode), nil)
		return nil, 0, *logAction.Error
	}

	// Decode the response body
	var plexResponse PlexLibraryItemsWrapper
	logErr = DecodeJSONBody(ctx, respBody, &plexResponse, "PlexLibraryItemsWrapper")
	if logErr.Message != "" {
		return nil, 0, logErr
	}

	var items []MediaItem
	for _, item := range plexResponse.MediaContainer.Metadata {

		var itemInfo MediaItem
		itemInfo.RatingKey = item.RatingKey
		itemInfo.Type = item.Type
		itemInfo.Title = item.Title
		itemInfo.Year = item.Year
		itemInfo.LibraryTitle = plexResponse.MediaContainer.LibrarySectionTitle
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

		if item.Type == "movie" {
			itemInfo.Movie = &MediaItemMovie{
				File: MediaItemFile{
					Path:     item.Media[0].Part[0].File,
					Size:     item.Media[0].Part[0].Size,
					Duration: item.Media[0].Part[0].Duration,
				},
			}
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

	if len(items) == 0 {
		logAction.Status = logging.LevelWarn
		logAction.AppendWarning("message", fmt.Sprintf("Library section '%s' is empty", section.Title))
	}

	return items, plexResponse.MediaContainer.TotalSize, logging.LogErrorInfo{}
}

//////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////

func EJ_FetchLibrarySectionItems(ctx context.Context, section LibrarySection, sectionStartIndex string, limit string) ([]MediaItem, int, logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, fmt.Sprintf("Getting all content for '%s' (ID: %s | Start Index: %s) from %s",
		section.Title, section.ID, sectionStartIndex, Global_Config.MediaServer.Type), logging.LevelDebug)
	defer logAction.Complete()

	// If limit is not provided, set it to 500
	if limit == "" {
		limit = "500" // Default limit if not provided
	}

	// Construct the URL for the Emby/Jellyfin server API request
	u, err := url.Parse(Global_Config.MediaServer.URL)
	if err != nil {
		logAction.SetError("Failed to parse URL", err.Error(), nil)
		return nil, 0, *logAction.Error
	}
	u.Path = path.Join(u.Path, "Users", Global_Config.MediaServer.UserID, "Items")
	query := u.Query()
	query.Add("Recursive", "true")
	query.Add("SortBy", "Name")
	query.Add("SortOrder", "Ascending")
	query.Add("IncludeItemTypes", "Movie,Series")
	query.Add("Fields", "DateLastContentAdded,PremiereDate,DateCreated,ProviderIds,BasicSyncInfo,CanDelete,CanDownload,PrimaryImageAspectRatio,ProductionYear,Status,EndDate")
	query.Add("ParentId", section.ID)
	query.Add("StartIndex", sectionStartIndex)
	query.Add("Limit", limit)
	u.RawQuery = query.Encode()
	URL := u.String()
	// Make the HTTP request to Emby/Jellyfin
	httpResp, respBody, logErr := MakeHTTPRequest(ctx, URL, http.MethodGet, nil, 60, nil, "MediaServer")
	if logErr.Message != "" {
		return nil, 0, logErr
	}
	defer httpResp.Body.Close()

	var responseSection EmbyJellyLibraryItemsResponse
	logErr = DecodeJSONBody(ctx, respBody, &responseSection, "EmbyJellyLibraryItemsResponse")
	if logErr.Message != "" {
		return nil, 0, logErr
	}

	// Check to see if any items were returned
	if len(responseSection.Items) == 0 {
		logAction.Status = logging.LevelWarn
		logAction.AppendWarning("message", fmt.Sprintf("Library section '%s' is empty", section.Title))
		return []MediaItem{}, 0, logging.LogErrorInfo{}
	}

	var items []MediaItem
	for _, item := range responseSection.Items {
		var itemInfo MediaItem

		// If Type is Boxset, then split them up
		if item.Type == "BoxSet" {
			// Split the BoxSet into individual items
			boxSetItems, boxSetErr := SplitCollectionIntoIndividualItems(ctx, item.Name, item.ID, section.Title)
			if boxSetErr.Message != "" {
				return nil, 0, boxSetErr
			}
			items = append(items, boxSetItems...)
			continue
		}

		itemInfo.RatingKey = item.ID
		itemInfo.Type = map[string]string{
			"Movie":  "movie",
			"Series": "show",
		}[item.Type]
		itemInfo.Title = item.Name
		itemInfo.Year = item.ProductionYear
		itemInfo.Thumb = item.ImageTags.Thumb
		itemInfo.LibraryTitle = section.Title
		if item.ProviderIds.Tmdb != "" {
			itemInfo.Guids = append(itemInfo.Guids, Guid{Provider: "tmdb", ID: item.ProviderIds.Tmdb})
			itemInfo.TMDB_ID = item.ProviderIds.Tmdb
		}
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
	}

	return items, responseSection.TotalRecordCount, logging.LogErrorInfo{}
}

func SplitCollectionIntoIndividualItems(ctx context.Context, collectionName, parentID, sectionTitle string) ([]MediaItem, logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx,
		fmt.Sprintf("Splitting collection '%s' (ID: %s) into individual items", collectionName, parentID), logging.LevelDebug)
	defer logAction.Complete()

	// Construct the URL for the Emby/Jellyfin server API request
	u, err := url.Parse(Global_Config.MediaServer.URL)
	if err != nil {
		logAction.SetError("Failed to parse URL", err.Error(), nil)
		return nil, *logAction.Error
	}
	u.Path = path.Join(u.Path, "Users", Global_Config.MediaServer.UserID, "Items")
	query := u.Query()
	query.Add("Recursive", "true")
	query.Add("SortBy", "Name")
	query.Add("SortOrder", "Ascending")
	query.Add("IncludeItemTypes", "Movie,Series,BoxSet")
	query.Add("Fields", "DateLastContentAdded,PremiereDate,DateCreated,ProviderIds,BasicSyncInfo,CanDelete,CanDownload,PrimaryImageAspectRatio,ProductionYear,Status,EndDate")
	query.Add("ParentId", parentID)
	u.RawQuery = query.Encode()
	URL := u.String()

	// Make a GET request to the Emby server
	httpResp, respBody, logErr := MakeHTTPRequest(ctx, URL, http.MethodGet, nil, 60, nil, "MediaServer")
	if logErr.Message != "" {
		return nil, logErr
	}
	defer httpResp.Body.Close()

	var responseSection EmbyJellyLibraryItemsResponse
	logErr = DecodeJSONBody(ctx, respBody, &responseSection, "EmbyJellyLibraryItemsResponse")
	if logErr.Message != "" {
		return nil, logErr
	}

	if len(responseSection.Items) == 0 {
		logAction.Status = logging.LevelWarn
		logAction.AppendWarning("message", fmt.Sprintf("Collection '%s' is empty", collectionName))
		return []MediaItem{}, logging.LogErrorInfo{}
	}

	var items []MediaItem
	for _, item := range responseSection.Items {
		var itemInfo MediaItem

		itemInfo.RatingKey = item.ID
		itemInfo.Type = map[string]string{
			"Movie":  "movie",
			"Series": "show",
		}[item.Type]
		itemInfo.Title = item.Name
		itemInfo.Year = item.ProductionYear
		itemInfo.Thumb = item.ImageTags.Thumb
		itemInfo.LibraryTitle = sectionTitle
		if item.ProviderIds.Tmdb != "" {
			itemInfo.Guids = append(itemInfo.Guids, Guid{Provider: "tmdb", ID: item.ProviderIds.Tmdb})
			itemInfo.TMDB_ID = item.ProviderIds.Tmdb
		}
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
	}
	return items, logging.LogErrorInfo{}
}
