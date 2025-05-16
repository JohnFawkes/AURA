package emby_jellyfin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"poster-setter/internal/config"
	"poster-setter/internal/logging"
	"poster-setter/internal/modals"
	"poster-setter/internal/utils"
)

func FetchItemContent(ratingKey string, sectionTitle string) (modals.MediaItem, logging.ErrorLog) {

	logging.LOG.Trace(fmt.Sprintf("Fetching item content for rating key: %s", ratingKey))

	baseURL, logErr := utils.MakeMediaServerAPIURL(fmt.Sprintf("/Users/%s/Items/%s", config.Global.MediaServer.UserID, ratingKey),
		config.Global.MediaServer.URL)
	if logErr.Err != nil {
		return modals.MediaItem{}, logErr
	}

	// Add query parameters
	params := url.Values{}
	params.Add("fields", "ShareLevel")
	params.Add("ExcludeFields", "VideoChapters,VideoMediaSources,MediaStreams")
	baseURL.RawQuery = params.Encode()

	logging.LOG.Trace(fmt.Sprintf("Making request to: %s", baseURL.String()))

	// Make a GET request to the Emby/Jellyfin server
	response, body, logErr := utils.MakeHTTPRequest(baseURL.String(), "GET", nil, 60, nil, "MediaServer")
	if logErr.Err != nil {
		return modals.MediaItem{}, logErr
	}
	defer response.Body.Close()

	// Check if the response status is OK
	if response.StatusCode != http.StatusOK {
		return modals.MediaItem{}, logging.ErrorLog{Err: fmt.Errorf("received status code '%d' from %s server", response.StatusCode, config.Global.MediaServer.Type),
			Log: logging.Log{Message: fmt.Sprintf("Received status code '%d' from %s", response.StatusCode, config.Global.MediaServer.Type)}}
	}

	var resp modals.EmbyJellyItemContentResponse
	err := json.Unmarshal(body, &resp)
	if err != nil {
		logging.LOG.Error(fmt.Sprintf("Failed to parse JSON response: %v", err))
		return modals.MediaItem{}, logging.ErrorLog{Err: err,
			Log: logging.Log{Message: "Failed to parse JSON response"},
		}
	}

	var itemInfo modals.MediaItem
	itemInfo.RatingKey = resp.ID
	itemInfo.Type = resp.Type
	itemInfo.Title = resp.Name
	itemInfo.Year = resp.ProductionYear
	itemInfo.LibraryTitle = sectionTitle
	itemInfo.Thumb = resp.ImageTags.Thumb
	itemInfo.ContentRating = resp.OfficialRating
	itemInfo.Summary = resp.Overview
	if resp.ProviderIds.Imdb != "" {
		itemInfo.Guids = append(itemInfo.Guids, modals.Guid{Provider: "imdb", ID: resp.ProviderIds.Imdb})
	}
	if resp.ProviderIds.Tmdb != "" {
		itemInfo.Guids = append(itemInfo.Guids, modals.Guid{Provider: "tmdb", ID: resp.ProviderIds.Tmdb})
	}
	if resp.ProviderIds.Tvdb != "" {
		itemInfo.Guids = append(itemInfo.Guids, modals.Guid{Provider: "tvdb", ID: resp.ProviderIds.Tvdb})
	}
	itemInfo.Guids = append(itemInfo.Guids, modals.Guid{
		Provider: "community",
		Rating:   fmt.Sprintf("%.1f", float64(resp.CommunityRating)),
	})
	itemInfo.Guids = append(itemInfo.Guids, modals.Guid{
		Provider: "critic",
		Rating:   fmt.Sprintf("%.1f", float64(resp.CriticRating)/10.0),
	})

	if resp.Type == "Movie" {
		itemInfo.Type = "movie"
		itemInfo.Movie = &modals.MediaItemMovie{
			File: modals.MediaItemFile{
				Path:     resp.Path,
				Size:     resp.Size,
				Duration: resp.RunTimeTicks / 10000,
			},
		}
	}
	if resp.Type == "Series" {
		itemInfo.Type = "show"
		itemInfo, logErr = fetchSeasonsForShow(&itemInfo)
		if logErr.Err != nil {
			return itemInfo, logErr
		}
		itemInfo.Series.Location = resp.Path
		itemInfo.Series.SeasonCount = resp.ChildCount
		itemInfo.Series.EpisodeCount = 0
		for _, season := range itemInfo.Series.Seasons {
			itemInfo.Series.EpisodeCount += len(season.Episodes)
		}

	}

	return itemInfo, logging.ErrorLog{}

}

func fetchSeasonsForShow(itemInfo *modals.MediaItem) (modals.MediaItem, logging.ErrorLog) {
	logging.LOG.Trace(fmt.Sprintf("Fetching seasons for show: %s", itemInfo.Title))

	baseURL, logErr := utils.MakeMediaServerAPIURL(fmt.Sprintf("/Shows/%s/Seasons", itemInfo.RatingKey),
		config.Global.MediaServer.URL)
	if logErr.Err != nil {
		return *itemInfo, logErr
	}

	// Add query parameters
	params := url.Values{}
	params.Add("Fields", "BasicSyncInfo,CanDelete,CanDownload,PrimaryImageAspectRatio,Overview")
	baseURL.RawQuery = params.Encode()

	// Make a GET request to the Emby/Jellyfin server
	response, body, logErr := utils.MakeHTTPRequest(baseURL.String(), "GET", nil, 60, nil, "MediaServer")
	if logErr.Err != nil {
		return *itemInfo, logErr
	}
	defer response.Body.Close()

	// Check if the response status is OK
	if response.StatusCode != http.StatusOK {
		return *itemInfo, logging.ErrorLog{Err: fmt.Errorf("received status code '%d' from %s server", response.StatusCode, config.Global.MediaServer.Type),
			Log: logging.Log{Message: fmt.Sprintf("Received status code '%d' from %s", response.StatusCode, config.Global.MediaServer.Type)}}
	}

	var resp modals.EmbyJellyItemContentChildResponse
	err := json.Unmarshal(body, &resp)
	if err != nil {
		return *itemInfo, logging.ErrorLog{Err: err,
			Log: logging.Log{Message: "Failed to parse JSON response"},
		}
	}

	var seasons []modals.MediaItemSeason
	for _, season := range resp.Items {
		season := modals.MediaItemSeason{
			RatingKey:    season.ID,
			SeasonNumber: season.IndexNumber,
			Title:        season.Name,
			Episodes:     []modals.MediaItemEpisode{},
		}
		season, logErr = fetchEpisodesForSeason(itemInfo.RatingKey, season)
		if logErr.Err != nil {
			return *itemInfo, logErr
		}
		seasons = append(seasons, season)
	}

	itemInfo.Series = &modals.MediaItemSeries{Seasons: seasons}

	return *itemInfo, logging.ErrorLog{}
}

func fetchEpisodesForSeason(showRatingKey string, season modals.MediaItemSeason) (modals.MediaItemSeason, logging.ErrorLog) {
	logging.LOG.Trace(fmt.Sprintf("Fetching episodes for season: %s", season.Title))

	baseURL, logErr := utils.MakeMediaServerAPIURL(fmt.Sprintf("/Shows/%s/Episodes", showRatingKey),
		config.Global.MediaServer.URL)
	if logErr.Err != nil {
		return season, logErr
	}

	// Add query parameters
	params := url.Values{}
	params.Add("SeasonId", season.RatingKey)
	baseURL.RawQuery = params.Encode()

	// Make a GET request to the Emby/Jellyfin server
	response, body, logErr := utils.MakeHTTPRequest(baseURL.String(), "GET", nil, 60, nil, "MediaServer")
	if logErr.Err != nil {
		return season, logErr
	}
	defer response.Body.Close()

	// Check if the response status is OK
	if response.StatusCode != http.StatusOK {
		return season, logging.ErrorLog{Err: fmt.Errorf("received status code '%d' from %s server", response.StatusCode, config.Global.MediaServer.Type),
			Log: logging.Log{Message: fmt.Sprintf("Received status code '%d' from %s", response.StatusCode, config.Global.MediaServer.Type)}}
	}

	var resp modals.EmbyJellyItemContentChildResponse
	err := json.Unmarshal(body, &resp)
	if err != nil {
		return season, logging.ErrorLog{Err: err,
			Log: logging.Log{Message: "Failed to parse JSON response"},
		}
	}

	for _, episode := range resp.Items {
		episode := modals.MediaItemEpisode{
			RatingKey:     episode.ID,
			Title:         episode.Name,
			SeasonNumber:  episode.ParentIndexNumber,
			EpisodeNumber: episode.IndexNumber,
		}

		// For Emby/Jellyfin, the file path is not directly available in the response.
		// We need to fetch the base Item and then get the file path.
		fetchEpisodeInfo(&episode)
		season.Episodes = append(season.Episodes, episode)
	}

	return season, logging.ErrorLog{}
}

func fetchEpisodeInfo(episode *modals.MediaItemEpisode) (modals.MediaItemEpisode, logging.ErrorLog) {
	logging.LOG.Trace(fmt.Sprintf("Fetching episode info for episode: %s", episode.Title))

	baseURL, logErr := utils.MakeMediaServerAPIURL(fmt.Sprintf("/Users/%s/Items/%s", config.Global.MediaServer.UserID, episode.RatingKey),
		config.Global.MediaServer.URL)
	if logErr.Err != nil {
		return *episode, logErr
	}

	// Make a GET request to the Emby/Jellyfin server
	response, body, logErr := utils.MakeHTTPRequest(baseURL.String(), "GET", nil, 60, nil, "MediaServer")
	if logErr.Err != nil {
		return *episode, logErr
	}
	defer response.Body.Close()

	// Check if the response status is OK
	if response.StatusCode != http.StatusOK {
		return *episode, logging.ErrorLog{Err: fmt.Errorf("received status code '%d' from %s server", response.StatusCode, config.Global.MediaServer.Type),
			Log: logging.Log{Message: fmt.Sprintf("Received status code '%d' from %s", response.StatusCode, config.Global.MediaServer.Type)}}
	}

	var resp modals.EmbyJellyItemContentResponse
	err := json.Unmarshal(body, &resp)
	if err != nil {
		return *episode, logging.ErrorLog{Err: err,
			Log: logging.Log{Message: "Failed to parse JSON response"},
		}
	}

	episode.File.Path = resp.Path
	episode.File.Size = resp.Size
	episode.File.Duration = resp.RunTimeTicks / 10000

	return *episode, logging.ErrorLog{}
}
