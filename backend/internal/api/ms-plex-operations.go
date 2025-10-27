package api

import (
	"aura/internal/logging"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func Plex_RefreshItem(ratingKey string) logging.StandardError {
	logging.LOG.Trace(fmt.Sprintf("Refreshing Plex item with rating key: %s", ratingKey))

	url := fmt.Sprintf("%s/library/metadata/%s/refresh", Global_Config.MediaServer.URL, ratingKey)
	response, _, Err := MakeHTTPRequest(url, "PUT", nil, 60, nil, "MediaServer")
	if Err.Message != "" {
		return Err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		Err.Message = fmt.Sprintf("Failed to refresh Plex item, received status code: %d", response.StatusCode)
		Err.HelpText = fmt.Sprintf("Ensure the item with rating key %s exists. If it does, check the Plex server logs for more information. If it doesn't, please try refreshing aura from the Home page.", ratingKey)
		return Err
	}

	return logging.StandardError{}
}

// Plex_GetPoster attempts to retrieve the poster for a Plex item by ratingKey.
// It tries up to 3 times, refreshing the Plex item between attempts if needed.
// Returns the poster's ratingKey (URL or path) if found, or a StandardError if not.
func Plex_GetPoster(ratingKey string) (string, logging.StandardError) {
	logging.LOG.Trace(fmt.Sprintf("Getting posters for rating key: %s", ratingKey))
	posterURL := fmt.Sprintf("%s/library/metadata/%s/posters", Global_Config.MediaServer.URL, ratingKey)
	logging.LOG.Trace(fmt.Sprintf("Poster URL: %s", posterURL))
	Err := logging.NewStandardError()

	var response *http.Response
	var body []byte

	// Retry logic for the entire process (up to 3 attempts)
	for attempt := 1; attempt <= 3; attempt++ {
		logging.LOG.Trace(fmt.Sprintf("Attempt %d to get posters for rating key: %s", attempt, ratingKey))

		// Make the HTTP request to Plex
		response, body, Err = MakeHTTPRequest(posterURL, "GET", nil, 60, nil, "MediaServer")
		if Err.Message != "" {
			logging.LOG.Trace(fmt.Sprintf("Attempt %d failed: %v", attempt, Err.Message))
		} else {
			defer response.Body.Close()

			// Check if the response status code is OK
			if response.StatusCode == http.StatusOK {
				// Parse the response body into a PlexGetAllImagesWrapper struct
				var plexPosters PlexGetAllImagesWrapper
				err := json.Unmarshal(body, &plexPosters)
				if err != nil {
					// JSON parsing failed
					logging.LOG.Trace(fmt.Sprintf("Failed to parse JSON response: %v", err))
					Err.Message = "Failed to parse JSON response"
					Err.HelpText = "Ensure the Plex server is returning a valid JSON response."
					Err.Details = fmt.Sprintf("Error parsing JSON response for rating key: %s - %s", ratingKey, err.Error())
					// No need to continue processing; break to retry or return error
				} else {
					// Successfully parsed JSON; check for posters
					if len(plexPosters.MediaContainer.Metadata) > 0 {
						// Look for the first poster with a provider of "local"
						for _, poster := range plexPosters.MediaContainer.Metadata {
							if poster.Provider == "local" && poster.RatingKey != "" {
								logging.LOG.Trace(fmt.Sprintf("Poster RatingKey: %s", poster.RatingKey))
								return poster.RatingKey, logging.StandardError{}
							}
						}
						// No local posters found, but posters exist
						Err.Message = "No local posters found for the item"
						Err.HelpText = "Ensure the item has local posters available."
						Err.Details = fmt.Sprintf("No local posters found for rating key: %s", ratingKey)
					} else {
						// No posters found at all
						Err.Message = "No posters found for the item"
						Err.HelpText = "Ensure the item has posters available."
						Err.Details = fmt.Sprintf("No posters found for rating key: %s", ratingKey)
					}
				}
			} else {
				// Non-OK status code from Plex
				Err.Message = fmt.Sprintf("Received status code '%d' from Plex server", response.StatusCode)
				Err.HelpText = fmt.Sprintf("Ensure the item with rating key %s exists. If it does, check the Plex server logs for more information. If it doesn't, please try refreshing aura from the Home page.", ratingKey)
				Err.Details = fmt.Sprintf("Received status code '%d' for rating key: %s", response.StatusCode, ratingKey)
			}
		}

		// If not the last attempt, refresh the Plex item and retry
		if attempt < 3 {
			numberOfSeconds := 2
			logging.LOG.Warn(fmt.Sprintf("Attempt %d to get posters failed: %s", attempt, Err.Message))
			logging.LOG.Trace(fmt.Sprintf("Retrying to get posters for rating key: %s in %d seconds", ratingKey, numberOfSeconds))
			time.Sleep(time.Duration(numberOfSeconds) * time.Second) // Wait before retrying
			refreshErr := Plex_RefreshItem(ratingKey)
			if refreshErr.Message != "" {
				logging.LOG.Trace(fmt.Sprintf("Failed to refresh Plex item: %v", refreshErr.Message))
			}
		} else {
			// Last attempt failed; log and return error
			logging.LOG.Error(fmt.Sprintf("All attempts to get posters failed. URL: %s", posterURL))
			logging.LOG.Error(fmt.Sprintf("Final error: %s", Err.Details))
			return "", Err
		}
	}

	// If all attempts fail, return the last error
	return "", Err
}

func Plex_SetPoster(ratingKey, posterKey, posterType string) logging.StandardError {
	// PUT Method is used when saving images locally
	// PUT Method requires the posterType to be singular (poster or art)
	//
	// POST Method is used when not using a local image
	// POST Method requires the posterType to be plural (posters or arts)

	var requestMethod string
	if strings.HasPrefix(posterKey, "metadata://") {
		// Local asset, use PUT method
		requestMethod = "PUT"
		if posterType == "backdrop" {
			posterType = "art"
		} else {
			posterType = "poster"
		}
	} else if strings.HasPrefix(posterKey, "http://") || strings.HasPrefix(posterKey, "https://") {
		// Remote URL, use POST method
		requestMethod = "POST"
		if posterType == "backdrop" {
			posterType = "arts"
		} else {
			posterType = "posters"
		}
	} else {
		Err := logging.NewStandardError()
		Err.Message = "Invalid poster key format"
		Err.HelpText = "The poster key must start with 'metadata://', 'http://', or 'https://'."
		Err.Details = fmt.Sprintf("Received invalid poster key: %s", posterKey)
		return Err
	}

	// Use net/url to escape the rating key
	escapedPosterKey := url.QueryEscape(posterKey)

	// Construct the URL for setting the poster
	url := fmt.Sprintf("%s/library/metadata/%s/%s?url=%s", Global_Config.MediaServer.URL, ratingKey, posterType, escapedPosterKey)

	logging.LOG.Trace(fmt.Sprintf("Setting %s for rating key '%s' using '%s' method\nURL: %s", posterType, ratingKey, requestMethod, url))

	response, body, Err := MakeHTTPRequest(url, requestMethod, nil, 60, nil, "MediaServer")
	if Err.Message != "" {
		return Err
	}
	defer response.Body.Close()

	if !strings.HasPrefix(string(body), "/library/metadata/") {
		Err.Message = "Failed to set poster"
		Err.HelpText = fmt.Sprintf("Ensure the item with rating key %s exists. If it does, check the Plex server logs for more information. If it doesn't, please try refreshing aura from the Home page.", ratingKey)
		Err.Details = fmt.Sprintf("Received response: %s", string(body))
		return Err
	}

	return logging.StandardError{}
}

func Plex_HandleLabels(item MediaItem) logging.StandardError {

	if item.Type != "movie" && item.Type != "show" {
		return logging.StandardError{}
	}

	// Get all of the applications configured for labels and tags
	for _, app := range Global_Config.LabelsAndTags.Applications {
		if app.Application != "Plex" {
			continue
		}
		// Only proceed if the application is enabled
		if app.Application == "Plex" && app.Enabled {

			// Check to see there at least one label to add or remove
			if len(app.Add) == 0 && len(app.Remove) == 0 {
				return logging.StandardError{}
			}

			Err := logging.NewStandardError()

			sectionName := item.LibraryTitle
			if sectionName == "" {
				Err.Message = "Library title is empty"
				Err.HelpText = "Ensure the item exists in Plex and has a valid library title."
				Err.Details = fmt.Sprintf("No library title found for item '%s'", item.Title)
				logging.LOG.Error(Err.Message)
				return Err
			}

			// Get the section ID from the library title
			sectionID := ""
			librarySection, found := Global_Cache_LibraryStore.GetSectionByTitle(item.LibraryTitle)
			if !found {
				Err.Message = "Library section not found in cache"
				Err.HelpText = fmt.Sprintf("Ensure the library '%s' exists in Plex and is correctly configured.", sectionName)
				Err.Details = fmt.Sprintf("Library section '%s' not found in cache", sectionName)
				logging.LOG.Error(Err.Message)
				return Err
			} else {
				sectionID = librarySection.ID
				logging.LOG.Trace(fmt.Sprintf("Found library section ID '%s' for library '%s' in cache", sectionID, sectionName))
			}

			if sectionID == "" {
				Err.Message = "Library section ID not found"
				Err.HelpText = fmt.Sprintf("Ensure the library '%s' exists in Plex and is correctly configured.", sectionName)
				Err.Details = fmt.Sprintf("No section ID found for library '%s'", sectionName)
				logging.LOG.Error(Err.Message)
				return Err
			}

			// Determine the Type Number based on item.Type
			var typeNumber int
			switch item.Type {
			case "movie":
				typeNumber = 1
			case "show":
				typeNumber = 2
			default:
				Err.Message = "Unsupported item type for label removal"
				Err.HelpText = fmt.Sprintf("Label removal is only supported for 'movie' and 'show' item types. Found type: '%s'", item.Type)
				Err.Details = fmt.Sprintf("Unsupported item type '%s' for item '%s'", item.Type, item.Title)
				logging.LOG.Error(Err.Message)
				return Err
			}

			// Make a comma-separated string of labels to remove
			labelsToRemove := ""
			if len(app.Remove) > 0 {
				labelsToRemove = strings.Join(app.Remove, ",")
			}

			// %5B = [
			// %5D = ]
			// Construct the removal parameter for the URL
			// Structure: label%5B%5D.tag.tag-={label1},{label2}
			// Example: label%5B%5D.tag.tag-=Overlay,4K
			removalParam := ""
			if labelsToRemove != "" {
				removalParam = fmt.Sprintf("label%%5B%%5D.tag.tag-=%s", url.QueryEscape(labelsToRemove))
				logging.LOG.Trace(fmt.Sprintf("Removing label(s) '%s' from: %s (%s)", labelsToRemove, item.Title, item.LibraryTitle))
			}

			// Construct the addition parameter for the URL
			// Structure: label%5B{index}%5D.tag.tag={label1},{label2}
			// Example: label%5B0%5D.tag.tag=Overlay&label%5B0%5D.tag.tag=4K
			// Note: The index should start at 0 and increment for each label to add
			labelsToAdd := ""
			additionParams := ""
			if len(app.Add) > 0 {
				for index, label := range app.Add {
					if index > 0 {
						additionParams += "&"
					}
					additionParams += fmt.Sprintf("label%%5B%d%%5D.tag.tag=%s", index, url.QueryEscape(label))
					labelsToAdd += label
					if index < len(app.Add)-1 {
						labelsToAdd += ","
					}
				}
			}
			if labelsToAdd != "" {
				logging.LOG.Trace(fmt.Sprintf("Adding label(s) '%s' to: %s (%s)", labelsToAdd, item.Title, item.LibraryTitle))
			}

			// If no labels to add or remove, return early
			if removalParam == "" && additionParams == "" {
				return logging.StandardError{}
			}

			// Combine removal and addition parameters
			var combinedParams string
			if labelsToRemove != "" && additionParams != "" {
				combinedParams = fmt.Sprintf("%s&%s", removalParam, additionParams)
			} else if labelsToRemove != "" {
				combinedParams = fmt.Sprintf("label%%5B%%5D.tag.tag-=%s", url.QueryEscape(labelsToRemove))
			} else if additionParams != "" {
				combinedParams = additionParams
			} else {
				// No labels to add or remove
				return logging.StandardError{}
			}

			// Construct the request URL using the removal parameter
			requestUrl := fmt.Sprintf("%s/library/sections/%s/all?type=%d&id=%s&%s",
				Global_Config.MediaServer.URL, sectionID, typeNumber, item.RatingKey, combinedParams)

			// Send the request via PUT
			response, _, Err := MakeHTTPRequest(requestUrl, "PUT", nil, 60, nil, "MediaServer")
			if Err.Message != "" {
				return Err
			}
			defer response.Body.Close()

			logging.LOG.Trace(fmt.Sprintf("Label removal response: %s", response.Status))
		}
	}
	return logging.StandardError{}
}
