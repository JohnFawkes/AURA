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
	"time"

	"github.com/go-chi/chi/v5"
)

func GetItemContent(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	logging.LOG.Trace(r.URL.Path)

	// Get the SKU from the URL
	ratingKey := chi.URLParam(r, "ratingKey")
	if ratingKey == "" {
		utils.SendErrorJSONResponse(w, http.StatusInternalServerError, logging.ErrorLog{
			Log: logging.Log{
				Message: "Missing rating key in URL",
				Elapsed: utils.ElapsedTime(startTime),
			},
		})
		return
	}

	itemInfo, logErr := fetchItemContent(ratingKey)
	if logErr.Err != nil {
		utils.SendErrorJSONResponse(w, http.StatusInternalServerError, logErr)
		return
	}

	if itemInfo.RatingKey == "" {
		utils.SendErrorJSONResponse(w, http.StatusInternalServerError, logging.ErrorLog{
			Log: logging.Log{
				Message: "No item found with the given rating key",
				Elapsed: utils.ElapsedTime(startTime),
			},
		})
		return
	}

	// Respond with a success message
	utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
		Status:  "success",
		Message: "Retrieved item content from Plex",
		Elapsed: utils.ElapsedTime(startTime),
		Data:    itemInfo,
	})
}

func fetchItemContent(ratingKey string) (modals.MediaItem, logging.ErrorLog) {
	url := fmt.Sprintf("%s/library/metadata/%s",
		config.Global.Plex.URL, ratingKey)

	var itemInfo modals.MediaItem

	// Make a GET request to the Plex server
	response, body, logErr := utils.MakeHTTPRequest(url, http.MethodGet, nil, 30, nil, "Plex")
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

	// If the item is a movie
	if responseSection.Videos != nil && len(responseSection.Videos) > 0 && responseSection.Directory == nil {
		itemInfo.RatingKey = responseSection.Videos[0].RatingKey
		itemInfo.Type = responseSection.Videos[0].Type
		itemInfo.Title = responseSection.Videos[0].Title
		itemInfo.Year = responseSection.Videos[0].Year
		itemInfo.Thumb = responseSection.Videos[0].Thumb
		itemInfo.AudienceRating = responseSection.Videos[0].AudienceRating
		itemInfo.UserRating = responseSection.Videos[0].UserRating
		itemInfo.ContentRating = responseSection.Videos[0].ContentRating
		itemInfo.Summary = responseSection.Videos[0].Summary
		itemInfo.UpdatedAt = responseSection.Videos[0].UpdatedAt
		itemInfo.Movie = &modals.PlexMovie{
			File: modals.PlexFile{
				Path:     responseSection.Videos[0].Media[0].Part[0].File,
				Size:     responseSection.Videos[0].Media[0].Part[0].Size,
				Duration: responseSection.Videos[0].Media[0].Part[0].Duration,
			},
		}
		return itemInfo, logging.ErrorLog{}
	}

	// If the item is a series
	if responseSection.Directory != nil && len(responseSection.Directory) > 0 && responseSection.Videos == nil {
		itemInfo.RatingKey = responseSection.Directory[0].RatingKey
		itemInfo.Type = responseSection.Directory[0].Type
		itemInfo.Title = responseSection.Directory[0].Title
		itemInfo.Year = responseSection.Directory[0].Year
		itemInfo.Thumb = responseSection.Directory[0].Thumb
		itemInfo.AudienceRating = responseSection.Directory[0].AudienceRating
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
		config.Global.Plex.URL, itemInfo.RatingKey)

	// Make a GET request to fetch children content
	response, body, logErr := utils.MakeHTTPRequest(url, http.MethodGet, nil, 30, nil, "Plex")
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
		var seasons []modals.PlexSeason
		for _, directory := range responseSection.Directory {
			if directory.Title == "All episodes" && directory.Index == 0 {
				continue
			}
			season := modals.PlexSeason{
				RatingKey:    directory.RatingKey,
				SeasonNumber: directory.Index,
				Title:        directory.Title,
				Episodes:     []modals.PlexEpisode{},
			}

			// Fetch episodes for the season
			season, logErr = fetchEpisodesForSeason(season)
			if logErr.Err != nil {
				return *itemInfo, logErr
			}

			seasons = append(seasons, season)
		}
		itemInfo.Series = &modals.PlexSeries{Seasons: seasons}
	}

	return *itemInfo, logging.ErrorLog{}
}

func fetchEpisodesForSeason(season modals.PlexSeason) (modals.PlexSeason, logging.ErrorLog) {
	logging.LOG.Trace(fmt.Sprintf("Fetching episodes for season: %s", season.Title))

	url := fmt.Sprintf("%s/library/metadata/%s/children",
		config.Global.Plex.URL, season.RatingKey)

	// Make a GET request to fetch episodes
	response, body, logErr := utils.MakeHTTPRequest(url, http.MethodGet, nil, 30, nil, "Plex")
	if logErr.Err != nil {
		return season, logErr
	}
	defer response.Body.Close()

	// Check if the response status is OK
	if response.StatusCode != http.StatusOK {
		return season, logging.ErrorLog{
			Err: errors.New(fmt.Sprintf("Received status code '%d' from Plex server", response.StatusCode)),
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
		episode := modals.PlexEpisode{
			RatingKey:     video.RatingKey,
			Title:         video.Title,
			SeasonNumber:  video.ParentIndex,
			EpisodeNumber: video.Index,
			File: modals.PlexFile{
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
