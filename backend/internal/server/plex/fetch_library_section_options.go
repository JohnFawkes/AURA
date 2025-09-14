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

func FetchLibrarySectionOptions() ([]string, logging.StandardError) {
	Err := logging.NewStandardError()

	// Construct the URL for the Plex server API request
	url, Err := utils.MakeMediaServerAPIURL("library/sections", config.Global.MediaServer.URL)
	if Err.Message != "" {
		return nil, Err
	}
	// Make a GET request to the Plex server
	response, body, Err := utils.MakeHTTPRequest(url.String(), http.MethodGet, nil, 60, nil, "MediaServer")
	if Err.Message != "" {
		logging.LOG.Error(Err.Message)
		return nil, Err
	}
	defer response.Body.Close()

	var responseSection modals.PlexResponse
	err := xml.Unmarshal(body, &responseSection)
	if err != nil {
		Err.Message = "Failed to parse XML response"
		Err.HelpText = "Ensure the Plex server is returning a valid XML response."
		Err.Details = fmt.Sprintf("Error: %s", err.Error())
		return nil, Err
	}

	var options []string
	for _, section := range responseSection.Directory {
		options = append(options, section.Title)

	}

	return options, logging.StandardError{}
}
