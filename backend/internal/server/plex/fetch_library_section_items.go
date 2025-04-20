package plex

import (
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"poster-setter/internal/config"
	"poster-setter/internal/logging"
	"poster-setter/internal/modals"
	"poster-setter/internal/utils"
)

// Get all items/metadata for a specific item in a specific library section
func FetchLibrarySectionItems(sectionID string) ([]modals.MediaItem, logging.ErrorLog) {
	logging.LOG.Trace(fmt.Sprintf("Getting all content for section ID: %s", sectionID))

	// Construct the URL for the Plex server API request
	url := fmt.Sprintf("%s/library/sections/%s/all", config.Global.MediaServer.URL, sectionID)

	// Make a GET request to the Plex server
	response, body, logErr := utils.MakeHTTPRequest(url, http.MethodGet, nil, 180, nil, "MediaServer")
	if logErr.Err != nil {
		return nil, logErr
	}
	defer response.Body.Close()

	// Check if the response status is OK
	if response.StatusCode != http.StatusOK {
		return nil, logging.ErrorLog{Err: errors.New("plex server error"),
			Log: logging.Log{Message: fmt.Sprintf("Received status code '%d' from Plex server", response.StatusCode)},
		}
	}

	// Parse the response body into a PlexResponse struct
	var responseSection modals.PlexResponse
	err := xml.Unmarshal(body, &responseSection)
	if err != nil {
		return nil, logging.ErrorLog{Err: err,
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
			itemInfo.Thumb = item.Thumb
			itemInfo.AudienceRating = item.AudienceRating
			itemInfo.UserRating = item.UserRating
			itemInfo.ContentRating = item.ContentRating
			itemInfo.Summary = item.Summary
			itemInfo.UpdatedAt = item.UpdatedAt
			itemInfo.Movie = &modals.MediaItemMovie{
				File: modals.MediaItemFile{
					Path:     item.Media[0].Part[0].File,
					Size:     item.Media[0].Part[0].Size,
					Duration: item.Media[0].Part[0].Duration,
				},
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
			itemInfo.Thumb = item.Thumb
			itemInfo.AudienceRating = item.AudienceRating
			itemInfo.UserRating = item.UserRating
			itemInfo.ContentRating = item.ContentRating
			itemInfo.Summary = item.Summary
			itemInfo.UpdatedAt = item.UpdatedAt
			itemInfo.Series = &modals.MediaItemSeries{
				SeasonCount:  item.ChildCount,
				EpisodeCount: item.LeafCount,
			}
			items = append(items, itemInfo)
		}
	}

	return items, logging.ErrorLog{}
}
