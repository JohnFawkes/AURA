package plex

import (
	"aura/internal/config"
	"aura/internal/logging"
	"aura/internal/modals"
	"aura/internal/utils"
	"encoding/xml"
	"fmt"
	"net/http"
)

func FetchLibrarySectionInfo(library *modals.Config_MediaServerLibrary) (bool, logging.StandardError) {
	Err := logging.NewStandardError()

	// Construct the URL for the Plex server API request
	url, Err := utils.MakeMediaServerAPIURL("library/sections", config.Global.MediaServer.URL)
	if Err.Message != "" {
		return false, Err
	}
	// Make a GET request to the Plex server
	response, body, Err := utils.MakeHTTPRequest(url.String(), http.MethodGet, nil, 60, nil, "MediaServer")
	if Err.Message != "" {
		logging.LOG.Error(Err.Message)
		return false, Err
	}
	defer response.Body.Close()

	var responseSection modals.PlexResponse
	err := xml.Unmarshal(body, &responseSection)
	if err != nil {
		Err.Message = "Failed to parse XML response"
		Err.HelpText = "Ensure the Plex server is returning a valid XML response."
		Err.Details = fmt.Sprintf("Error: %s", err.Error())
		return false, Err
	}

	// Find the library section with the matching Name
	for _, section := range responseSection.Directory {
		if section.Title == library.Name {
			library.Type = section.Type
			library.SectionID = section.Key
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
