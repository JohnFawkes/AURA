package mediux

import (
	"aura/internal/config"
	"aura/internal/logging"
	"aura/internal/modals"
	"aura/internal/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

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
	var searchType string = itemType
	if itemType == "movie" {
		searchType = "movies"
	}
	// If the itemType is "show", change it to "tv" for the search
	if itemType == "show" {
		searchType = "tv"
	}

	// Construct the URL for the Plex server API request
	baseURL, Err := utils.MakeMediaServerAPIURL("/library/search", config.Global.MediaServer.URL)
	if Err.Message != "" {
		return "", Err
	}

	// Add Query Parameters to the URL
	params := url.Values{}
	params.Add("query", itemTitle)
	params.Add("searchTypes", searchType)
	baseURL.RawQuery = params.Encode()

	// Make a GET request to the Plex server
	response, body, Err := utils.MakeHTTPRequest(baseURL.String(), "GET", nil, 60, nil, "MediaServer")
	if Err.Message != "" {
		return "", Err
	}
	defer response.Body.Close()

	// Parse the response body into a PlexSearchResponse struct
	var plexSearchResponse modals.PlexSearchResponse

	// Output the response body
	err := json.Unmarshal(body, &plexSearchResponse)
	if err != nil {
		Err.Message = "Failed to parse Plex search response"
		Err.HelpText = "Ensure the Plex server is returning a valid JSON response."
		Err.Details = fmt.Sprintf("Error: %s", err.Error())
		return "", Err
	}

	if len(plexSearchResponse.SearchResults) == 0 {
		logging.LOG.Warn(fmt.Sprintf("No items found for %s", itemTitle))
		Err.Message = "No items found for the given search criteria"
		Err.HelpText = fmt.Sprintf("No items found for %s in %s library section", itemTitle, librarySection)
		Err.Details = fmt.Sprintf("Search criteria: tmdbID: %s, itemType: %s, itemTitle: %s, librarySection: %s", tmdbID, itemType, itemTitle, librarySection)
		return "", Err
	}

	logging.LOG.Trace(fmt.Sprintf("Found %d possible matches for %s", len(plexSearchResponse.SearchResults), itemTitle))

	// For each result, go through to find the one with a match to the TMDB ID
	for _, result := range plexSearchResponse.SearchResults {
		searchItem := result.Metadata[0]
		// If the library section does not match, skip it
		if searchItem.LibrarySectionTitle != librarySection {
			continue
		}

		// If the itemTypes do not match, skip it
		if searchItem.Type != itemType {
			continue
		}

		logging.LOG.Trace(fmt.Sprintf("Checking to see if %s is a match for %s", searchItem.Title, itemTitle))

		url, Err := utils.MakeMediaServerAPIURL(fmt.Sprintf("library/metadata/%s", searchItem.RatingKey), config.Global.MediaServer.URL)
		if Err.Message != "" {
			continue
		}
		// Make a GET request to the Plex server
		fullResponse, fullBody, Err := utils.MakeHTTPRequest(url.String(), http.MethodGet, nil, 60, nil, "MediaServer")
		if Err.Message != "" {
			continue
		}
		defer fullResponse.Body.Close()

		// Parse the response body into a PlexLibraryItemsWrapper struct
		var fullResponseSection modals.PlexLibraryItemsWrapper
		err := json.Unmarshal(fullBody, &fullResponseSection)
		if err != nil {
			continue
		}

		for _, guid := range fullResponseSection.MediaContainer.Metadata[0].Guids {
			parts := strings.SplitN(guid.ID, "://", 2)
			if len(parts) == 2 {
				if parts[0] == "tmdb" || parts[0] == "themoviedb" {
					if parts[1] == tmdbID {
						logging.LOG.Debug(fmt.Sprintf("Found TMDB ID match for %s: %s", itemTitle, parts[1]))
						return searchItem.RatingKey, logging.StandardError{}
					}
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
