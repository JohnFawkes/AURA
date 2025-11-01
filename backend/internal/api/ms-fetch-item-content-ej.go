package api

import (
	"aura/internal/logging"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
)

func EJ_FetchItemContent(ctx context.Context, ratingKey string, sectionTitle string) (MediaItem, logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, fmt.Sprintf("Fetch Item Content from %s for ID '%s'", Global_Config.MediaServer.Type, ratingKey), logging.LevelDebug)
	defer logAction.Complete()

	var itemInfo MediaItem

	// Construct the URL for the Emby/Jellyfin server API request
	u, err := url.Parse(Global_Config.MediaServer.URL)
	if err != nil {
		logAction.SetError("Failed to parse Emby/Jellyfin base URL", err.Error(), nil)
		return itemInfo, *logAction.Error
	}
	u.Path = path.Join(u.Path, "Users", Global_Config.MediaServer.UserID, "Items", ratingKey)
	query := u.Query()
	query.Set("fields", "ShareLevel")
	query.Set("ExcludeFields", "VideoChapters,VideoMediaSources,MediaStreams")
	query.Set("Fields", "DateLastContentAdded,PremiereDate,DateCreated,ProviderIds,BasicSyncInfo,CanDelete,CanDownload,PrimaryImageAspectRatio,ProductionYear,Status,EndDate")
	u.RawQuery = query.Encode()
	URL := u.String()

	// Make the Auth Headers for Request
	headers := MakeAuthHeader("X-Emby-Token", Global_Config.MediaServer.Token)

	// Make the API request to Emby/Jellyfin
	httpResp, respBody, logErr := MakeHTTPRequest(ctx, URL, http.MethodGet, headers, 60, nil, Global_Config.MediaServer.Type)
	if logErr.Message != "" {
		return itemInfo, logErr
	}
	defer httpResp.Body.Close()

	// Parse the response body into an EmbyJellyItemContentResponse struct
	var ejResponse EmbyJellyItemContentResponse
	logErr = DecodeJSONBody(ctx, respBody, &ejResponse, "EmbyJellyItemContentResponse")
	if logErr.Message != "" {
		return itemInfo, logErr
	}

	// If item doesn't have a TMDB ID, skip it
	if ejResponse.ProviderIds.Tmdb == "" {
		logAction.SetError("Item does not have a TMDB ID",
			"Only items with a TMDB ID are supported for fetching item content.",
			map[string]any{
				"ID":           ratingKey,
				"SectionTitle": sectionTitle,
				"Title":        ejResponse.Name,
				"ProviderIDs":  ejResponse.ProviderIds,
			})
		return itemInfo, *logAction.Error
	}

	itemInfo.RatingKey = ejResponse.ID
	itemInfo.Type = ejResponse.Type
	itemInfo.Title = ejResponse.Name
	itemInfo.Year = ejResponse.ProductionYear
	itemInfo.LibraryTitle = sectionTitle
	itemInfo.Thumb = ejResponse.ImageTags.Thumb
	itemInfo.ContentRating = ejResponse.OfficialRating
	itemInfo.UpdatedAt = ejResponse.DateCreated.UnixMilli()
	itemInfo.ReleasedAt = ejResponse.PremiereDate.UnixMilli()
	itemInfo.Summary = ejResponse.Overview
	if ejResponse.ProviderIds.Imdb != "" {
		itemInfo.Guids = append(itemInfo.Guids, Guid{Provider: "imdb", ID: ejResponse.ProviderIds.Imdb})
	}
	if ejResponse.ProviderIds.Tmdb != "" {
		itemInfo.Guids = append(itemInfo.Guids, Guid{Provider: "tmdb", ID: ejResponse.ProviderIds.Tmdb})
		itemInfo.TMDB_ID = ejResponse.ProviderIds.Tmdb
	}
	if ejResponse.ProviderIds.Tvdb != "" {
		itemInfo.Guids = append(itemInfo.Guids, Guid{Provider: "tvdb", ID: ejResponse.ProviderIds.Tvdb})
	}
	itemInfo.Guids = append(itemInfo.Guids, Guid{
		Provider: "community",
		Rating:   fmt.Sprintf("%.1f", float64(ejResponse.CommunityRating)),
	})
	itemInfo.Guids = append(itemInfo.Guids, Guid{
		Provider: "critic",
		Rating:   fmt.Sprintf("%.1f", float64(ejResponse.CriticRating)/10.0),
	})

	if ejResponse.Type == "Movie" {
		itemInfo.Type = "movie"
		itemInfo.Movie = &MediaItemMovie{
			File: MediaItemFile{
				Path:     ejResponse.Path,
				Size:     ejResponse.Size,
				Duration: ejResponse.RunTimeTicks / 10000,
			},
		}
	}
	if ejResponse.Type == "Series" {
		itemInfo.Type = "show"
		itemInfo, Err := fetchSeasonsForShow(ctx, &itemInfo)
		if Err.Message != "" {
			return itemInfo, Err
		}
		itemInfo.Series.Location = ejResponse.Path
		itemInfo.Series.SeasonCount = ejResponse.ChildCount
		itemInfo.Series.EpisodeCount = 0
		for _, season := range itemInfo.Series.Seasons {
			itemInfo.Series.EpisodeCount += len(season.Episodes)
		}

	}

	existsInDB, posterSets, _ := DB_CheckIfMediaItemExists(ctx, itemInfo.TMDB_ID, itemInfo.LibraryTitle)
	if existsInDB {
		itemInfo.ExistInDatabase = true
		itemInfo.DBSavedSets = posterSets
	} else {
		itemInfo.ExistInDatabase = false
	}

	// Update item in cache
	Global_Cache_LibraryStore.UpdateMediaItem(itemInfo.LibraryTitle, &itemInfo)

	return itemInfo, logging.LogErrorInfo{}
}

func fetchSeasonsForShow(ctx context.Context, itemInfo *MediaItem) (MediaItem, logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, fmt.Sprintf("Fetching Seasons for Show '%s' from %s", itemInfo.Title, Global_Config.MediaServer.Type), logging.LevelDebug)
	defer logAction.Complete()

	// Construct the URL for the Emby/Jellyfin server API request
	u, err := url.Parse(Global_Config.MediaServer.URL)
	if err != nil {
		logAction.SetError("Failed to parse Emby/Jellyfin base URL", err.Error(), nil)
		return *itemInfo, *logAction.Error
	}
	u.Path = path.Join("Shows", itemInfo.RatingKey, "Seasons")
	query := u.Query()
	query.Set("Fields", "BasicSyncInfo,CanDelete,CanDownload,PrimaryImageAspectRatio,Overview")
	u.RawQuery = query.Encode()
	URL := u.String()

	// Make the Auth Headers for Request
	headers := MakeAuthHeader("X-Emby-Token", Global_Config.MediaServer.Token)

	// Make the API request to Emby/Jellyfin
	httpResp, respBody, logErr := MakeHTTPRequest(ctx, URL, http.MethodGet, headers, 60, nil, Global_Config.MediaServer.Type)
	if logErr.Message != "" {
		return *itemInfo, logErr
	}
	defer httpResp.Body.Close()

	// Parse the response body into an EmbyJellyItemContentChildResponse struct
	var resp EmbyJellyItemContentChildResponse
	logErr = DecodeJSONBody(ctx, respBody, &resp, "EmbyJellyItemContentChildResponse")
	if logErr.Message != "" {
		return *itemInfo, logErr
	}

	var seasons []MediaItemSeason
	for _, season := range resp.Items {
		season := MediaItemSeason{
			RatingKey:    season.ID,
			SeasonNumber: season.IndexNumber,
			Title:        season.Name,
			Episodes:     []MediaItemEpisode{},
		}
		season, Err := fetchEpisodesForSeason(ctx, itemInfo.RatingKey, season)
		if Err.Message != "" {
			return *itemInfo, Err
		}
		seasons = append(seasons, season)
	}

	itemInfo.Series = &MediaItemSeries{Seasons: seasons}

	return *itemInfo, logging.LogErrorInfo{}
}

func fetchEpisodesForSeason(ctx context.Context, showRatingKey string, season MediaItemSeason) (MediaItemSeason, logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, fmt.Sprintf("Fetching Episodes for Season %d of Show ID '%s' from %s", season.SeasonNumber, showRatingKey, Global_Config.MediaServer.Type), logging.LevelDebug)
	defer logAction.Complete()

	// Construct the URL for the Emby/Jellyfin server API request
	u, err := url.Parse(Global_Config.MediaServer.URL)
	if err != nil {
		logAction.SetError("Failed to parse Emby/Jellyfin base URL", err.Error(), nil)
		return season, *logAction.Error
	}
	u.Path = path.Join(u.Path, "Shows", showRatingKey, "Episodes")
	query := u.Query()
	query.Set("SeasonId", season.RatingKey)
	query.Set("Fields", "ID,Name,IndexNumber,ParentIndexNumber,Path,Size,RunTimeTicks")
	u.RawQuery = query.Encode()
	URL := u.String()

	// Make the Auth Headers for Request
	headers := MakeAuthHeader("X-Emby-Token", Global_Config.MediaServer.Token)

	// Make the API request to Emby/Jellyfin
	httpResp, respBody, logErr := MakeHTTPRequest(ctx, URL, http.MethodGet, headers, 60, nil, Global_Config.MediaServer.Type)
	if logErr.Message != "" {
		return season, logErr
	}
	defer httpResp.Body.Close()

	// Parse the response body into an EmbyJellyItemContentChildResponse struct
	var resp EmbyJellyItemContentChildResponse
	logErr = DecodeJSONBody(ctx, respBody, &resp, "EmbyJellyItemContentChildResponse")
	if logErr.Message != "" {
		return season, logErr
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

	return season, logging.LogErrorInfo{}
}
