package api

import (
	"aura/internal/logging"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"
)

func Plex_FetchItemContent(ctx context.Context, ratingKey string) (MediaItem, logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, fmt.Sprintf("Fetching Item Content for Rating Key '%s' from Plex", ratingKey), logging.LevelTrace)
	defer logAction.Complete()

	var itemInfo MediaItem

	// Construct the URL for the Plex server API request
	u, err := url.Parse(Global_Config.MediaServer.URL)
	if err != nil {
		logAction.SetError("Failed to parse base URL", "Ensure the URL is valid", map[string]any{"error": err.Error()})
		return itemInfo, *logAction.Error
	}
	u.Path = path.Join(u.Path, "library/metadata", ratingKey)
	URL := u.String()

	// Make the API request to Plex
	httpResp, respBody, logErr := MakeHTTPRequest(ctx, URL, http.MethodGet, nil, 60, nil, "Plex")
	if logErr.Message != "" {
		return itemInfo, logErr
	}
	defer httpResp.Body.Close()

	// Check the response status code
	if httpResp.StatusCode != http.StatusOK {
		logAction.SetError("Failed to fetch item content from Plex",
			fmt.Sprintf("Plex server returned status code %d", httpResp.StatusCode),
			map[string]any{
				"URL":        URL,
				"StatusCode": httpResp.StatusCode,
			})
		return itemInfo, *logAction.Error
	}

	// Parse the response body into a PlexLibraryItemsWrapper struct
	var plexResponse PlexLibraryItemsWrapper
	logErr = DecodeJSONBody(ctx, respBody, &plexResponse, "PlexLibraryItemsWrapper")
	if logErr.Message != "" {
		return itemInfo, logErr
	}

	// Check if any metadata was returned
	if len(plexResponse.MediaContainer.Metadata) == 0 {
		logAction.SetError("No item found in Plex response",
			"The Plex server did not return any metadata for the requested item.",
			map[string]any{
				"URL": URL,
			})
		return itemInfo, *logAction.Error
	}

	itemInfo.LibraryTitle = plexResponse.MediaContainer.Metadata[0].LibrarySectionTitle
	itemInfo.RatingKey = plexResponse.MediaContainer.Metadata[0].RatingKey
	itemInfo.Type = plexResponse.MediaContainer.Metadata[0].Type
	itemInfo.Title = plexResponse.MediaContainer.Metadata[0].Title
	itemInfo.Year = plexResponse.MediaContainer.Metadata[0].Year
	itemInfo.Thumb = plexResponse.MediaContainer.Metadata[0].Thumb
	itemInfo.ContentRating = plexResponse.MediaContainer.Metadata[0].ContentRating
	itemInfo.Summary = plexResponse.MediaContainer.Metadata[0].Summary
	itemInfo.UpdatedAt = plexResponse.MediaContainer.Metadata[0].UpdatedAt
	itemInfo.AddedAt = plexResponse.MediaContainer.Metadata[0].AddedAt

	if t, err := time.Parse("2006-01-02", plexResponse.MediaContainer.Metadata[0].OriginallyAvailableAt); err == nil {
		itemInfo.ReleasedAt = t.Unix()
	} else {
		itemInfo.ReleasedAt = 0
	}

	switch plexResponse.MediaContainer.Metadata[0].Type {
	case "movie":
		itemInfo.Movie = &MediaItemMovie{
			File: MediaItemFile{
				Path:     plexResponse.MediaContainer.Metadata[0].Media[0].Part[0].File,
				Size:     plexResponse.MediaContainer.Metadata[0].Media[0].Part[0].Size,
				Duration: plexResponse.MediaContainer.Metadata[0].Media[0].Part[0].Duration,
			},
		}
	case "show":
		itemInfo, Err := fetchSeasonsAndEpisodesForShow(ctx, &itemInfo)
		if Err.Message != "" {
			return itemInfo, Err
		}
		itemInfo.Series.SeasonCount = plexResponse.MediaContainer.Metadata[0].ChildCount
		itemInfo.Series.EpisodeCount = plexResponse.MediaContainer.Metadata[0].LeafCount
		itemInfo.Series.Location = plexResponse.MediaContainer.Metadata[0].Location[0].Path
	}

	// Extract GUIDs and Ratings from the response
	guids, _ := getGUIDsAndRatingsFromResponse(ctx, plexResponse.MediaContainer.Metadata[0].Guids,
		plexResponse.MediaContainer.Metadata[0].Ratings,
		fmt.Sprintf("%.1f", plexResponse.MediaContainer.Metadata[0].AudienceRating))
	itemInfo.Guids = guids

	// Check to see if the item has a valid TMDB ID
	for _, guid := range itemInfo.Guids {
		if guid.Provider == "tmdb" && guid.ID != "" {
			itemInfo.TMDB_ID = guid.ID
			break
		}
	}
	if itemInfo.TMDB_ID == "" {
		logAction.SetError("Item does not have a valid TMDB ID",
			"Ensure the item has a valid TMDB ID in its GUIDs to proceed.",
			map[string]any{
				"RatingKey":     itemInfo.RatingKey,
				"Title":         itemInfo.Title,
				"ProviderGUIDs": plexResponse.MediaContainer.Metadata[0].Guids,
			})
		return itemInfo, *logAction.Error
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

	// Return the populated itemInfo struct
	return itemInfo, logging.LogErrorInfo{}
}

func fetchSeasonsAndEpisodesForShow(ctx context.Context, itemInfo *MediaItem) (MediaItem, logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, fmt.Sprintf("Fetching Seasons and Episodes for Show '%s' from Plex", itemInfo.Title), logging.LevelTrace)
	defer logAction.Complete()

	// Make the URL
	u, err := url.Parse(Global_Config.MediaServer.URL)
	if err != nil {
		logAction.SetError("Failed to parse Plex base URL", err.Error(), nil)
		return *itemInfo, *logAction.Error
	}
	u.Path = path.Join(u.Path, "library", "metadata", itemInfo.RatingKey, "allLeaves")
	URL := u.String()

	// Make the API request to Plex
	httpResp, respBody, logErr := MakeHTTPRequest(ctx, URL, http.MethodGet, nil, 60, nil, "Plex")
	if logErr.Message != "" {
		return *itemInfo, logErr
	}
	defer httpResp.Body.Close()

	// Parse the response body into a PlexLibraryItemsWrapper struct
	var plexResponse PlexLibraryItemsWrapper
	logErr = DecodeJSONBody(ctx, respBody, &plexResponse, "PlexLibraryItemsWrapper")
	if logErr.Message != "" {
		return *itemInfo, logErr
	}

	// Group episodes by season number
	seasonsMap := make(map[int]*MediaItemSeason)
	for _, video := range plexResponse.MediaContainer.Metadata {
		seasonNum := video.ParentIndex
		episode := MediaItemEpisode{
			RatingKey:     video.RatingKey,
			Title:         video.Title,
			SeasonNumber:  video.ParentIndex,
			EpisodeNumber: video.Index,
			File: MediaItemFile{
				Path:     video.Media[0].Part[0].File,
				Size:     video.Media[0].Part[0].Size,
				Duration: video.Media[0].Part[0].Duration,
			},
		}
		if _, exists := seasonsMap[seasonNum]; !exists {
			seasonsMap[seasonNum] = &MediaItemSeason{
				SeasonNumber: seasonNum,
				Title:        video.ParentTitle,
				RatingKey:    video.ParentRatingKey,
				Episodes:     []MediaItemEpisode{},
			}
		}
		seasonsMap[seasonNum].Episodes = append(seasonsMap[seasonNum].Episodes, episode)
	}

	// Convert map to slice and sort by season number
	var seasons []MediaItemSeason
	for _, season := range seasonsMap {
		seasons = append(seasons, *season)
	}
	// Optional: sort seasons by SeasonNumber
	sort.Slice(seasons, func(i, j int) bool {
		return seasons[i].SeasonNumber < seasons[j].SeasonNumber
	})

	itemInfo.Series = &MediaItemSeries{Seasons: seasons}
	return *itemInfo, logging.LogErrorInfo{}
}

func getGUIDsAndRatingsFromResponse(ctx context.Context, plexGUIDS []PlexTagField, plexRatings []PlexRatings, audienceRating string) ([]Guid, error) {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Processing GUIDs and Ratings from Plex Response", logging.LevelTrace)
	defer logAction.Complete()

	// Example Ratings From GUIDS:
	// type PlexRatings struct {
	// 	Image string `json:"image"` // Use this to get the provider name as well
	// 	Value string `json:"value"`
	// 	Type  string `json:"type"`
	// }
	// Use the image field to determine the provider
	// Example Ratings:
	// Rating: {image:imdb://image.rating value:7.9 type:audience}
	// Rating: {image:rottentomatoes://image.rating.ripe value:8.1 type:critic}
	// Rating: {image:rottentomatoes://image.rating.upright value:8.2 type:audience}
	// Rating: {image:themoviedb://image.rating value:7.6 type:audience}

	var returnGUIDs []Guid

	// First, process the GUIDs to add them to the returnGUIDs slice
	for _, plexGUID := range plexGUIDS {
		parts := strings.SplitN(plexGUID.ID, "://", 2)
		if len(parts) == 2 {
			provider := strings.ToLower(parts[0])
			id := parts[1]
			returnGUIDs = append(returnGUIDs, Guid{
				Provider: provider,
				ID:       id,
			})
		}
	}

	// Next, process the Ratings to associate them with the correct GUIDs
	for _, plexRating := range plexRatings {
		// Extract provider from the image field
		parts := strings.SplitN(plexRating.Image, "://", 2)
		if len(parts) != 2 {
			continue // Skip if the format is unexpected
		}

		provider := strings.ToLower(parts[0])
		ratingValue := strconv.FormatFloat(plexRating.Value, 'f', -1, 64)

		// Normalize provider if needed
		if provider == "themoviedb" {
			provider = "tmdb"
		}

		// Check if the provider already exists in the returnGUIDs slice using an index-based loop
		found := false
		for i := 0; i < len(returnGUIDs); i++ {
			if returnGUIDs[i].Provider == provider {
				returnGUIDs[i].Rating = ratingValue // assign rating as a single string
				found = true
				break
			}
		}

		// If the provider was not found, add a new GUID with the rating.
		if !found {
			returnGUIDs = append(returnGUIDs, Guid{
				Provider: provider,
				Rating:   ratingValue,
			})
		}
	}

	// Finally, handle the audienceRating if it's provided and valid
	if audienceRating != "" {
		returnGUIDs = append(returnGUIDs, Guid{
			Provider: "community",
			Rating:   audienceRating,
		})
	}

	// Return the final slice of GUIDs with associated ratings
	return returnGUIDs, nil
}
