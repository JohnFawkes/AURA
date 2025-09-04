package mediux

import (
	"aura/internal/config"
	"aura/internal/database"
	"aura/internal/logging"
	"aura/internal/modals"
	"aura/internal/utils"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
)

func SearchForItemAndGetRatingKey(tmdbID, itemType, itemTitle, librarySection string) (string, logging.StandardError) {

	// Check if the media server is Plex or Emby/Jellyfin
	switch config.Global.MediaServer.Type {
	case "Plex":
		return PlexSearchForItemAndGetRatingKey(tmdbID, itemType, itemTitle, librarySection)
	case "Emby", "Jellyfin":
		return EmbyJellySearchForItemAndGetRatingKey(tmdbID, itemType, itemTitle, librarySection)
	}
	Err := logging.NewStandardError()

	Err.Message = fmt.Sprintf("Unsupported media server type: %s", config.Global.MediaServer.Type)
	Err.Details = fmt.Sprintf("Media server type must be either 'Plex', 'Emby', or 'Jellyfin', but got '%s'", config.Global.MediaServer.Type)
	return "", Err
}

func PlexSearchForItemAndGetRatingKey(tmdbID, itemType, itemTitle, librarySection string) (string, logging.StandardError) {
	logging.LOG.Trace(fmt.Sprintf("Searching for %s in %s", itemTitle, librarySection))
	Err := logging.NewStandardError()

	// If any of the parameters are empty, return an error
	if tmdbID == "" || itemType == "" || itemTitle == "" || librarySection == "" {

		Err.Message = "Missing parameters for Plex search"
		Err.HelpText = "Ensure that tmdbID, itemType, itemTitle, and librarySection are provided."
		Err.Details = fmt.Sprintf("tmdbID: %s, itemType: %s, itemTitle: %s, librarySection: %s", tmdbID, itemType, itemTitle, librarySection)
		return "", Err
	}

	// If the itemType is not "movie" or "show", return an error
	if itemType != "movie" && itemType != "show" {

		Err.Message = "Invalid itemType for Plex search"
		Err.HelpText = "itemType must be either 'movie' or 'show'."
		Err.Details = fmt.Sprintf("Received itemType: %s", itemType)
		return "", Err
	}

	// If the itemType is "movie", change it to "movies" for the search
	if itemType == "movie" {
		itemType = "movies"
	}
	// If the itemType is "show", change it to "tv" for the search
	if itemType == "show" {
		itemType = "tv"
	}

	// Construct the URL for the Plex server API request
	baseURL, Err := utils.MakeMediaServerAPIURL("/library/search", config.Global.MediaServer.URL)
	if Err.Message != "" {
		return "", Err
	}

	// Add Query Parameters to the URL
	params := url.Values{}
	params.Add("query", itemTitle)
	params.Add("searchTypes", itemType)
	baseURL.RawQuery = params.Encode()

	// Make a GET request to the Plex server
	response, body, Err := utils.MakeHTTPRequest(baseURL.String(), "GET", nil, 60, nil, "MediaServer")
	if Err.Message != "" {
		return "", Err
	}
	defer response.Body.Close()

	// Parse the response body into a PlexSearchResponse struct
	var responseSection modals.PlexSearchResponse

	// Output the response body
	err := xml.Unmarshal(body, &responseSection)
	if err != nil {

		Err.Message = "Failed to parse Plex search response"
		Err.HelpText = "Ensure the Plex server is returning a valid XML response."
		Err.Details = fmt.Sprintf("Error: %s", err.Error())
		return "", Err
	}

	// If the item is a movie section/library
	var items []modals.MediaItem
	for _, result := range responseSection.SearchResults {
		var itemInfo modals.MediaItem
		if itemType == "movies" {
			if result.Video.Type != "movie" ||
				result.Video.LibrarySectionTitle != librarySection {
				continue
			}
			itemInfo.RatingKey = result.Video.RatingKey
			itemInfo.Type = result.Video.Type
			itemInfo.Title = result.Video.Title
			itemInfo.Year = result.Video.Year
			itemInfo.LibraryTitle = result.Video.LibrarySectionTitle
			itemInfo.ExistInDatabase = false
			existsInDB, _ := database.CheckIfMediaItemExistsInDatabase(itemInfo.RatingKey)
			if existsInDB {
				itemInfo.ExistInDatabase = true
			}
			items = append(items, itemInfo)
		} else if itemType == "tv" {
			if result.Directory.Type != "show" ||
				result.Directory.LibrarySectionTitle != librarySection {
				continue
			}
			itemInfo.RatingKey = result.Directory.RatingKey
			itemInfo.Type = result.Directory.Type
			itemInfo.Title = result.Directory.Title
			itemInfo.Year = result.Directory.Year
			itemInfo.LibraryTitle = result.Directory.LibrarySectionTitle
			itemInfo.ExistInDatabase = false
			existsInDB, _ := database.CheckIfMediaItemExistsInDatabase(itemInfo.RatingKey)
			if existsInDB {
				itemInfo.ExistInDatabase = true
			}
			items = append(items, itemInfo)
		}
	}

	// If no items were found, return an error
	if len(items) == 0 {
		logging.LOG.Warn(fmt.Sprintf("No items found for %s", itemTitle))

		Err.Message = "No items found for the given search criteria"
		Err.HelpText = fmt.Sprintf("No items found for %s in %s library section", itemTitle, librarySection)
		Err.Details = fmt.Sprintf("Search criteria: tmdbID: %s, itemType: %s, itemTitle: %s, librarySection: %s", tmdbID, itemType, itemTitle, librarySection)
		return "", Err
	}

	// For each item, grab the full info using the rating key
	// then check to see if the item TMDB ID in the GUID section is a match to the TMDB ID
	// If it is a match, return the rating key
	logging.LOG.Trace(fmt.Sprintf("Found %d items for %s", len(items), itemTitle))
	for _, item := range items {
		logging.LOG.Trace(fmt.Sprintf("Checking to see if %s is a match for %s", item.Title, itemTitle))
		url, Err := utils.MakeMediaServerAPIURL(fmt.Sprintf("library/metadata/%s", item.RatingKey), config.Global.MediaServer.URL)
		if Err.Message != "" {
			continue
		}
		// Make a GET request to the Plex server
		fullResponse, fullBody, Err := utils.MakeHTTPRequest(url.String(), http.MethodGet, nil, 60, nil, "MediaServer")
		if Err.Message != "" {
			continue
		}
		defer fullResponse.Body.Close()

		// Parse the response body into a PlexResponse struct
		var fullResponseSection modals.PlexResponse
		err := xml.Unmarshal(fullBody, &fullResponseSection)
		if err != nil {
			continue
		}
		// Get GUIDs and Ratings from the response body
		guidRegex := regexp.MustCompile(`(?i)tmdb://(\d+)`)
		guidMatches := guidRegex.FindAllStringSubmatch(string(fullBody), -1)
		if len(guidMatches) > 0 {
			logging.LOG.Trace(fmt.Sprintf("Found %d GUID matches for %s", len(guidMatches), itemTitle))
			for _, match := range guidMatches {
				logging.LOG.Trace(fmt.Sprintf("Match: %s", match[0]))
				if len(match) > 1 && match[1] == tmdbID {
					logging.LOG.Debug(fmt.Sprintf("Found TMDB ID match for %s: %s", itemTitle, match[1]))
					// If the TMDB ID matches, return the rating key
					return item.RatingKey, logging.StandardError{}
				}
			}
		}
	}

	logging.LOG.Warn(fmt.Sprintf("No TMDB ID match found for %s", itemTitle))
	return "", logging.StandardError{}
}

func EmbyJellySearchForItemAndGetRatingKey(tmdbID, itemType, itemTitle, librarySection string) (string, logging.StandardError) {
	logging.LOG.Trace(fmt.Sprintf("Searching for '%s' in %s", itemTitle, librarySection))

	baseURL, Err := utils.MakeMediaServerAPIURL("/Items", config.Global.MediaServer.URL)
	if Err.Message != "" {
		return "", Err
	}

	// Add Query Parameters to the URL
	params := url.Values{}
	params.Add("userId", config.Global.MediaServer.UserID)
	params.Add("limit", "100")
	params.Add("recursive", "true")
	params.Add("searchTerm", itemTitle)
	params.Add("IncludeItemTypes", "Movie,Series")
	params.Add("Fields", "ProviderIds")
	baseURL.RawQuery = params.Encode()

	response, body, Err := utils.MakeHTTPRequest(baseURL.String(), "GET", nil, 60, nil, "MediaServer")
	if Err.Message != "" {
		return "", Err
	}
	defer response.Body.Close()

	var resp modals.EmbyJellyLibraryItemsResponse
	err := json.Unmarshal(body, &resp)
	if err != nil {
		logging.LOG.Error(fmt.Sprintf("Failed to parse JSON response: %v", err))

		Err.Message = "Failed to parse Emby/Jellyfin search response"
		Err.HelpText = "Ensure the Emby/Jellyfin server is returning a valid JSON response."
		Err.Details = fmt.Sprintf("Error: %s", err.Error())
		return "", Err
	}

	for _, item := range resp.Items {
		if item.ProviderIds.Tmdb == "" {
			continue
		}
		if tmdbID == item.ProviderIds.Tmdb {
			logging.LOG.Debug(fmt.Sprintf("Found TMDB ID match for '%s': %s", itemTitle, item.ProviderIds.Tmdb))
			return item.ID, logging.StandardError{}
		}
	}

	logging.LOG.Warn(fmt.Sprintf("No TMDB ID match found for '%s'", itemTitle))
	return "", logging.StandardError{}
}
