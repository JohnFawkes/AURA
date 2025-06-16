package utils

import (
	"aura/internal/logging"
	"fmt"
	"net/url"
	"path"
)

func MakeMediaServerAPIURL(endpoint, baseURL string) (*url.URL, logging.StandardError) {
	Err := logging.NewStandardError()
	// Check if the base URL is empty
	if baseURL == "" {
		Err.Message = "Base URL is empty"
		Err.HelpText = "Ensure the base URL is set in the configuration."
		return nil, Err
	}

	// Parse the base URL
	baseURLParsed, err := url.Parse(baseURL)
	if err != nil {
		Err.Message = "Failed to parse base URL"
		Err.HelpText = fmt.Sprintf("Ensure the base URL is a valid URL. Error: %s", err.Error())
		return nil, Err
	}

	// Construct the full URL by appending the endpoint
	baseURLParsed.Path = path.Join(baseURLParsed.Path, endpoint)

	return baseURLParsed, logging.StandardError{}
}
