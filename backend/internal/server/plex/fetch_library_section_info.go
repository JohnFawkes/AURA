package plex

import (
	"aura/internal/config"
	"aura/internal/logging"
	"aura/internal/modals"
	"aura/internal/utils"
	"encoding/json"
	"fmt"
	"net/http"
)

func FetchLibrarySectionInfo(library *modals.Config_MediaServerLibrary) (bool, logging.StandardError) {
	Err := logging.NewStandardError()

	// Construct the URL for the Plex server API request
	url, Err := utils.MakeMediaServerAPIURL("library/sections/all", config.Global.MediaServer.URL)
	if Err.Message != "" {
		return false, Err
	}

	// Make a GET request to the Plex server
	resp, body, Err := utils.MakeHTTPRequest(url.String(), http.MethodGet, nil, 60, nil, "MediaServer")
	if Err.Message != "" {
		return false, Err
	}
	defer resp.Body.Close()

	// Parse the response body into a PlexLibrarySectionsWrapper struct
	var plexResponse modals.PlexLibrarySectionsWrapper
	err := json.Unmarshal(body, &plexResponse)
	if err != nil {
		Err.Message = "Failed to parse JSON response"
		Err.HelpText = "Ensure the Plex server is returning a valid JSON response."
		Err.Details = fmt.Sprintf("Error: %s", err.Error())
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
		Err.Details = fmt.Sprintf("No section with name '%s' found in the Plex server response.", library.Name)
		return false, Err
	}

	return true, logging.StandardError{}
}
