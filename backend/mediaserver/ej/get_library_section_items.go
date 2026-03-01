package ej

import (
	"aura/cache"
	"aura/config"
	"aura/database"
	"aura/logging"
	"aura/mediux"
	"aura/models"
	"aura/utils/httpx"
	"context"
	"fmt"
	"net/url"
	"path"
)

func (e *EJ) GetLibrarySectionItems(ctx context.Context, section models.LibrarySection, sectionStartIndex string, limit string) (items []models.MediaItem, totalSize int, Err logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, fmt.Sprintf(
		"%s: Fetching Items for Library Section: %s", config.Current.MediaServer.Type, section.Title,
	), logging.LevelInfo)
	defer logAction.Complete()

	items = []models.MediaItem{}
	totalSize = 0
	Err = logging.LogErrorInfo{}

	// If limit is empty, set a default limit
	if limit == "" {
		limit = "500"
	}

	// Construct the URL for the EJ server API request
	u, err := url.Parse(config.Current.MediaServer.URL)
	if err != nil {
		logAction.SetError(logging.Error_BaseUrlParsing(err))
		return items, totalSize, *logAction.Error
	}
	u.Path = path.Join(u.Path, "Users", config.Current.MediaServer.UserID, "Items")
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

	// Make the HTTP Request to EJ
	resp, respBody, Err := makeRequest(ctx, config.Current.MediaServer, URL, "GET", nil)
	if Err.Message != "" {
		return items, totalSize, *logAction.Error
	}
	defer resp.Body.Close()

	// Decode the Response
	var ejResp EmbyJellyLibraryItemsResponse
	Err = httpx.DecodeResponseToJSON(ctx, respBody, &ejResp, fmt.Sprintf("%s Library Section Items Response", config.Current.MediaServer.Type))
	if Err.Message != "" {
		return items, totalSize, *logAction.Error
	}

	// Check to see if any items were returned
	if len(ejResp.Items) == 0 {
		logAction.AppendWarning("message", fmt.Sprintf("Library Section '%s' is empty", section.Title))
		return items, totalSize, Err
	}

	totalSize = ejResp.TotalRecordCount

	for _, ejItem := range ejResp.Items {
		var item models.MediaItem

		// If Type is Boxset, then split them up
		if ejItem.Type == "BoxSet" {
			// Split the BoxSet into individual items
			boxSetItems, boxSetErr := splitCollectionIntoIndividualItems(ctx, ejItem.Name, ejItem.ID, section.Title)
			if boxSetErr.Message != "" {
				return nil, 0, boxSetErr
			}
			items = append(items, boxSetItems...)
			continue
		}

		item.RatingKey = ejItem.ID
		item.Type = map[string]string{
			"Movie":  "movie",
			"Series": "show",
		}[ejItem.Type]
		item.Title = ejItem.Name
		item.Year = ejItem.ProductionYear
		item.LibraryTitle = section.Title
		if ejItem.ProviderIds.Tmdb != "" {
			item.Guids = append(item.Guids, models.MediaItemGuid{Provider: "tmdb", ID: ejItem.ProviderIds.Tmdb})
			item.Guids = append(item.Guids, models.MediaItemGuid{Provider: "tvdb", ID: ejItem.ProviderIds.Tvdb})
			item.TMDB_ID = ejItem.ProviderIds.Tmdb
		}
		item.AddedAt = ejItem.DateCreated.UnixMilli()
		item.ReleasedAt = ejItem.PremiereDate.UnixMilli()

		// If no TMDB ID found, get the value from MediUX using the GUID[tvdb]
		if item.TMDB_ID == "" {
			for _, guid := range item.Guids {
				if guid.Provider == "tvdb" {
					tmdbID, found, Err := mediux.SearchTMDBIDByTVDBID(ctx, guid.ID, item.Type)
					if Err.Message != "" {
						logAction.AppendWarning("search_tmdb_id_error", "Failed to search TMDB ID from MediUX")
					}
					if found {
						item.TMDB_ID = tmdbID
						break
					}
				}
			}
		}
		if item.TMDB_ID == "" {
			totalSize-- // Decrement total size as this item will be skipped
			continue    // Skip items without TMDB ID
		}

		// Check if Media Item exists in DB
		ignored, ignoredMode, sets, logErr := database.CheckIfMediaItemExists(ctx, item.TMDB_ID, item.LibraryTitle)
		if logErr.Message != "" {
			logAction.AppendWarning("message", "Failed to check if media item exists in database")
			logAction.AppendWarning("error", Err)
		}
		if !ignored {
			item.DBSavedSets = sets
		} else {
			item.IgnoredInDB = true
			item.IgnoredMode = ignoredMode
		}

		// Check if Media Item exists in MediUX with a set
		if cache.MediuxItems.CheckItemExists(item.Type, item.TMDB_ID) {
			item.HasMediuxSets = true
		}

		items = append(items, item)
	}

	return items, totalSize, logging.LogErrorInfo{}
}

func splitCollectionIntoIndividualItems(ctx context.Context, collectionName, parentID, sectionTitle string) (items []models.MediaItem, Err logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, fmt.Sprintf(
		"Splitting BoxSet Collection: %s in Section: %s into Individual Items", collectionName, sectionTitle,
	), logging.LevelInfo)
	defer logAction.Complete()

	items = []models.MediaItem{}
	Err = logging.LogErrorInfo{}

	// Construct the URL for the EJ server API request
	u, err := url.Parse(config.Current.MediaServer.URL)
	if err != nil {
		logAction.SetError(logging.Error_BaseUrlParsing(err))
		return items, *logAction.Error
	}
	u.Path = path.Join(u.Path, "Users", config.Current.MediaServer.UserID, "Items")
	query := u.Query()
	query.Add("Recursive", "true")
	query.Add("SortBy", "Name")
	query.Add("SortOrder", "Ascending")
	query.Add("IncludeItemTypes", "Movie,Series")
	query.Add("Fields", "DateLastContentAdded,PremiereDate,DateCreated,ProviderIds,BasicSyncInfo,CanDelete,CanDownload,PrimaryImageAspectRatio,ProductionYear,Status,EndDate")
	query.Add("ParentId", parentID)
	u.RawQuery = query.Encode()
	URL := u.String()

	// Make the HTTP Request to EJ
	resp, respBody, Err := makeRequest(ctx, config.Current.MediaServer, URL, "GET", nil)
	if Err.Message != "" {
		return items, *logAction.Error
	}
	defer resp.Body.Close()

	// Decode the Response
	var ejResp EmbyJellyLibraryItemsResponse
	Err = httpx.DecodeResponseToJSON(ctx, respBody, &ejResp, fmt.Sprintf("%s BoxSet Individual Items Response", config.Current.MediaServer.Type))
	if Err.Message != "" {
		return items, *logAction.Error
	}

	// Check to see if any items were returned
	if len(ejResp.Items) == 0 {
		logAction.AppendWarning("message", fmt.Sprintf("BoxSet Collection '%s' is empty", collectionName))
		return items, Err
	}

	for _, item := range ejResp.Items {
		var itemInfo models.MediaItem

		itemInfo.RatingKey = item.ID
		itemInfo.Type = map[string]string{
			"Movie":  "movie",
			"Series": "show",
		}[item.Type]

		itemInfo.Title = item.Name
		itemInfo.Year = item.ProductionYear
		itemInfo.LibraryTitle = sectionTitle
		if item.ProviderIds.Tmdb != "" {
			itemInfo.Guids = append(itemInfo.Guids, models.MediaItemGuid{Provider: "tmdb", ID: item.ProviderIds.Tmdb})
			itemInfo.TMDB_ID = item.ProviderIds.Tmdb
		}
		itemInfo.AddedAt = item.DateCreated.UnixMilli()
		itemInfo.ReleasedAt = item.PremiereDate.UnixMilli()

		items = append(items, itemInfo)
	}
	return items, logging.LogErrorInfo{}
}
