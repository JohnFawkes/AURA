package api

import (
	"aura/internal/logging"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func (p *PlexServer) FetchLibrarySectionItems(section LibrarySection, sectionStartIndex string) ([]MediaItem, int, logging.StandardError) {
	return Plex_FetchLibrarySectionItems(section, sectionStartIndex, "500")
}

func (e *EmbyJellyServer) FetchLibrarySectionItems(section LibrarySection, sectionStartIndex string) ([]MediaItem, int, logging.StandardError) {
	return EJ_FetchLibrarySectionItems(section, sectionStartIndex, "500")
}

/////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////

func CallFetchLibrarySectionItems(sectionID, sectionTitle, sectionType string, sectionStartIndex string) (LibrarySection, logging.StandardError) {
	var section LibrarySection
	section.ID = sectionID
	section.Title = sectionTitle
	section.Type = sectionType

	mediaServer, Err := GetMediaServerInterface(Config_MediaServer{})
	if Err.Message != "" {
		return section, Err
	}

	// Fetch the section items from the media server
	mediaItems, totalSize, Err := mediaServer.FetchLibrarySectionItems(section, sectionStartIndex)
	if Err.Message != "" {
		logging.LOG.Warn(Err.Message)
		return section, Err
	}
	if len(mediaItems) == 0 {
		logging.LOG.Warn(fmt.Sprintf("Library section '%s' is empty", section.Title))
		Err.Message = "No items found in the library section"
		Err.HelpText = fmt.Sprintf("Ensure the section '%s' has items.", section.Title)
		Err.Details = map[string]any{
			"error": fmt.Sprintf("No items found for section ID '%s' with title '%s'.", section.ID, section.Title),
		}
		return section, Err
	}
	section.MediaItems = mediaItems
	section.TotalSize = totalSize
	Global_Cache_LibraryStore.UpdateSection(&section)
	return section, logging.StandardError{}
}

/////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////

// Get all items/metadata for a specific item in a specific library section
func Plex_FetchLibrarySectionItems(section LibrarySection, sectionStartIndex string, limit string) ([]MediaItem, int, logging.StandardError) {
	logging.LOG.Trace(fmt.Sprintf("Getting all content for '%s' (ID: %s | Starting Index: %s)", section.Title, section.ID, sectionStartIndex))

	// Construct Base URL
	baseURL, Err := MakeMediaServerAPIURL(fmt.Sprintf("library/sections/%s/all", section.ID), Global_Config.MediaServer.URL)
	if Err.Message != "" {
		return nil, 0, Err
	}

	// If limit is not provided, set it to 500
	if limit == "" {
		limit = "500" // Default limit if not provided
	}

	// Add parameters to the URL
	params := url.Values{}
	params.Add("X-Plex-Container-Start", sectionStartIndex)
	params.Add("X-Plex-Container-Size", limit)
	params.Add("includeGuids", "1")
	baseURL.RawQuery = params.Encode()

	// Make a GET request to the Plex server
	resp, body, Err := MakeHTTPRequest(baseURL.String(), http.MethodGet, nil, 180, nil, "MediaServer")
	if Err.Message != "" {
		return nil, 0, Err
	}
	defer resp.Body.Close()

	// Parse the response body into a PlexResponse struct
	var plexResponse PlexLibraryItemsWrapper
	err := json.Unmarshal(body, &plexResponse)
	if err != nil {
		Err.Message = "Failed to parse JSON response"
		Err.HelpText = "Ensure the Plex server is returning a valid JSON response."
		Err.Details = map[string]any{
			"error":   err.Error(),
			"request": baseURL.String(),
		}
		return nil, 0, Err
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

		existsInDB, posterSets, Err := DB_CheckIfMediaItemExists(itemInfo.TMDB_ID, itemInfo.LibraryTitle)
		if Err.Message != "" {
			logging.LOG.Warn(fmt.Sprintf("Failed to check if media item exists in database: %v", Err.Details))
		}
		if existsInDB {
			itemInfo.ExistInDatabase = true
			itemInfo.DBSavedSets = posterSets
		} else {
			itemInfo.ExistInDatabase = false
		}

		items = append(items, itemInfo)

	}

	return items, plexResponse.MediaContainer.TotalSize, logging.StandardError{}
}

/////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////

func EJ_FetchLibrarySectionItems(section LibrarySection, sectionStartIndex string, limit string) ([]MediaItem, int, logging.StandardError) {
	logging.LOG.Trace(fmt.Sprintf("Getting all content for '%s' (ID: %s)", section.Title, section.ID))
	Err := logging.NewStandardError()

	baseURL, Err := MakeMediaServerAPIURL(fmt.Sprintf("Users/%s/Items", Global_Config.MediaServer.UserID), Global_Config.MediaServer.URL)
	if Err.Message != "" {
		return nil, 0, Err
	}

	if limit == "" {
		limit = "500" // Default limit if not provided
	}

	// Add query parameters
	params := url.Values{}
	params.Add("Recursive", "true")
	params.Add("SortBy", "Name")
	params.Add("SortOrder", "Ascending")
	params.Add("IncludeItemTypes", "Movie,Series")
	params.Add("Fields", "DateLastContentAdded,PremiereDate,DateCreated,ProviderIds,BasicSyncInfo,CanDelete,CanDownload,PrimaryImageAspectRatio,ProductionYear,Status,EndDate")
	params.Add("ParentId", section.ID)
	params.Add("StartIndex", sectionStartIndex)
	params.Add("Limit", limit)

	baseURL.RawQuery = params.Encode()

	// Make a GET request to the Emby server
	response, body, Err := MakeHTTPRequest(baseURL.String(), http.MethodGet, nil, 60, nil, "MediaServer")
	if Err.Message != "" {
		logging.LOG.Error(Err.Message)
		return nil, 0, Err
	}
	defer response.Body.Close()

	var responseSection EmbyJellyLibraryItemsResponse
	err := json.Unmarshal(body, &responseSection)
	if err != nil {
		Err.Message = "Failed to parse JSON response"
		Err.HelpText = "Ensure the Emby/Jellyfin server is returning a valid JSON response."
		Err.Details = map[string]any{
			"error": err.Error(),
		}
		return nil, 0, Err
	}

	// Check to see if any items were returned
	if len(responseSection.Items) == 0 {
		Err.Message = "No items found in the library section"
		Err.HelpText = fmt.Sprintf("Ensure the library section '%s' has items.", section.Title)
		Err.Details = map[string]any{
			"sectionID":    section.ID,
			"sectionTitle": section.Title,
		}
		return nil, 0, Err
	}

	var items []MediaItem
	for _, item := range responseSection.Items {
		var itemInfo MediaItem

		// If Type is Boxset, then split them up
		if item.Type == "BoxSet" {
			// Split the BoxSet into individual items
			boxSetItems, boxSetErr := SplitCollectionIntoIndividualItems(item.Name, item.ID, section.Title)
			if boxSetErr.Message != "" {
				logging.LOG.Error(boxSetErr.Message)
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
		existsInDB, posterSets, Err := DB_CheckIfMediaItemExists(itemInfo.TMDB_ID, itemInfo.LibraryTitle)
		if Err.Message != "" {
			logging.LOG.Warn(fmt.Sprintf("Failed to check if media item exists in database: %v", Err.Details))
		}
		if existsInDB {
			itemInfo.ExistInDatabase = true
			itemInfo.DBSavedSets = posterSets
		} else {
			itemInfo.ExistInDatabase = false
		}

		items = append(items, itemInfo)
	}

	return items, responseSection.TotalRecordCount, logging.StandardError{}
}

func SplitCollectionIntoIndividualItems(collectionName, parentID, sectionTitle string) ([]MediaItem, logging.StandardError) {
	logging.LOG.Trace(fmt.Sprintf("Splitting '%s' into individual items (ID: %s)", collectionName, parentID))
	Err := logging.NewStandardError()

	baseURL, Err := MakeMediaServerAPIURL(fmt.Sprintf("Users/%s/Items", Global_Config.MediaServer.UserID), Global_Config.MediaServer.URL)
	if Err.Message != "" {
		return nil, Err
	}

	// Add query parameters
	params := url.Values{}
	params.Add("Recursive", "true")
	params.Add("SortBy", "Name")
	params.Add("SortOrder", "Ascending")
	params.Add("IncludeItemTypes", "Movie,Series,BoxSet")
	params.Add("Fields", "DateLastContentAdded,PremiereDate,DateCreated,ProviderIds,BasicSyncInfo,CanDelete,CanDownload,PrimaryImageAspectRatio,ProductionYear,Status,EndDate")
	params.Add("ParentId", parentID)

	baseURL.RawQuery = params.Encode()

	// Make a GET request to the Emby server
	response, body, Err := MakeHTTPRequest(baseURL.String(), http.MethodGet, nil, 60, nil, "MediaServer")
	if Err.Message != "" {
		logging.LOG.Error(Err.Message)
		return nil, Err
	}
	defer response.Body.Close()

	var responseSection EmbyJellyLibraryItemsResponse
	err := json.Unmarshal(body, &responseSection)
	if err != nil {
		Err.Message = "Failed to parse JSON response"
		Err.HelpText = "Ensure the Emby/Jellyfin server is returning a valid JSON response."
		Err.Details = map[string]any{
			"error": err.Error(),
		}
		return nil, Err
	}

	// Check to see if any items were returned
	if len(responseSection.Items) == 0 {
		Err.Message = "No items found in the library section"
		Err.HelpText = fmt.Sprintf("Ensure the library section '%s' has items.", sectionTitle)
		Err.Details = map[string]any{
			"sectionID":    parentID,
			"sectionTitle": sectionTitle,
		}
		return nil, Err
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
		existsInDB, posterSets, Err := DB_CheckIfMediaItemExists(itemInfo.TMDB_ID, itemInfo.LibraryTitle)
		if Err.Message != "" {
			logging.LOG.Warn(fmt.Sprintf("Failed to check if media item exists in database: %v", Err.Details))
		}
		if existsInDB {
			itemInfo.ExistInDatabase = true
			itemInfo.DBSavedSets = posterSets
		} else {
			itemInfo.ExistInDatabase = false
		}

		items = append(items, itemInfo)
	}

	return items, Err
}
