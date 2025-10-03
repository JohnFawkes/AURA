package plex

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
	resp, body, Err := utils.MakeHTTPRequest(baseURL.String(), http.MethodGet, nil, 180, nil, "MediaServer")
	if Err.Message != "" {
		return nil, 0, Err
	}
	defer resp.Body.Close()

	// Parse the response body into a PlexResponse struct
	var plexResponse modals.PlexLibraryItemsWrapper
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

	var items []modals.MediaItem
	for _, item := range plexResponse.MediaContainer.Metadata {

		var itemInfo modals.MediaItem
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
			itemInfo.Movie = &modals.MediaItemMovie{
				File: modals.MediaItemFile{
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
						itemInfo.Guids = append(itemInfo.Guids, modals.Guid{
							Provider: provider,
							ID:       id,
						})
					}
				}
			}
		}

		existsInDB, _ := database.CheckIfMediaItemExistsInDatabase(itemInfo.RatingKey)
		if existsInDB {
			itemInfo.ExistInDatabase = true
		} else {
			itemInfo.ExistInDatabase = false
		}

		items = append(items, itemInfo)

	}

	return items, plexResponse.MediaContainer.TotalSize, logging.StandardError{}
}
