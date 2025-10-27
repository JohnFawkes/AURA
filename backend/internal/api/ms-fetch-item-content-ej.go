package api

import (
	"aura/internal/logging"
	"encoding/json"
	"fmt"
	"net/url"
)

func EJ_FetchItemContent(ratingKey string, sectionTitle string) (MediaItem, logging.StandardError) {
	logging.LOG.Trace(fmt.Sprintf("Fetching item content for rating key: %s", ratingKey))
	Err := logging.NewStandardError()

	baseURL, Err := MakeMediaServerAPIURL(fmt.Sprintf("/Users/%s/Items/%s", Global_Config.MediaServer.UserID, ratingKey),
		Global_Config.MediaServer.URL)
	if Err.Message != "" {
		return MediaItem{}, Err
	}

	// Add query parameters
	params := url.Values{}
	params.Add("fields", "ShareLevel")
	params.Add("ExcludeFields", "VideoChapters,VideoMediaSources,MediaStreams")
	params.Add("Fields", "DateLastContentAdded,PremiereDate,DateCreated,ProviderIds,BasicSyncInfo,CanDelete,CanDownload,PrimaryImageAspectRatio,ProductionYear,Status,EndDate")
	baseURL.RawQuery = params.Encode()

	// Make a GET request to the Emby/Jellyfin server
	response, body, Err := MakeHTTPRequest(baseURL.String(), "GET", nil, 60, nil, "MediaServer")
	if Err.Message != "" {
		return MediaItem{}, Err
	}
	defer response.Body.Close()

	var resp EmbyJellyItemContentResponse
	err := json.Unmarshal(body, &resp)
	if err != nil {
		Err.Message = "Failed to parse JSON response"
		Err.HelpText = "Ensure the Emby/Jellyfin server is returning a valid JSON response."
		Err.Details = map[string]any{
			"error": err.Error(),
		}
		return MediaItem{}, Err
	}

	// If item doesn't have a TMDB ID, skip it
	if resp.ProviderIds.Tmdb == "" {
		Err.Message = "Item does not have a valid TMDB ID"
		Err.HelpText = "Only items with a valid TMDB ID can be processed."
		Err.Details = map[string]any{
			"ratingKey":   ratingKey,
			"title":       resp.Name,
			"providerIDs": resp.ProviderIds,
		}
		return MediaItem{}, Err
	}

	var itemInfo MediaItem
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
		itemInfo.Guids = append(itemInfo.Guids, Guid{Provider: "imdb", ID: resp.ProviderIds.Imdb})
	}
	if resp.ProviderIds.Tmdb != "" {
		itemInfo.Guids = append(itemInfo.Guids, Guid{Provider: "tmdb", ID: resp.ProviderIds.Tmdb})
		itemInfo.TMDB_ID = resp.ProviderIds.Tmdb
	}
	if resp.ProviderIds.Tvdb != "" {
		itemInfo.Guids = append(itemInfo.Guids, Guid{Provider: "tvdb", ID: resp.ProviderIds.Tvdb})
	}
	itemInfo.Guids = append(itemInfo.Guids, Guid{
		Provider: "community",
		Rating:   fmt.Sprintf("%.1f", float64(resp.CommunityRating)),
	})
	itemInfo.Guids = append(itemInfo.Guids, Guid{
		Provider: "critic",
		Rating:   fmt.Sprintf("%.1f", float64(resp.CriticRating)/10.0),
	})

	if resp.Type == "Movie" {
		itemInfo.Type = "movie"
		itemInfo.Movie = &MediaItemMovie{
			File: MediaItemFile{
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

	existsInDB, posterSets, _ := DB_CheckIfMediaItemExists(itemInfo.TMDB_ID, itemInfo.LibraryTitle)
	if existsInDB {
		itemInfo.ExistInDatabase = true
		itemInfo.DBSavedSets = posterSets
	} else {
		itemInfo.ExistInDatabase = false
	}

	// Update item in cache
	Global_Cache_LibraryStore.UpdateMediaItem(itemInfo.LibraryTitle, &itemInfo)

	return itemInfo, logging.StandardError{}

}

func fetchSeasonsForShow(itemInfo *MediaItem) (MediaItem, logging.StandardError) {
	logging.LOG.Trace(fmt.Sprintf("Fetching seasons for show: %s", itemInfo.Title))
	Err := logging.NewStandardError()

	baseURL, Err := MakeMediaServerAPIURL(fmt.Sprintf("/Shows/%s/Seasons", itemInfo.RatingKey),
		Global_Config.MediaServer.URL)
	if Err.Message != "" {
		return *itemInfo, Err
	}

	// Add query parameters
	params := url.Values{}
	params.Add("Fields", "BasicSyncInfo,CanDelete,CanDownload,PrimaryImageAspectRatio,Overview")
	baseURL.RawQuery = params.Encode()

	// Make a GET request to the Emby/Jellyfin server
	response, body, Err := MakeHTTPRequest(baseURL.String(), "GET", nil, 60, nil, "MediaServer")
	if Err.Message != "" {
		return *itemInfo, Err
	}
	defer response.Body.Close()

	var resp EmbyJellyItemContentChildResponse
	err := json.Unmarshal(body, &resp)
	if err != nil {
		Err.Message = "Failed to parse JSON response"
		Err.HelpText = "Ensure the Emby/Jellyfin server is returning a valid JSON response."
		Err.Details = map[string]any{
			"error": err.Error(),
		}
		return *itemInfo, Err
	}

	var seasons []MediaItemSeason
	for _, season := range resp.Items {
		season := MediaItemSeason{
			RatingKey:    season.ID,
			SeasonNumber: season.IndexNumber,
			Title:        season.Name,
			Episodes:     []MediaItemEpisode{},
		}
		season, Err = fetchEpisodesForSeason(itemInfo.RatingKey, season)
		if Err.Message != "" {
			return *itemInfo, Err
		}
		seasons = append(seasons, season)
	}

	itemInfo.Series = &MediaItemSeries{Seasons: seasons}

	return *itemInfo, logging.StandardError{}
}

func fetchEpisodesForSeason(showRatingKey string, season MediaItemSeason) (MediaItemSeason, logging.StandardError) {
	logging.LOG.Trace(fmt.Sprintf("Fetching episodes for season: %s", season.Title))
	Err := logging.NewStandardError()

	baseURL, Err := MakeMediaServerAPIURL(fmt.Sprintf("/Shows/%s/Episodes", showRatingKey),
		Global_Config.MediaServer.URL)
	if Err.Message != "" {
		return season, Err
	}

	// Add query parameters
	params := url.Values{}
	params.Add("SeasonId", season.RatingKey)
	params.Add("Fields", "ID,Name,IndexNumber,ParentIndexNumber,Path,Size,RunTimeTicks")
	baseURL.RawQuery = params.Encode()

	// Make a GET request to the Emby/Jellyfin server
	response, body, Err := MakeHTTPRequest(baseURL.String(), "GET", nil, 60, nil, "MediaServer")
	if Err.Message != "" {
		return season, Err
	}
	defer response.Body.Close()

	var resp EmbyJellyItemContentChildResponse
	err := json.Unmarshal(body, &resp)
	if err != nil {
		Err.Message = "Failed to parse JSON response"
		Err.HelpText = "Ensure the Emby/Jellyfin server is returning a valid JSON response."
		Err.Details = map[string]any{
			"error": err.Error(),
		}
		return season, Err
	}

	for _, episode := range resp.Items {
		episode := MediaItemEpisode{
			RatingKey:     episode.ID,
			Title:         episode.Name,
			SeasonNumber:  episode.ParentIndexNumber,
			EpisodeNumber: episode.IndexNumber,
			File: MediaItemFile{
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

func fetchEpisodeInfo(episode *MediaItemEpisode) (MediaItemEpisode, logging.StandardError) {
	logging.LOG.Trace(fmt.Sprintf("Fetching episode info for episode: %s", episode.Title))
	Err := logging.NewStandardError()

	baseURL, Err := MakeMediaServerAPIURL(fmt.Sprintf("/Users/%s/Items/%s", Global_Config.MediaServer.UserID, episode.RatingKey),
		Global_Config.MediaServer.URL)
	if Err.Message != "" {
		return *episode, Err
	}

	// Make a GET request to the Emby/Jellyfin server
	response, body, Err := MakeHTTPRequest(baseURL.String(), "GET", nil, 60, nil, "MediaServer")
	if Err.Message != "" {
		return *episode, Err
	}
	defer response.Body.Close()

	var resp EmbyJellyItemContentResponse
	err := json.Unmarshal(body, &resp)
	if err != nil {
		Err.Message = "Failed to parse JSON response"
		Err.HelpText = "Ensure the Emby/Jellyfin server is returning a valid JSON response."
		Err.Details = map[string]any{
			"error": err.Error(),
		}
		return *episode, Err
	}

	episode.File.Path = resp.Path
	episode.File.Size = resp.Size
	episode.File.Duration = resp.RunTimeTicks / 10000

	return *episode, logging.StandardError{}
}
