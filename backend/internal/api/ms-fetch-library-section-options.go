package api

import (
	"aura/internal/logging"
	"encoding/json"
	"fmt"
	"net/http"
)

func (p *PlexServer) FetchLibrarySectionOptions() ([]string, logging.StandardError) {
	// Fetch the library section from Plex
	return Plex_FetchLibrarySectionOptions()
}

func (e *EmbyJellyServer) FetchLibrarySectionOptions() ([]string, logging.StandardError) {
	// Fetch the library section from Emby/Jellyfin
	MediaServer_Init(Global_Config.MediaServer)
	return EJ_FetchLibrarySectionOptions()
}

/////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////

func CallFetchLibrarySectionOptions() ([]string, logging.StandardError) {
	mediaServer, Err := GetMediaServerInterface(Config_MediaServer{})
	if Err.Message != "" {
		return nil, Err
	}

	return mediaServer.FetchLibrarySectionOptions()
}

//////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////

func Plex_FetchLibrarySectionOptions() ([]string, logging.StandardError) {
	Err := logging.NewStandardError()

	// Construct the URL for the Plex server API request
	url, Err := MakeMediaServerAPIURL("library/sections/all", Global_Config.MediaServer.URL)
	if Err.Message != "" {
		return nil, Err
	}
	// Make a GET request to the Plex server
	resp, body, Err := MakeHTTPRequest(url.String(), http.MethodGet, nil, 60, nil, "MediaServer")
	if Err.Message != "" {
		logging.LOG.Error(Err.Message)
		return nil, Err
	}
	defer resp.Body.Close()

	// Parse the response body into a PlexLibrarySectionsWrapper struct
	var plexResponse PlexLibrarySectionsWrapper
	err := json.Unmarshal(body, &plexResponse)
	if err != nil {
		Err.Message = "Failed to parse JSON response"
		Err.HelpText = "Ensure the Plex server is returning a valid JSON response."
		Err.Details = map[string]any{
			"error":   err.Error(),
			"request": url.String(),
		}
		return nil, Err
	}

	var options []string
	for _, section := range plexResponse.MediaContainer.Directory {
		options = append(options, section.Title)

	}

	return options, logging.StandardError{}
}

///////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////

func EJ_FetchLibrarySectionOptions() ([]string, logging.StandardError) {

	url, Err := MakeMediaServerAPIURL(fmt.Sprintf("/Users/%s/Items", Global_Config.MediaServer.UserID), Global_Config.MediaServer.URL)
	if Err.Message != "" {
		return nil, Err
	}

	// Make a GET request to the Emby server
	response, body, Err := MakeHTTPRequest(url.String(), http.MethodGet, nil, 60, nil, "MediaServer")
	if Err.Message != "" {
		logging.LOG.Error(Err.Message)
		return nil, Err
	}
	defer response.Body.Close()

	Err = logging.NewStandardError()

	var responseSection EmbyJellyLibrarySectionsResponse
	err := json.Unmarshal(body, &responseSection)
	if err != nil {
		Err.Message = "Failed to parse JSON response"
		Err.HelpText = "Ensure the Emby/Jellyfin server is returning a valid JSON response."
		Err.Details = map[string]any{
			"error": err.Error(),
		}
		return nil, Err
	}

	var options []string
	for _, item := range responseSection.Items {
		if item.CollectionType != "tvshows" && item.CollectionType != "movies" {
			continue
		}
		options = append(options, item.Name)
	}

	return options, logging.StandardError{}
}
