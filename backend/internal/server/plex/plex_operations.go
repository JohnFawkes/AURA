package plex

import (
	"aura/internal/config"
	"aura/internal/logging"
	"aura/internal/modals"
	"aura/internal/utils"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func refreshPlexItem(ratingKey string) logging.StandardError {
	logging.LOG.Trace(fmt.Sprintf("Refreshing Plex item with rating key: %s", ratingKey))

	url := fmt.Sprintf("%s/library/metadata/%s/refresh", config.Global.MediaServer.URL, ratingKey)
	response, _, Err := utils.MakeHTTPRequest(url, "PUT", nil, 60, nil, "MediaServer")
	if Err.Message != "" {
		return Err
	}
	defer response.Body.Close()

	time.Sleep(500 * time.Millisecond)
	return logging.StandardError{}
}

func getPosters(ratingKey string) (string, logging.StandardError) {
	logging.LOG.Trace(fmt.Sprintf("Getting posters for rating key: %s", ratingKey))
	posterURL := fmt.Sprintf("%s/library/metadata/%s/posters", config.Global.MediaServer.URL, ratingKey)
	Err := logging.NewStandardError()

	var response *http.Response
	var body []byte

	// Retry logic for the entire process
	for attempt := 1; attempt <= 3; attempt++ {
		logging.LOG.Trace(fmt.Sprintf("Attempt %d to get posters for rating key: %s", attempt, ratingKey))

		// Make the HTTP request
		response, body, Err = utils.MakeHTTPRequest(posterURL, "GET", nil, 60, nil, "MediaServer")
		if Err.Message != "" {
			logging.LOG.Trace(fmt.Sprintf("Attempt %d failed: %v", attempt, Err.Message))
		} else {
			defer response.Body.Close()
			// Check if the response body starts with valid XML
			if strings.HasPrefix(string(body), "<?xml version=\"1.0\"") {
				// Check if the response status code is OK
				if response.StatusCode == http.StatusOK {
					// Parse the response body into a PlexPhotosResponse struct
					var plexPosters modals.PlexPhotosResponse
					err := xml.Unmarshal(body, &plexPosters)
					if err == nil {
						// Check if the response contains any posters
						if len(plexPosters.Photos) > 0 {
							// Look for the first poster with a provider of "local"
							for _, poster := range plexPosters.Photos {
								if poster.Provider == "local" {
									if poster.RatingKey != "" {
										logging.LOG.Trace(fmt.Sprintf("Poster RatingKey: %s", poster.RatingKey))
										return poster.RatingKey, logging.StandardError{}
									}
								}
							}
							Err.Message = "No local posters found for the item"
							Err.HelpText = "Ensure the item has local posters available."
							Err.Details = fmt.Sprintf("No local posters found for rating key: %s", ratingKey)

						}
						Err.Message = "No posters found for the item"
						Err.HelpText = "Ensure the item has posters available."
						Err.Details = fmt.Sprintf("No posters found for rating key: %s", ratingKey)
					}
					logging.LOG.Trace(fmt.Sprintf("Failed to parse XML response: %v", err))
					Err.Message = "Failed to parse XML response"
					Err.HelpText = "Ensure the Plex server is returning a valid XML response."
					Err.Details = fmt.Sprintf("Error parsing XML response for rating key: %s", ratingKey)
				} else {
					Err.Message = fmt.Sprintf("Received status code '%d' from Plex server", response.StatusCode)
					Err.HelpText = "Ensure the Plex server is running and the item with rating key exists."
					Err.Details = fmt.Sprintf("Received status code '%d' for rating key: %s", response.StatusCode, ratingKey)
				}
			} else {
				Err.Message = "Invalid XML response from Plex server"
				Err.HelpText = "Ensure the Plex server is returning a valid XML response."
				Err.Details = fmt.Sprintf("Response from Plex server: %s", string(body))
			}
		}

		// If this is not the last attempt, refresh the Plex item and retry
		if attempt < 3 {
			logging.LOG.Trace(fmt.Sprintf("Retrying to get posters for rating key: %s in 2 seconds", ratingKey))
			time.Sleep(2 * time.Second) // Wait before retrying
			refreshErr := refreshPlexItem(ratingKey)
			if refreshErr.Message != "" {
				logging.LOG.Trace(fmt.Sprintf("Failed to refresh Plex item: %v", refreshErr.Message))
			}

		} else {
			// If this is the last attempt, return the error
			return "", Err
		}
	}

	// If all attempts fail, return the last error
	return "", Err
}

func setPoster(ratingKey, posterKey, posterType string, failedOnGetPosters bool) logging.StandardError {
	// If failedOnGetPosters is true, always treat SaveImageNextToContent.Enabled as false
	// This means we will use the POST method and the plural endpoint ("arts"/"posters")
	// Otherwise, use the config value to decide between PUT (next to content) and POST (upload)
	// For "backdrop" posterType:
	//   - Use "art" (PUT) or "arts" (POST)
	// For all other poster types:
	//   - Use "poster" (PUT) or "posters" (POST)

	saveNextToContent := config.Global.Images.SaveImageNextToContent.Enabled
	if failedOnGetPosters {
		saveNextToContent = false
	}

	requestMethod := "PUT"
	if !saveNextToContent {
		requestMethod = "POST"
		if posterType == "backdrop" {
			posterType = "arts"
		} else {
			posterType = "posters"
		}
	} else {
		if posterType == "backdrop" {
			posterType = "art"
		} else {
			posterType = "poster"
		}
	}

	logging.LOG.Trace(fmt.Sprintf("Setting %s for rating key: %s", posterType, ratingKey))

	// Use net/url to escape the rating key
	escapedPosterKey := url.QueryEscape(posterKey)

	// Construct the URL for setting the poster
	url := fmt.Sprintf("%s/library/metadata/%s/%s?url=%s", config.Global.MediaServer.URL, ratingKey, posterType, escapedPosterKey)

	response, body, Err := utils.MakeHTTPRequest(url, requestMethod, nil, 60, nil, "MediaServer")
	if Err.Message != "" {
		return Err
	}
	defer response.Body.Close()

	if !strings.HasPrefix(string(body), "/library/metadata/") {

		Err.Message = "Failed to set poster"
		Err.HelpText = "Ensure the Plex server is running and the item with rating key exists."
		Err.Details = fmt.Sprintf("Received response: %s", string(body))
		return Err
	}

	return logging.StandardError{}
}

func removeLabel(ratingKey string, label string) logging.StandardError {
	logging.LOG.Trace(fmt.Sprintf("Removing label '%s' from rating key: %s", label, ratingKey))

	// Construct the query parameter that removes the label.
	// "label%5B%5D.tag.tag-=" is the URL encoded version for "label[].tag.tag-="
	removalParam := fmt.Sprintf("label%%5B%%5D.tag.tag-=%s", url.QueryEscape(label))
	// Construct the request URL using the removal parameter
	requestUrl := fmt.Sprintf("%s/library/metadata/%s?%s", config.Global.MediaServer.URL, ratingKey, removalParam)

	// Send the request via PUT
	response, body, Err := utils.MakeHTTPRequest(requestUrl, "PUT", nil, 60, nil, "MediaServer")
	if Err.Message != "" {
		return Err
	}
	defer response.Body.Close()

	logging.LOG.Trace(fmt.Sprintf("Label removal response: %s", string(body)))
	return logging.StandardError{}
}
