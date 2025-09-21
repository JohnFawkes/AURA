package emby_jellyfin

import (
	"aura/internal/cache"
	"aura/internal/config"
	"aura/internal/database"
	"aura/internal/logging"
	"aura/internal/modals"
	"aura/internal/utils"
	"encoding/json"
	"fmt"
	"net/url"
)

func FetchItemContent(ratingKey string, sectionTitle string) (modals.MediaItem, logging.StandardError) {
	logging.LOG.Trace(fmt.Sprintf("Fetching item content for rating key: %s", ratingKey))
	Err := logging.NewStandardError()

	baseURL, Err := utils.MakeMediaServerAPIURL(fmt.Sprintf("/Users/%s/Items/%s", config.Global.MediaServer.UserID, ratingKey),
		config.Global.MediaServer.URL)
	if Err.Message != "" {
		return modals.MediaItem{}, Err
	}

	// Add query parameters
	params := url.Values{}
	params.Add("fields", "ShareLevel")
	params.Add("ExcludeFields", "VideoChapters,VideoMediaSources,MediaStreams")
	params.Add("Fields", "DateLastContentAdded,PremiereDate,DateCreated,ProviderIds,BasicSyncInfo,CanDelete,CanDownload,PrimaryImageAspectRatio,ProductionYear,Status,EndDate")
	baseURL.RawQuery = params.Encode()

	// Make a GET request to the Emby/Jellyfin server
	response, body, Err := utils.MakeHTTPRequest(baseURL.String(), "GET", nil, 60, nil, "MediaServer")
	if Err.Message != "" {
		return modals.MediaItem{}, Err
	}
	defer response.Body.Close()

	var resp modals.EmbyJellyItemContentResponse
	err := json.Unmarshal(body, &resp)
	if err != nil {
		Err.Message = "Failed to parse JSON response"
		Err.HelpText = "Ensure the Emby/Jellyfin server is returning a valid JSON response."
		Err.Details = fmt.Sprintf("Error: %s", err.Error())
		return modals.MediaItem{}, Err
	}

	var itemInfo modals.MediaItem
	itemInfo.RatingKey = resp.ID
	itemInfo.Type = resp.Type
	itemInfo.Title = resp.Name
	itemInfo.Year = resp.ProductionYear
	itemInfo.LibraryTitle = sectionTitle
	itemInfo.Thumb = resp.ImageTags.Thumb
	itemInfo.ContentRating = resp.OfficialRating
	itemInfo.UpdatedAt = resp.DateCreated.UnixMilli()
	itemInfo.ReleasedAt = resp.PremiereDate.UnixMilli()
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
		itemInfo, Err = fetchSeasonsForShow(&itemInfo)
		if Err.Message != "" {
			return itemInfo, Err
		}
		itemInfo.Series.Location = resp.Path
		itemInfo.Series.SeasonCount = resp.ChildCount
		itemInfo.Series.EpisodeCount = 0
		for _, season := range itemInfo.Series.Seasons {
			itemInfo.Series.EpisodeCount += len(season.Episodes)
		}

	}

	existsInDB, _ := database.CheckIfMediaItemExistsInDatabase(itemInfo.RatingKey)
	if existsInDB {
		itemInfo.ExistInDatabase = true
	} else {
		itemInfo.ExistInDatabase = false
	}

	// Update item in cache
	cache.LibraryCacheStore.UpdateMediaItem(itemInfo.LibraryTitle, &itemInfo)

	return itemInfo, logging.StandardError{}

}

func fetchSeasonsForShow(itemInfo *modals.MediaItem) (modals.MediaItem, logging.StandardError) {
	logging.LOG.Trace(fmt.Sprintf("Fetching seasons for show: %s", itemInfo.Title))
	Err := logging.NewStandardError()

	baseURL, Err := utils.MakeMediaServerAPIURL(fmt.Sprintf("/Shows/%s/Seasons", itemInfo.RatingKey),
		config.Global.MediaServer.URL)
	if Err.Message != "" {
		return *itemInfo, Err
	}

	// Add query parameters
	params := url.Values{}
	params.Add("Fields", "BasicSyncInfo,CanDelete,CanDownload,PrimaryImageAspectRatio,Overview")
	baseURL.RawQuery = params.Encode()

	// Make a GET request to the Emby/Jellyfin server
	response, body, Err := utils.MakeHTTPRequest(baseURL.String(), "GET", nil, 60, nil, "MediaServer")
	if Err.Message != "" {
		return *itemInfo, Err
	}
	defer response.Body.Close()

	var resp modals.EmbyJellyItemContentChildResponse
	err := json.Unmarshal(body, &resp)
	if err != nil {
		Err.Message = "Failed to parse JSON response"
		Err.HelpText = "Ensure the Emby/Jellyfin server is returning a valid JSON response."
		Err.Details = fmt.Sprintf("Error: %s", err.Error())
		return *itemInfo, Err
	}

	var seasons []modals.MediaItemSeason
	for _, season := range resp.Items {
		season := modals.MediaItemSeason{
			RatingKey:    season.ID,
			SeasonNumber: season.IndexNumber,
			Title:        season.Name,
			Episodes:     []modals.MediaItemEpisode{},
		}
		season, Err = fetchEpisodesForSeason(itemInfo.RatingKey, season)
		if Err.Message != "" {
			return *itemInfo, Err
		}
		seasons = append(seasons, season)
	}

	itemInfo.Series = &modals.MediaItemSeries{Seasons: seasons}

	return *itemInfo, logging.StandardError{}
}

func fetchEpisodesForSeason(showRatingKey string, season modals.MediaItemSeason) (modals.MediaItemSeason, logging.StandardError) {
	logging.LOG.Trace(fmt.Sprintf("Fetching episodes for season: %s", season.Title))
	Err := logging.NewStandardError()

	baseURL, Err := utils.MakeMediaServerAPIURL(fmt.Sprintf("/Shows/%s/Episodes", showRatingKey),
		config.Global.MediaServer.URL)
	if Err.Message != "" {
		return season, Err
	}

	// Add query parameters
	params := url.Values{}
	params.Add("SeasonId", season.RatingKey)
	params.Add("Fields", "ID,Name,IndexNumber,ParentIndexNumber,Path,Size,RunTimeTicks")
	baseURL.RawQuery = params.Encode()

	// Make a GET request to the Emby/Jellyfin server
	response, body, Err := utils.MakeHTTPRequest(baseURL.String(), "GET", nil, 60, nil, "MediaServer")
	if Err.Message != "" {
		return season, Err
	}
	defer response.Body.Close()

	var resp modals.EmbyJellyItemContentChildResponse
	err := json.Unmarshal(body, &resp)
	if err != nil {
		Err.Message = "Failed to parse JSON response"
		Err.HelpText = "Ensure the Emby/Jellyfin server is returning a valid JSON response."
		Err.Details = fmt.Sprintf("Error: %s", err.Error())
		return season, Err
	}

	for _, episode := range resp.Items {
		episode := modals.MediaItemEpisode{
			RatingKey:     episode.ID,
			Title:         episode.Name,
			SeasonNumber:  episode.ParentIndexNumber,
			EpisodeNumber: episode.IndexNumber,
			File: modals.MediaItemFile{
				Path:     episode.Path,
				Size:     episode.Size,
				Duration: episode.RunTimeTicks / 10000,
			},
		}

		// For Emby/Jellyfin, the file path is not directly available in the response.
		// We need to fetch the base Item and then get the file path.
		//fetchEpisodeInfo(&episode)
		season.Episodes = append(season.Episodes, episode)
	}

	return season, logging.StandardError{}
}

func fetchEpisodeInfo(episode *modals.MediaItemEpisode) (modals.MediaItemEpisode, logging.StandardError) {
	logging.LOG.Trace(fmt.Sprintf("Fetching episode info for episode: %s", episode.Title))
	Err := logging.NewStandardError()

	baseURL, Err := utils.MakeMediaServerAPIURL(fmt.Sprintf("/Users/%s/Items/%s", config.Global.MediaServer.UserID, episode.RatingKey),
		config.Global.MediaServer.URL)
	if Err.Message != "" {
		return *episode, Err
	}

	// Make a GET request to the Emby/Jellyfin server
	response, body, Err := utils.MakeHTTPRequest(baseURL.String(), "GET", nil, 60, nil, "MediaServer")
	if Err.Message != "" {
		return *episode, Err
	}
	defer response.Body.Close()

	var resp modals.EmbyJellyItemContentResponse
	err := json.Unmarshal(body, &resp)
	if err != nil {
		Err.Message = "Failed to parse JSON response"
		Err.HelpText = "Ensure the Emby/Jellyfin server is returning a valid JSON response."
		Err.Details = fmt.Sprintf("Error: %s", err.Error())
		return *episode, Err
	}

	episode.File.Path = resp.Path
	episode.File.Size = resp.Size
	episode.File.Duration = resp.RunTimeTicks / 10000

	return *episode, logging.StandardError{}
}
