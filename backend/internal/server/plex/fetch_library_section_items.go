package plex

import (
	"aura/internal/config"
	"aura/internal/database"
	"aura/internal/logging"
	"aura/internal/modals"
	"aura/internal/utils"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Get all items/metadata for a specific item in a specific library section
func FetchLibrarySectionItems(section modals.LibrarySection, sectionStartIndex string, limit string) ([]modals.MediaItem, int, logging.StandardError) {
	logging.LOG.Trace(fmt.Sprintf("Getting all content for section ID: %s and title: %s (starting index %s)", section.ID, section.Title, sectionStartIndex))

	// Construct Base URL
	baseURL, Err := utils.MakeMediaServerAPIURL(fmt.Sprintf("library/sections/%s/all", section.ID), config.Global.MediaServer.URL)
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
	response, body, Err := utils.MakeHTTPRequest(baseURL.String(), http.MethodGet, nil, 180, nil, "MediaServer")
	if Err.Message != "" {
		return nil, 0, Err
	}
	defer response.Body.Close()

	// Parse the response body into a PlexResponse struct
	var responseSection modals.PlexResponse
	err := xml.Unmarshal(body, &responseSection)
	if err != nil {
		Err.Message = "Failed to parse XML response"
		Err.HelpText = "Ensure the Plex server is returning a valid XML response."
		Err.Details = fmt.Sprintf("Error: %s", err.Error())
		return nil, 0, Err
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
			itemInfo.Movie = &modals.MediaItemMovie{
				File: modals.MediaItemFile{Path: item.Media[0].Part[0].File,
					Size:     item.Media[0].Part[0].Size,
					Duration: item.Media[0].Part[0].Duration,
				},
			}
			itemInfo.UpdatedAt = item.UpdatedAt
			itemInfo.AddedAt = item.AddedAt
			if t, err := time.Parse("2006-01-02", item.ReleasedAt); err == nil {
				itemInfo.ReleasedAt = t.Unix()
			} else {
				itemInfo.ReleasedAt = 0
			}

			existsInDB, _ := database.CheckIfMediaItemExistsInDatabase(itemInfo.RatingKey)
			if existsInDB {
				itemInfo.ExistInDatabase = true
			} else {
				itemInfo.ExistInDatabase = false
			}

			// Get GUIDs and Ratings from the response body
			if len(item.Guids) > 0 {
				for _, guid := range item.Guids {
					if guid.ID != "" {
						// Sample guid.id : tmdb://######
						// Split into provider and id
						parts := strings.Split(guid.ID, "://")
						if len(parts) == 2 {
							provider := parts[0]
							id := parts[1]
							itemInfo.Guids = append(itemInfo.Guids, modals.Guid{
								Provider: provider,
								ID:       id,
							})
						}

					}
				}
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
			itemInfo.UpdatedAt = item.UpdatedAt
			itemInfo.AddedAt = item.AddedAt
			if t, err := time.Parse("2006-01-02", item.ReleasedAt); err == nil {
				itemInfo.ReleasedAt = t.Unix()
			} else {
				itemInfo.ReleasedAt = 0
			}

			existsInDB, _ := database.CheckIfMediaItemExistsInDatabase(itemInfo.RatingKey)
			if existsInDB {
				itemInfo.ExistInDatabase = true
			} else {
				itemInfo.ExistInDatabase = false
			}

			// Get GUIDs and Ratings from the response body
			if len(item.Guids) > 0 {
				for _, guid := range item.Guids {
					if guid.ID != "" {
						// Sample guid.id : tmdb://######
						// Split into provider and id
						parts := strings.Split(guid.ID, "://")
						if len(parts) == 2 {
							provider := parts[0]
							id := parts[1]
							itemInfo.Guids = append(itemInfo.Guids, modals.Guid{
								Provider: provider,
								ID:       id,
							})
						}
					}
				}
			}

			items = append(items, itemInfo)
		}
	}

	return items, responseSection.TotalSize, logging.StandardError{}
}
