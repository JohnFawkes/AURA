package emby_jellyfin

import (
	"aura/internal/config"
	"aura/internal/database"
	"aura/internal/logging"
	"aura/internal/modals"
	"aura/internal/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func FetchLibrarySectionItems(section modals.LibrarySection, sectionStartIndex string, limit string) ([]modals.MediaItem, int, logging.StandardError) {
	logging.LOG.Trace(fmt.Sprintf("Getting all content for section ID: %s and title: %s", section.ID, section.Title))
	Err := logging.NewStandardError()

	baseURL, Err := utils.MakeMediaServerAPIURL(fmt.Sprintf("Users/%s/Items", config.Global.MediaServer.UserID), config.Global.MediaServer.URL)
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
	response, body, Err := utils.MakeHTTPRequest(baseURL.String(), http.MethodGet, nil, 60, nil, "MediaServer")
	if Err.Message != "" {
		logging.LOG.Error(Err.Message)
		return nil, 0, Err
	}
	defer response.Body.Close()

	var responseSection modals.EmbyJellyLibraryItemsResponse
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

	var items []modals.MediaItem
	for _, item := range responseSection.Items {
		var itemInfo modals.MediaItem

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
			itemInfo.Guids = append(itemInfo.Guids, modals.Guid{Provider: "tmdb", ID: item.ProviderIds.Tmdb})
		}
		itemInfo.AddedAt = item.DateCreated.UnixMilli()
		itemInfo.ReleasedAt = item.PremiereDate.UnixMilli()
		existsInDB, _ := database.CheckIfMediaItemExistsInDatabase(item.ID)

		if existsInDB {
			itemInfo.ExistInDatabase = true
		} else {
			itemInfo.ExistInDatabase = false
		}

		items = append(items, itemInfo)
	}

	return items, responseSection.TotalRecordCount, logging.StandardError{}
}

func SplitCollectionIntoIndividualItems(collectionName, parentID, sectionTitle string) ([]modals.MediaItem, logging.StandardError) {
	logging.LOG.Trace(fmt.Sprintf("Splitting '%s' into individual items (ID: %s)", collectionName, parentID))
	Err := logging.NewStandardError()

	baseURL, Err := utils.MakeMediaServerAPIURL(fmt.Sprintf("Users/%s/Items", config.Global.MediaServer.UserID), config.Global.MediaServer.URL)
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
	response, body, Err := utils.MakeHTTPRequest(baseURL.String(), http.MethodGet, nil, 60, nil, "MediaServer")
	if Err.Message != "" {
		logging.LOG.Error(Err.Message)
		return nil, Err
	}
	defer response.Body.Close()

	var responseSection modals.EmbyJellyLibraryItemsResponse
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

	var items []modals.MediaItem
	for _, item := range responseSection.Items {
		var itemInfo modals.MediaItem

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
			itemInfo.Guids = append(itemInfo.Guids, modals.Guid{Provider: "tmdb", ID: item.ProviderIds.Tmdb})
		}
		itemInfo.AddedAt = item.DateCreated.UnixMilli()
		itemInfo.ReleasedAt = item.PremiereDate.UnixMilli()
		existsInDB, _ := database.CheckIfMediaItemExistsInDatabase(item.ID)

		if existsInDB {
			itemInfo.ExistInDatabase = true
		} else {
			itemInfo.ExistInDatabase = false
		}

		items = append(items, itemInfo)
	}

	return items, Err
}
