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

func FetchLibrarySectionItems(section modals.LibrarySection, sectionStartIndex string) ([]modals.MediaItem, int, logging.ErrorLog) {
	logging.LOG.Trace(fmt.Sprintf("Getting all content for section ID: %s and title: %s", section.ID, section.Title))

	baseURL, logErr := utils.MakeMediaServerAPIURL(fmt.Sprintf("Users/%s/Items", config.Global.MediaServer.UserID), config.Global.MediaServer.URL)
	if logErr.Err != nil {
		return nil, 0, logErr
	}

	// Add query parameters
	params := url.Values{}
	params.Add("Recursive", "true")
	params.Add("SortBy", "Name")
	params.Add("SortOrder", "Ascending")
	params.Add("IncludeItemTypes", "Movie,Series")
	params.Add("Fields", "BasicSyncInfo,CanDelete,CanDownload,PrimaryImageAspectRatio,ProductionYear,Status,EndDate")
	params.Add("ParentId", section.ID)
	params.Add("StartIndex", sectionStartIndex)
	params.Add("Limit", "500")

	baseURL.RawQuery = params.Encode()

	// Make a GET request to the Emby server
	response, body, logErr := utils.MakeHTTPRequest(baseURL.String(), http.MethodGet, nil, 60, nil, "MediaServer")
	if logErr.Err != nil {
		logging.LOG.Error(logErr.Log.Message)
		return nil, 0, logErr
	}
	defer response.Body.Close()

	// Check if the response status is OK
	if response.StatusCode != http.StatusOK {
		return nil, 0, logging.ErrorLog{Err: fmt.Errorf("received status code '%d' from %s server", response.StatusCode, config.Global.MediaServer.Type),
			Log: logging.Log{Message: fmt.Sprintf("Received status code '%d' from %s server", response.StatusCode, config.Global.MediaServer.Type)}}
	}

	var responseSection modals.EmbyJellyLibraryItemsResponse
	err := json.Unmarshal(body, &responseSection)
	if err != nil {
		return nil, 0, logging.ErrorLog{Err: err, Log: logging.Log{Message: "Failed to parse JSON response"}}
	}

	// Check to see if any items were returned
	if len(responseSection.Items) == 0 {
		return nil, 0, logging.ErrorLog{Log: logging.Log{Message: "No items found in the response"}}
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
		itemInfo.LibraryTitle = section.Title

		existsInDB, _ := database.CheckIfMediaItemExistsInDatabase(item.ID)

		if existsInDB {
			itemInfo.ExistInDatabase = true
		} else {
			itemInfo.ExistInDatabase = false
		}

		items = append(items, itemInfo)
	}

	return items, responseSection.TotalRecordCount, logging.ErrorLog{}
}
