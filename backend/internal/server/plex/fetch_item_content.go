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
	"regexp"
	"strconv"
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

	// Get GUIDs from the response body
	guids, _ := getGUIDsFromBody(body)
	if len(guids) > 0 {
		itemInfo.Guids = guids
	}

	// Get IMDB rating from the response body
	imdbRating, _ := getIMDBRatingFromBody(body)
	if parsedRating, err := strconv.ParseFloat(imdbRating, 64); err == nil {
		itemInfo.AudienceRating = parsedRating
	}

	// If the item is a movie
	if len(responseSection.Videos) > 0 && responseSection.Directory == nil {
		itemInfo.LibraryTitle = responseSection.LibrarySectionTitle
		itemInfo.RatingKey = responseSection.Videos[0].RatingKey
		itemInfo.Type = responseSection.Videos[0].Type
		itemInfo.Title = responseSection.Videos[0].Title
		itemInfo.Year = responseSection.Videos[0].Year
		itemInfo.Thumb = responseSection.Videos[0].Thumb
		//itemInfo.AudienceRating = responseSection.Videos[0].AudienceRating
		itemInfo.UserRating = responseSection.Videos[0].UserRating
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
		//itemInfo.AudienceRating = responseSection.Directory[0].AudienceRating
		itemInfo.UserRating = responseSection.Directory[0].UserRating
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

func getGUIDsFromBody(body []byte) ([]modals.Guid, error) {
	// Use Regex to search for GUIDs manually in the XML response
	// Grab the provider and ID from the GUIDs
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
			provider := match[1]
			id := match[2]
			guids = append(guids, modals.Guid{
				Provider: provider,
				ID:       id,
			})
		}
	}
	return guids, nil
}

func getIMDBRatingFromBody(body []byte) (string, error) {
	// Regex looks for a Rating tag with image starting with "imdb://" and type="audience"
	// Example tag: <Rating image="imdb://image.rating" value="6.8" type="audience"/>
	ratingRegex := regexp.MustCompile(`(?i)<Rating\s+image="imdb://[^"]+"\s+value="([^"]+)"\s+type="audience" ?/?>`)
	matches := ratingRegex.FindStringSubmatch(string(body))
	if len(matches) < 2 {
		return "", fmt.Errorf("IMDB rating not found in the XML response")
	}
	return matches[1], nil
}
