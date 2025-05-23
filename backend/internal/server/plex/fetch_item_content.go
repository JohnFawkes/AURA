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
	"regexp"
	"strings"
)

func FetchItemContent(ratingKey string) (modals.MediaItem, logging.ErrorLog) {

	url, logErr := utils.MakeMediaServerAPIURL(fmt.Sprintf("library/metadata/%s", ratingKey), config.Global.MediaServer.URL)
	if logErr.Err != nil {
		return modals.MediaItem{}, logErr
	}
	var itemInfo modals.MediaItem

	// Make a GET request to the Plex server
	response, body, logErr := utils.MakeHTTPRequest(url.String(), http.MethodGet, nil, 60, nil, "MediaServer")
	if logErr.Err != nil {
		return itemInfo, logErr
	}
	defer response.Body.Close()

	// Check if the response status is OK
	if response.StatusCode != http.StatusOK {
		return itemInfo, logging.ErrorLog{Err: logErr.Err,
			Log: logging.Log{Message: fmt.Sprintf("Received status code '%d' from Plex server", response.StatusCode)},
		}
	}

	// Parse the response body into a PlexResponse struct
	var responseSection modals.PlexResponse
	err := xml.Unmarshal(body, &responseSection)
	if err != nil {
		return itemInfo, logging.ErrorLog{Err: err,
			Log: logging.Log{Message: "Failed to parse XML response"},
		}
	}

	// Get GUIDs and Ratings from the response body
	guids, _ := getGUIDsAndRatingsFromBody(body)
	if len(guids) > 0 {
		itemInfo.Guids = guids
	}

	// If the item is a movie
	if len(responseSection.Videos) > 0 && responseSection.Directory == nil {
		itemInfo.LibraryTitle = responseSection.LibrarySectionTitle
		itemInfo.RatingKey = responseSection.Videos[0].RatingKey
		itemInfo.Type = responseSection.Videos[0].Type
		itemInfo.Title = responseSection.Videos[0].Title
		itemInfo.Year = responseSection.Videos[0].Year
		itemInfo.Thumb = responseSection.Videos[0].Thumb
		itemInfo.ContentRating = responseSection.Videos[0].ContentRating
		itemInfo.Summary = responseSection.Videos[0].Summary
		itemInfo.UpdatedAt = responseSection.Videos[0].UpdatedAt
		itemInfo.Movie = &modals.MediaItemMovie{
			File: modals.MediaItemFile{
				Path:     responseSection.Videos[0].Media[0].Part[0].File,
				Size:     responseSection.Videos[0].Media[0].Part[0].Size,
				Duration: responseSection.Videos[0].Media[0].Part[0].Duration,
			},
		}
		// Append the community rating to the guids
		itemInfo.Guids = append(itemInfo.Guids, modals.Guid{
			Provider: "community",
			Rating:   fmt.Sprintf("%.1f", responseSection.Videos[0].AudienceRating),
		})
		return itemInfo, logging.ErrorLog{}
	}

	// If the item is a series
	if len(responseSection.Directory) > 0 && responseSection.Videos == nil {
		itemInfo.LibraryTitle = responseSection.LibrarySectionTitle
		itemInfo.RatingKey = responseSection.Directory[0].RatingKey
		itemInfo.Type = responseSection.Directory[0].Type
		itemInfo.Title = responseSection.Directory[0].Title
		itemInfo.Year = responseSection.Directory[0].Year
		itemInfo.Thumb = responseSection.Directory[0].Thumb
		itemInfo.ContentRating = responseSection.Directory[0].ContentRating
		itemInfo.Summary = responseSection.Directory[0].Summary
		itemInfo.UpdatedAt = responseSection.Directory[0].UpdatedAt
		itemInfo, logErr = fetchSeasonsForShow(&itemInfo)
		if logErr.Err != nil {
			return itemInfo, logErr
		}
		itemInfo.Series.SeasonCount = responseSection.Directory[0].ChildCount
		itemInfo.Series.EpisodeCount = responseSection.Directory[0].LeafCount
		itemInfo.Series.Location = responseSection.Directory[0].Location.Path
		itemInfo.Guids = append(itemInfo.Guids, modals.Guid{
			Provider: "community",
			Rating:   fmt.Sprintf("%.1f", responseSection.Directory[0].AudienceRating),
		})
	}

	existsInDB, _ := database.CheckIfMediaItemExistsInDatabase(itemInfo.RatingKey)
	if existsInDB {
		itemInfo.ExistInDatabase = true
	} else {
		itemInfo.ExistInDatabase = false
	}

	return itemInfo, logging.ErrorLog{}
}

func fetchSeasonsForShow(itemInfo *modals.MediaItem) (modals.MediaItem, logging.ErrorLog) {
	logging.LOG.Trace(fmt.Sprintf("Fetching seasons for show: %s", itemInfo.Title))

	url := fmt.Sprintf("%s/library/metadata/%s/children",
		config.Global.MediaServer.URL, itemInfo.RatingKey)

	// Make a GET request to fetch children content
	response, body, logErr := utils.MakeHTTPRequest(url, http.MethodGet, nil, 60, nil, "MediaServer")
	if logErr.Err != nil {
		return *itemInfo, logErr
	}
	defer response.Body.Close()

	// Check if the response status is OK
	if response.StatusCode != http.StatusOK {
		return *itemInfo, logging.ErrorLog{
			Err: errors.New("received non-200 response from Plex server"),
			Log: logging.Log{Message: "Received non-200 response from Plex server"},
		}
	}

	// Parse the response body into a PlexResponse struct
	var responseSection modals.PlexResponse
	err := xml.Unmarshal(body, &responseSection)
	if err != nil {
		return *itemInfo, logging.ErrorLog{
			Err: err,
			Log: logging.Log{Message: "Failed to parse XML response for seasons"},
		}
	}

	if responseSection.ViewGroup == "season" {
		var seasons []modals.MediaItemSeason
		for _, directory := range responseSection.Directory {
			if directory.Title == "All episodes" && directory.Index == 0 {
				continue
			}
			season := modals.MediaItemSeason{
				RatingKey:    directory.RatingKey,
				SeasonNumber: directory.Index,
				Title:        directory.Title,
				Episodes:     []modals.MediaItemEpisode{},
			}

			// Fetch episodes for the season
			season, logErr = fetchEpisodesForSeason(season)
			if logErr.Err != nil {
				return *itemInfo, logErr
			}

			seasons = append(seasons, season)
		}
		itemInfo.Series = &modals.MediaItemSeries{Seasons: seasons}
	}

	return *itemInfo, logging.ErrorLog{}
}

func fetchEpisodesForSeason(season modals.MediaItemSeason) (modals.MediaItemSeason, logging.ErrorLog) {
	logging.LOG.Trace(fmt.Sprintf("Fetching episodes for season: %s", season.Title))

	url := fmt.Sprintf("%s/library/metadata/%s/children",
		config.Global.MediaServer.URL, season.RatingKey)

	// Make a GET request to fetch episodes
	response, body, logErr := utils.MakeHTTPRequest(url, http.MethodGet, nil, 60, nil, "MediaServer")
	if logErr.Err != nil {
		return season, logErr
	}
	defer response.Body.Close()

	// Check if the response status is OK
	if response.StatusCode != http.StatusOK {
		return season, logging.ErrorLog{
			Err: fmt.Errorf("received status code '%d' from Plex server", response.StatusCode),
			Log: logging.Log{Message: "Received non-200 response from Plex server"},
		}
	}

	// Parse the response body into a PlexResponse struct
	var responseSection modals.PlexResponse
	err := xml.Unmarshal(body, &responseSection)
	if err != nil {
		return season, logging.ErrorLog{
			Err: err,
			Log: logging.Log{Message: "Failed to parse XML response for episodes"},
		}
	}

	// Populate episodes for the season
	for _, video := range responseSection.Videos {
		episode := modals.MediaItemEpisode{
			RatingKey:     video.RatingKey,
			Title:         video.Title,
			SeasonNumber:  video.ParentIndex,
			EpisodeNumber: video.Index,
			File: modals.MediaItemFile{
				Path:     video.Media[0].Part[0].File,
				Size:     video.Media[0].Part[0].Size,
				Duration: video.Media[0].Part[0].Duration,
			},
		}
		season.Episodes = append(season.Episodes, episode)
	}

	return season, logging.ErrorLog{}
}

func getGUIDsAndRatingsFromBody(body []byte) ([]modals.Guid, error) {
	// Use Regex to search for GUIDs manually in the XML response
	// Example GUIDs:
	// <Guid id="imdb://tt######" />
	// <Guid id="tmdb://######" />
	// <Guid id="tvdb://#####" />
	guidRegex := regexp.MustCompile(`(?i)<guid id="([a-z]+)://([^"]+)" ?/?>`)
	guidMatches := guidRegex.FindAllStringSubmatch(string(body), -1)

	// Check if any GUIDs were found
	if len(guidMatches) == 0 {
		return nil, fmt.Errorf("no GUIDs found in the XML response")
	}

	// Create a slice to hold the GUIDs
	var guids []modals.Guid

	// Iterate over the matches and extract the provider and ID
	for _, match := range guidMatches {
		if len(match) == 3 {
			provider := strings.ToLower(match[1])
			id := match[2]
			guids = append(guids, modals.Guid{
				Provider: provider,
				ID:       id,
			})
		}
	}

	// Use Regex to search for Ratings manually in the XML response
	// This regex matches ratings for audience (and similar types if needed)
	// Example Ratings:
	// <Rating image="imdb://image.rating" value="7.9" type="audience" />
	// <Rating image="rottentomatoes://image.rating.ripe" value="8.1" type="critic" />
	// <Rating image="rottentomatoes://image.rating.upright" value="8.2" type="audience" />
	// <Rating image="themoviedb://image.rating" value="7.6" type="audience" />
	ratingRegex := regexp.MustCompile(`(?i)<Rating\s+image="([a-z]+)://[^"]+"\s+value="([^"]+)"\s+type="audience" ?/?>`)
	ratingMatches := ratingRegex.FindAllStringSubmatch(string(body), -1)

	// If no Ratings were found, simply return the GUIDs slice
	if len(ratingMatches) == 0 {
		return guids, nil
	}

	// Iterate over the rating matches and associate the rating with the proper provider
	for _, match := range ratingMatches {
		if len(match) == 3 {
			provider := strings.ToLower(match[1])
			ratingValue := match[2]

			// Normalize provider if needed
			if provider == "themoviedb" {
				provider = "tmdb"
			}

			// Check if the provider already exists in the GUIDs slice using an index-based loop
			found := false
			for i := 0; i < len(guids); i++ {
				if guids[i].Provider == provider {
					guids[i].Rating = ratingValue // assign rating as a single string
					found = true
					break
				}
			}

			// If the provider was not found, add a new GUID with the rating.
			if !found {
				guids = append(guids, modals.Guid{
					Provider: provider,
					Rating:   ratingValue,
				})
			}
		}
	}

	return guids, nil
}
