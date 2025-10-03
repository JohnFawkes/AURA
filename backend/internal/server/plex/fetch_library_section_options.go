package plex

import (
	"aura/internal/config"
	"aura/internal/logging"
	"aura/internal/modals"
	"aura/internal/utils"
	"encoding/json"
	"net/http"
)

func FetchLibrarySectionOptions() ([]string, logging.StandardError) {
	Err := logging.NewStandardError()

	// Construct the URL for the Plex server API request
	url, Err := utils.MakeMediaServerAPIURL("library/sections/all", config.Global.MediaServer.URL)
	if Err.Message != "" {
		return nil, Err
	}
	// Make a GET request to the Plex server
	resp, body, Err := utils.MakeHTTPRequest(url.String(), http.MethodGet, nil, 60, nil, "MediaServer")
	if Err.Message != "" {
		logging.LOG.Error(Err.Message)
		return nil, Err
	}
	defer resp.Body.Close()

	// Parse the response body into a PlexLibrarySectionsWrapper struct
	var plexResponse modals.PlexLibrarySectionsWrapper
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
