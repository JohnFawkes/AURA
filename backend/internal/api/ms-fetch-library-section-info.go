package api

import (
	"aura/internal/logging"
	"encoding/json"
	"fmt"
	"net/http"
)

func (p *PlexServer) FetchLibrarySectionInfo(library *Config_MediaServerLibrary) (bool, logging.StandardError) {
	return Plex_FetchLibrarySectionInfo(library)
}

func (e *EmbyJellyServer) FetchLibrarySectionInfo(library *Config_MediaServerLibrary) (bool, logging.StandardError) {
	return EJ_FetchLibrarySectionInfo(library)
}

//////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////

func CallFetchLibrarySectionInfo() ([]LibrarySection, logging.StandardError) {
	mediaServer, Err := GetMediaServerInterface(Config_MediaServer{})
	if Err.Message != "" {
		return nil, Err
	}

	var allSections []LibrarySection
	for _, library := range Global_Config.MediaServer.Libraries {
		found, Err := mediaServer.FetchLibrarySectionInfo(&library)
		if Err.Message != "" {
			logging.LOG.Warn(Err.Message)
			continue
		}
		if !found {
			logging.LOG.Warn(fmt.Sprintf("Library section '%s' not found in %s", library.Name, Global_Config.MediaServer.Type))
			continue
		}
		if library.Type != "movie" && library.Type != "show" {
			logging.LOG.Warn(fmt.Sprintf("Library section '%s' is not a movie/show section", library.Name))
			continue
		}

		var section LibrarySection
		section.ID = library.SectionID
		section.Type = library.Type
		section.Title = library.Name
		section.Path = library.Path
		allSections = append(allSections, section)
	}
	return allSections, logging.StandardError{}
}

//////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////

func Plex_FetchLibrarySectionInfo(library *Config_MediaServerLibrary) (bool, logging.StandardError) {
	Err := logging.NewStandardError()

	// Construct the URL for the Plex server API request
	url, Err := MakeMediaServerAPIURL("library/sections/all", Global_Config.MediaServer.URL)
	if Err.Message != "" {
		return false, Err
	}

	// Make a GET request to the Plex server
	resp, body, Err := MakeHTTPRequest(url.String(), http.MethodGet, nil, 60, nil, "MediaServer")
	if Err.Message != "" {
		return false, Err
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
		return false, Err
	}

	// Find the library section with the matching Name
	for _, section := range plexResponse.MediaContainer.Directory {
		if section.Title == library.Name {
			library.Type = section.Type
			library.SectionID = section.Key
			library.Path = section.Location[0].Path
			break
		}
	}
	if library.SectionID == "" {
		Err.Message = "Library section not found"
		Err.HelpText = fmt.Sprintf("Ensure the library section '%s' exists on the Plex server.", library.Name)
		Err.Details = map[string]any{
			"request": url.String(),
		}
		return false, Err
	}

	return true, logging.StandardError{}
}

//////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////

func EJ_FetchLibrarySectionInfo(library *Config_MediaServerLibrary) (bool, logging.StandardError) {

	url, Err := MakeMediaServerAPIURL(fmt.Sprintf("/Users/%s/Items", Global_Config.MediaServer.UserID), Global_Config.MediaServer.URL)
	if Err.Message != "" {
		return false, Err
	}

	// Make a GET request to the Emby server
	response, body, Err := MakeHTTPRequest(url.String(), http.MethodGet, nil, 60, nil, "MediaServer")
	if Err.Message != "" {
		logging.LOG.Error(Err.Message)
		return false, Err
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
		return false, Err
	}

	found := false
	for _, item := range responseSection.Items {
		if item.Name == library.Name {
			library.Type = map[string]string{
				"movies":  "movie",
				"tvshows": "show",
			}[item.CollectionType]
			library.SectionID = item.ID
			found = true
			break
		}
	}

	if !found {
		Err.Message = "Library section not found"
		Err.HelpText = fmt.Sprintf("Ensure the library section '%s' exists on the media server.", library.Name)
		return false, Err
	}

	return true, logging.StandardError{}
}
