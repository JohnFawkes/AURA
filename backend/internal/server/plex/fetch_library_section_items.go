package plex

import (
	"aura/internal/config"
	"aura/internal/database"
	"aura/internal/logging"
	"aura/internal/modals"
	"aura/internal/utils"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// Get all items/metadata for a specific item in a specific library section
func FetchLibrarySectionItems(section modals.LibrarySection, sectionStartIndex string) ([]modals.MediaItem, int, logging.ErrorLog) {
	logging.LOG.Trace(fmt.Sprintf("Getting all content for section ID: %s and title: %s", section.ID, section.Title))

	// Construct the URL for the Plex server API request including pagination parameters.
	url := fmt.Sprintf("%s/library/sections/%s/all?X-Plex-Container-Start=%s&X-Plex-Container-Size=%s",
		config.Global.MediaServer.URL, section.ID, sectionStartIndex, "500")

	// Make a GET request to the Plex server
	response, body, logErr := utils.MakeHTTPRequest(url, http.MethodGet, nil, 180, nil, "MediaServer")
	if logErr.Err != nil {
		return nil, 0, logErr
	}
	defer response.Body.Close()

	// Check if the response status is OK
	if response.StatusCode != http.StatusOK {
		return nil, 0, logging.ErrorLog{Err: errors.New("plex server error"),
			Log: logging.Log{Message: fmt.Sprintf("Received status code '%d' from Plex server", response.StatusCode)},
		}
	}

	// Parse the response body into a PlexResponse struct
	var responseSection modals.PlexResponse
	err := xml.Unmarshal(body, &responseSection)
	if err != nil {
		return nil, 0, logging.ErrorLog{Err: err,
			Log: logging.Log{Message: "Failed to parse XML response"},
		}
	}

	// If the item is a movie section/library
	var items []modals.MediaItem
	if len(responseSection.Videos) > 0 && responseSection.Directory == nil {
		for _, item := range responseSection.Videos {
			var itemInfo modals.MediaItem
			itemInfo.RatingKey = item.RatingKey
			itemInfo.Type = item.Type
			itemInfo.Title = item.Title
			itemInfo.Year = item.Year
			itemInfo.LibraryTitle = responseSection.LibrarySectionTitle

			existsInDB, _ := database.CheckIfMediaItemAlreadyInDatabase(itemInfo.RatingKey)
			if existsInDB {
				itemInfo.ExistInDatabase = true
			} else {
				itemInfo.ExistInDatabase = false
			}

			items = append(items, itemInfo)
		}
	}

	// If the item is a show section/library
	if len(responseSection.Directory) > 0 && responseSection.Videos == nil {
		for _, item := range responseSection.Directory {
			var itemInfo modals.MediaItem
			itemInfo.RatingKey = item.RatingKey
			itemInfo.Type = item.Type
			itemInfo.Title = item.Title
			itemInfo.Year = item.Year
			itemInfo.LibraryTitle = responseSection.LibrarySectionTitle

			existsInDB, _ := database.CheckIfMediaItemAlreadyInDatabase(itemInfo.RatingKey)
			if existsInDB {
				itemInfo.ExistInDatabase = true
			} else {
				itemInfo.ExistInDatabase = false
			}

			items = append(items, itemInfo)
		}
	}

	time.Sleep(200 * time.Millisecond)
	return items, responseSection.TotalSize, logging.ErrorLog{}
}
