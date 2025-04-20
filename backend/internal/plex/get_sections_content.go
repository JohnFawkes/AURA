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
	"time"
)

func GetSectionsContent(w http.ResponseWriter, r *http.Request) {
	logging.LOG.Trace(r.URL.Path)

	// For each of the library sections, get all items/metadata
	// This is defined in the config file
	var allSections []modals.LibrarySection

	for _, library := range config.Global.Plex.Libraries {
		// Fetch the library section from Plex
		found, logErr := fetchLibrarySection(&library)
		if logErr.Err != nil {
			logging.LOG.Warn(logErr.Log.Message)
			continue
		}
		if !found {
			logging.LOG.Warn(fmt.Sprintf("Library section '%s' not found in Plex", library.Name))
			continue
		}
		// If the section tpye is not movie/show, skip it
		if library.Type != "movie" && library.Type != "show" {
			logging.LOG.Warn(fmt.Sprintf("Library section '%s' is not a movie/show section", library.Name))
			continue
		}

		mediaItems, logErr := fetchSectionsContent(library.SectionID)
		if logErr.Err != nil {
			logging.LOG.Warn(logErr.Log.Message)
			continue
		}
		// Skip sections that are empty or not of type movie/show
		if len(mediaItems) == 0 {
			continue
		}
		var section modals.LibrarySection
		section.ID = library.SectionID
		section.Type = library.Type
		section.Title = library.Name
		section.MediaItems = mediaItems
		allSections = append(allSections, section)
	}

	if len(allSections) == 0 {
		utils.SendErrorJSONResponse(w, http.StatusInternalServerError, logging.ErrorLog{
			Log: logging.Log{
				Message: "No sections found",
				Elapsed: utils.ElapsedTime(time.Now()),
			},
		})
		return
	}

	// Respond with a success message
	utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
		Status:  "success",
		Message: "Retrieved all sections content from Plex",
		Elapsed: utils.ElapsedTime(time.Now()),
		Data:    allSections,
	})
}

// Get all items/metadata for a specific item in a specific library section
// Method: GET
func fetchSectionsContent(sectionID string) ([]modals.MediaItem, logging.ErrorLog) {
	logging.LOG.Trace(fmt.Sprintf("Getting all content for section ID: %s", sectionID))

	// Construct the URL for the Plex server API request
	url := fmt.Sprintf("%s/library/sections/%s/all", config.Global.Plex.URL, sectionID)

	// Make a GET request to the Plex server
	response, body, logErr := utils.MakeHTTPRequest(url, http.MethodGet, nil, 180, nil, "Plex")
	if logErr.Err != nil {
		return nil, logErr
	}
	defer response.Body.Close()

	// Check if the response status is OK
	if response.StatusCode != http.StatusOK {
		return nil, logging.ErrorLog{Err: errors.New("Plex server error"),
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
	if responseSection.Videos != nil && len(responseSection.Videos) > 0 && responseSection.Directory == nil {
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
			itemInfo.Movie = &modals.PlexMovie{
				File: modals.PlexFile{
					Path:     item.Media[0].Part[0].File,
					Size:     item.Media[0].Part[0].Size,
					Duration: item.Media[0].Part[0].Duration,
				},
			}

			items = append(items, itemInfo)
		}
	}

	// If the item is a show section/library
	if responseSection.Directory != nil && len(responseSection.Directory) > 0 && responseSection.Videos == nil {
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
			itemInfo.Series = &modals.PlexSeries{
				SeasonCount:  item.ChildCount,
				EpisodeCount: item.LeafCount,
			}
			items = append(items, itemInfo)
		}
	}

	return items, logging.ErrorLog{}
}

func fetchLibrarySection(library *modals.Config_PlexLibrary) (bool, logging.ErrorLog) {

	// Construct the URL for the Plex server API request
	url := fmt.Sprintf("%s/library/sections", config.Global.Plex.URL)

	// Make a GET request to the Plex server
	response, body, logErr := utils.MakeHTTPRequest(url, http.MethodGet, nil, 60, nil, "Plex")
	if logErr.Err != nil {
		logging.LOG.Error(logErr.Log.Message)
		return false, logErr
	}
	defer response.Body.Close()

	// Check if the response status is OK
	if response.StatusCode != http.StatusOK {
		logging.LOG.Error(fmt.Sprintf("Received status code '%d' from Plex server", response.StatusCode))
		return false, logging.ErrorLog{Log: logging.Log{Message: fmt.Sprintf("Received status code '%d' from Plex server", response.StatusCode)}}
	}

	var responseSection modals.PlexResponse
	err := xml.Unmarshal(body, &responseSection)
	if err != nil {
		logging.LOG.Error("Failed to parse XML response")
		return false, logging.ErrorLog{Err: err,
			Log: logging.Log{Message: "Failed to parse XML response"},
		}
	}

	// Find the library section with the matching Name
	for _, section := range responseSection.Directory {
		if section.Title == library.Name {
			library.Type = section.Type
			library.SectionID = section.Key
			break
		}
	}
	if library.SectionID == "" {
		logging.LOG.Error(fmt.Sprintf("Library section '%s' not found", library.Name))
		return false, logging.ErrorLog{Log: logging.Log{Message: fmt.Sprintf("Library section '%s' not found", library.Name)}}
	}

	return true, logging.ErrorLog{}
}
