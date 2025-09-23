package plex

import (
	"aura/internal/cache"
	"aura/internal/config"
	"aura/internal/database"
	"aura/internal/logging"
	"aura/internal/modals"
	"aura/internal/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

func FetchItemContent(ratingKey string) (modals.MediaItem, logging.StandardError) {
	Err := logging.NewStandardError()

	url, Err := utils.MakeMediaServerAPIURL(fmt.Sprintf("library/metadata/%s", ratingKey), config.Global.MediaServer.URL)
	if Err.Message != "" {
		return modals.MediaItem{}, Err
	}
	var itemInfo modals.MediaItem

	// Make a GET request to the Plex server
	resp, body, Err := utils.MakeHTTPRequest(url.String(), http.MethodGet, nil, 60, nil, "MediaServer")
	if Err.Message != "" {
		return itemInfo, Err
	}
	defer resp.Body.Close()

	// Parse the response body into a PlexLibraryItemsWrapper struct
	var plexResponse modals.PlexLibraryItemsWrapper
	err := json.Unmarshal(body, &plexResponse)
	if err != nil {
		Err.Message = "Failed to parse JSON response"
		Err.HelpText = "Ensure the Plex server is returning a valid JSON response."
		Err.Details = fmt.Sprintf("Error: %s", err.Error())
		return itemInfo, Err
	}

	if len(plexResponse.MediaContainer.Metadata) == 0 {
		Err.Message = "No metadata found for the given rating key"
		Err.HelpText = "Ensure the rating key is correct and the item exists on the Plex server."
		Err.Details = fmt.Sprintf("Rating Key: %s", ratingKey)
		return itemInfo, Err
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

	if plexResponse.MediaContainer.Metadata[0].Type == "movie" {
		itemInfo.Movie = &modals.MediaItemMovie{
			File: modals.MediaItemFile{
				Path:     plexResponse.MediaContainer.Metadata[0].Media[0].Part[0].File,
				Size:     plexResponse.MediaContainer.Metadata[0].Media[0].Part[0].Size,
				Duration: plexResponse.MediaContainer.Metadata[0].Media[0].Part[0].Duration,
			},
		}
	}

	if plexResponse.MediaContainer.Metadata[0].Type == "show" {
		itemInfo, Err = fetchSeasonsAndEpisodesForShow(&itemInfo)
		if Err.Message != "" {
			return itemInfo, Err
		}
		itemInfo.Series.SeasonCount = plexResponse.MediaContainer.Metadata[0].LeafCount
		itemInfo.Series.EpisodeCount = plexResponse.MediaContainer.Metadata[0].ChildCount
		itemInfo.Series.Location = plexResponse.MediaContainer.Metadata[0].Location[0].Path
	}

	// Extract GUIDs and Ratings from the response
	guids, _ := getGUIDsAndRatingsFromResponse(plexResponse.MediaContainer.Metadata[0].Guids,
		plexResponse.MediaContainer.Metadata[0].Ratings,
		fmt.Sprintf("%.1f", plexResponse.MediaContainer.Metadata[0].AudienceRating))
	itemInfo.Guids = guids

	existsInDB, _ := database.CheckIfMediaItemExistsInDatabase(itemInfo.RatingKey)
	if existsInDB {
		itemInfo.ExistInDatabase = true
	} else {
		itemInfo.ExistInDatabase = false
	}

	// Update item in cache
	cache.LibraryCacheStore.UpdateMediaItem(itemInfo.LibraryTitle, &itemInfo)

	// Return the populated itemInfo struct
	return itemInfo, logging.StandardError{}
}

func fetchSeasonsAndEpisodesForShow(itemInfo *modals.MediaItem) (modals.MediaItem, logging.StandardError) {
	logging.LOG.Trace(fmt.Sprintf("Fetching seasons and episodes for show: %s", itemInfo.Title))
	Err := logging.NewStandardError()

	url := fmt.Sprintf("%s/library/metadata/%s/allLeaves",
		config.Global.MediaServer.URL, itemInfo.RatingKey)

	// Make a GET request to fetch all leaves (episodes)
	response, body, Err := utils.MakeHTTPRequest(url, http.MethodGet, nil, 60, nil, "MediaServer")
	if Err.Message != "" {
		return *itemInfo, Err
	}
	defer response.Body.Close()

	// Parse the response body into a PlexResponseWrapper struct
	var responseSection modals.PlexLibraryItemsWrapper
	err := json.Unmarshal(body, &responseSection)
	if err != nil {
		Err.Message = "Failed to parse JSON response for all leaves"
		Err.HelpText = "Ensure the Plex server is returning a valid JSON response."
		Err.Details = fmt.Sprintf("Error: %s", err.Error())
		logging.LOG.Warn(err.Error())
		return *itemInfo, Err
	}

	// Group episodes by season number
	seasonsMap := make(map[int]*modals.MediaItemSeason)
	for _, video := range responseSection.MediaContainer.Metadata {
		seasonNum := video.ParentIndex
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
		if _, exists := seasonsMap[seasonNum]; !exists {
			seasonsMap[seasonNum] = &modals.MediaItemSeason{
				SeasonNumber: seasonNum,
				Title:        video.ParentTitle,
				RatingKey:    video.ParentRatingKey,
				Episodes:     []modals.MediaItemEpisode{},
			}
		}
		seasonsMap[seasonNum].Episodes = append(seasonsMap[seasonNum].Episodes, episode)
	}

	// Convert map to slice and sort by season number
	var seasons []modals.MediaItemSeason
	for _, season := range seasonsMap {
		seasons = append(seasons, *season)
	}
	// Optional: sort seasons by SeasonNumber
	sort.Slice(seasons, func(i, j int) bool {
		return seasons[i].SeasonNumber < seasons[j].SeasonNumber
	})

	itemInfo.Series = &modals.MediaItemSeries{Seasons: seasons}
	return *itemInfo, logging.StandardError{}
}

func getGUIDsAndRatingsFromResponse(plexGUIDS []modals.PlexTagField, plexRatings []modals.PlexRatings, audienceRating string) ([]modals.Guid, error) {

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

	var returnGUIDs []modals.Guid

	// First, process the GUIDs to add them to the returnGUIDs slice
	for _, plexGUID := range plexGUIDS {
		parts := strings.SplitN(plexGUID.ID, "://", 2)
		if len(parts) == 2 {
			provider := strings.ToLower(parts[0])
			id := parts[1]
			returnGUIDs = append(returnGUIDs, modals.Guid{
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
			returnGUIDs = append(returnGUIDs, modals.Guid{
				Provider: provider,
				Rating:   ratingValue,
			})
		}
	}

	// Finally, handle the audienceRating if it's provided and valid
	if audienceRating != "" {
		returnGUIDs = append(returnGUIDs, modals.Guid{
			Provider: "community",
			Rating:   audienceRating,
		})
	}

	// Return the final slice of GUIDs with associated ratings
	return returnGUIDs, nil
}
