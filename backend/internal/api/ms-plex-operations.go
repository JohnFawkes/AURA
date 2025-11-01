package api

import (
	"aura/internal/logging"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"
)

// Plex_RefreshItem sends a request to the Plex server to refresh the specified item.
func Plex_RefreshItem(ctx context.Context, itemRatingKey string) logging.LogErrorInfo {
	ctx, logAction := logging.AddSubActionToContext(ctx, fmt.Sprintf("Refreshing Item with Rating Key '%s' on Plex", itemRatingKey), logging.LevelDebug)
	defer logAction.Complete()

	// Construct the URL for the Plex server API request
	u, err := url.Parse(Global_Config.MediaServer.URL)
	if err != nil {
		logAction.SetError("Failed to parse base URL", "Ensure the URL is valid", map[string]any{"error": err.Error()})
		return *logAction.Error
	}
	u.Path = path.Join(u.Path, "library", "metadata", itemRatingKey, "refresh")
	URL := u.String()

	// Make the Auth Headers for Request
	headers := MakeAuthHeader("X-Plex-Token", Global_Config.MediaServer.Token)

	// Make the API request to Plex
	httpResp, _, logErr := MakeHTTPRequest(ctx, URL, http.MethodPut, headers, 60, nil, "Plex")
	if logErr.Message != "" {
		return logErr
	}
	defer httpResp.Body.Close()

	// Check the response status code
	if httpResp.StatusCode != http.StatusOK && httpResp.StatusCode != http.StatusAccepted {
		logAction.SetError("Failed to refresh item on Plex",
			fmt.Sprintf("Plex server returned status code %d", httpResp.StatusCode),
			map[string]any{
				"URL":        URL,
				"StatusCode": httpResp.StatusCode,
			})
		return *logAction.Error
	}

	return logging.LogErrorInfo{}
}

// Plex_GetPoster attempts to retrieve the poster for a Plex item by ratingKey.
// It tries up to 3 times, refreshing the Plex item between attempts if needed.
// Returns the poster's ratingKey (URL or path) if found, or a StandardError if not.
func Plex_GetPoster(ctx context.Context, itemRatingKey string, posterType string) (string, logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, fmt.Sprintf("Getting Poster for Item with Rating Key '%s' from Plex", itemRatingKey), logging.LevelDebug)
	defer logAction.Complete()

	if posterType == "backdrop" {
		posterType = "arts"
	} else {
		posterType = "posters"
	}

	// Make the URL
	u, err := url.Parse(Global_Config.MediaServer.URL)
	if err != nil {
		logAction.SetError("Failed to parse Plex base URL", err.Error(), nil)
		return "", *logAction.Error
	}
	u.Path = path.Join(u.Path, "library", "metadata", itemRatingKey, posterType)
	URL := u.String()

	// Make Auth Headers for Request
	headers := MakeAuthHeader("X-Plex-Token", Global_Config.MediaServer.Token)

	// Retry logic for the entire process (up to 3 attempts)
	for attempt := 1; attempt <= 3; attempt++ {
		attemptAction := logAction.AddSubAction(fmt.Sprintf("Attempt %d to fetch poster", attempt), logging.LevelDebug)
		// Make the API request to Plex
		response, body, logErr := MakeHTTPRequest(ctx, URL, http.MethodGet, headers, 60, nil, "Plex")
		if logErr.Message != "" {
			attemptAction.SetError("Failed to make HTTP request to Plex", logErr.Message, nil)
		} else {
			defer response.Body.Close()

			// Check if the response status code is OK
			if response.StatusCode == http.StatusOK {

				// Parse the response body into a PlexGetAllImagesWrapper struct
				var plexPosters PlexGetAllImagesWrapper
				logErr = DecodeJSONBody(ctx, body, &plexPosters, "PlexGetAllImagesWrapper")
				if logErr.Message != "" {
					attemptAction.SetError("Failed to decode Plex response", logErr.Message, nil)
				} else {
					// Successfully parsed JSON; check for posters
					if len(plexPosters.MediaContainer.Metadata) > 0 {
						// Look for the first poster with a provider of "local"
						for _, poster := range plexPosters.MediaContainer.Metadata {
							if poster.Provider == "local" && poster.RatingKey != "" {
								return poster.RatingKey, logging.LogErrorInfo{}
							}
						}
						// No local poster found, but posters exist
						attemptAction.AppendWarning("attempt_outcome", "no local posters found")
						attemptAction.AppendWarning("available_posters", len(plexPosters.MediaContainer.Metadata))
					} else {
						// No posters found at all
						attemptAction.SetError("No posters found", "Plex did not return any posters for this item.", nil)
					}
				}
			} else {
				// Non-OK status code
				attemptAction.SetError("Failed to fetch posters from Plex",
					fmt.Sprintf("Plex server returned status code %d", response.StatusCode),
					map[string]any{
						"URL":        URL,
						"StatusCode": response.StatusCode,
					})
			}
		}

		// If we reach here, the attempt failed; refresh the item before retrying
		if attempt < 3 {
			numberOfSeconds := 2
			attemptAction.AppendWarning("outcome", "retrying after refresh")
			attemptAction.AppendWarning("wait_seconds", numberOfSeconds)
			attemptAction.AppendWarning("attempt", attempt)
			time.Sleep(time.Duration(numberOfSeconds) * time.Second) // Wait before retrying
			refreshErr := Plex_RefreshItem(ctx, itemRatingKey)
			if refreshErr.Message != "" {
				attemptAction.SetError("Failed to refresh item on Plex", refreshErr.Message, nil)
			}
		} else {
			attemptAction.AppendResult("final_outcome", "all attempts failed")
			if attemptAction.Error != nil {
				return "", *attemptAction.Error
			}
			return "", logging.LogErrorInfo{
				Message: "Failed to retrieve poster after 3 attempts",
			}
		}
	}

	// If we reach here, all attempts failed
	return "", logging.LogErrorInfo{
		Message: "Failed to retrieve poster after multiple attempts",
	}
}

func Plex_SetPoster(ctx context.Context, itemRatingKey string, posterKey string, posterType string) logging.LogErrorInfo {
	// PUT Method is used when saving images locally
	// PUT Method requires the posterType to be singular (poster or art)
	//
	// POST Method is used when not using a local image
	// POST Method requires the posterType to be plural (posters or arts)

	ctx, logAction := logging.AddSubActionToContext(ctx, fmt.Sprintf("Setting Poster for Item with Rating Key '%s' on Plex", itemRatingKey), logging.LevelDebug)
	defer logAction.Complete()

	var requestMethod string
	if strings.HasPrefix(posterKey, "metadata://") {
		// Local asset, use PUT method
		requestMethod = http.MethodPut
		if posterType == "backdrop" {
			posterType = "art"
		} else {
			posterType = "poster"
		}
	} else if strings.HasPrefix(posterKey, "http://") || strings.HasPrefix(posterKey, "https://") {
		// Remote URL, use POST method
		requestMethod = http.MethodPost
		if posterType == "backdrop" {
			posterType = "arts"
		} else {
			posterType = "posters"
		}
	} else {
		logAction.SetError("Invalid poster key format",
			"Poster key must start with 'metadata://' for local assets or 'http(s)://' for remote URLs.",
			map[string]any{
				"PosterKey": posterKey,
			})
		return *logAction.Error
	}

	// Construct the URL for the Plex server API request
	u, err := url.Parse(Global_Config.MediaServer.URL)
	if err != nil {
		logAction.SetError("Failed to parse Plex base URL", err.Error(), nil)
		return *logAction.Error
	}
	u.Path = path.Join(u.Path, "library", "metadata", itemRatingKey, posterType)
	query := u.Query()
	query.Set("url", posterKey)
	u.RawQuery = query.Encode()
	URL := u.String()

	// Make the Auth Headers for Request
	headers := MakeAuthHeader("X-Plex-Token", Global_Config.MediaServer.Token)

	// Make the API request to Plex
	httpResp, respBody, logErr := MakeHTTPRequest(ctx, URL, requestMethod, headers, 60, nil, "Plex")
	if logErr.Message != "" {
		return logErr
	}
	defer httpResp.Body.Close()

	switch requestMethod {
	case http.MethodPut:
		if !strings.HasPrefix(string(respBody), "/library/metadata/") {
			logAction.SetError("Failed to set poster on Plex",
				"Plex did not return a valid metadata path after setting the poster.",
				map[string]any{
					"URL": URL,
				})
			return *logAction.Error
		}
	case http.MethodPost:
		// Check the response status code for POST method
		if httpResp.StatusCode != http.StatusOK {
			logAction.SetError("Failed to set poster on Plex",
				"Plex did not return a valid metadata path after setting the poster.",
				map[string]any{
					"URL": URL,
				})
			return *logAction.Error
		}
	}

	return logging.LogErrorInfo{}

}

func Plex_HandleLabels(item MediaItem) {
	// If there is no applications in the Global Config, exit
	if len(Global_Config.LabelsAndTags.Applications) == 0 {
		return
	}

	ctx, ld := logging.CreateLoggingContext(context.Background(), "Plex - Handle Labels")
	logAction := ld.AddAction("Handle Labels in Plex", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)
	defer ld.Log()
	defer logAction.Complete()

	// If Item Type is not movie or show, exit
	if item.Type != "movie" && item.Type != "show" {
		logAction.AppendWarning("outcome", "skipped")
		logAction.AppendWarning("reason", "unsupported_item_type")
		return
	}

	// Get all of the applications configured for labels and tags
	for _, app := range Global_Config.LabelsAndTags.Applications {
		if app.Application != "Plex" {
			continue
		}
		// Only proceed if the application is enabled
		if app.Application == "Plex" && app.Enabled {
			ctx, subAppAction := logging.AddSubActionToContext(ctx, "Processing Plex Labels", logging.LevelDebug)
			defer subAppAction.Complete()

			// Check to see there at least one label to add or remove
			if len(app.Add) == 0 && len(app.Remove) == 0 {
				subAppAction.AppendWarning("outcome", "skipped")
				subAppAction.AppendWarning("reason", "no_labels_to_add_or_remove")
				continue
			}

			sectionName := item.LibraryTitle
			if sectionName == "" {
				subAppAction.SetError("Library Section Title is empty",
					"Cannot handle labels without a valid Library Section Title.",
					map[string]any{
						"Item": item,
					})
				continue
			}

			// Get the section ID from the library title
			sectionID := ""
			librarySection, found := Global_Cache_LibraryStore.GetSectionByTitle(item.LibraryTitle)
			if !found {
				subAppAction.SetError("Library Section not found in cache",
					"Cannot handle labels without a valid Library Section in cache.",
					map[string]any{
						"LibraryTitle": item.LibraryTitle,
					})
				continue
			} else {
				sectionID = librarySection.ID
			}

			if sectionID == "" {
				subAppAction.SetError("Library Section ID is empty",
					"Cannot handle labels without a valid Library Section ID.",
					map[string]any{
						"LibraryTitle": item.LibraryTitle,
					})
				continue
			}

			// Determine the Type Number based on item.Type
			var typeNumber int
			switch item.Type {
			case "movie":
				typeNumber = 1
			case "show":
				typeNumber = 2
			default:
				subAppAction.SetError("Unsupported item type",
					"Cannot handle labels for unsupported item types.",
					map[string]any{
						"ItemType": item.Type,
					})
				continue
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
				subAppAction.AppendResult("labels_to_remove", app.Remove)
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
				subAppAction.AppendResult("labels_to_add", app.Add)
			}

			// If no labels to add or remove, return early
			if removalParam == "" && additionParams == "" {
				subAppAction.AppendWarning("outcome", "skipped")
				subAppAction.AppendWarning("reason", "no labels to add or remove after processing")
				continue
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
				subAppAction.AppendWarning("outcome", "skipped")
				subAppAction.AppendWarning("reason", "no labels to add or remove after processing")
				continue
			}

			// Construct the URL for the Plex server API request
			u, err := url.Parse(Global_Config.MediaServer.URL)
			if err != nil {
				subAppAction.SetError("Failed to parse base URL", "Ensure the URL is valid", map[string]any{"error": err.Error()})
				continue
			}
			u.Path = path.Join(u.Path, "library", "sections", sectionID, "all")
			query := u.Query()
			query.Set("type", strconv.Itoa(typeNumber))
			query.Set("id", item.RatingKey)
			// Encode the query parameters
			u.RawQuery = query.Encode()
			URL := u.String()
			URL += "&" + combinedParams

			// Make the Auth Headers for Request
			headers := MakeAuthHeader("X-Plex-Token", Global_Config.MediaServer.Token)

			// Make the API request to Plex
			httpResp, _, logErr := MakeHTTPRequest(ctx, URL, http.MethodPut, headers, 60, nil, "Plex")
			if logErr.Message != "" {
				subAppAction.SetError("Failed to make HTTP request to Plex", logErr.Message, nil)
				continue
			}
			defer httpResp.Body.Close()

			// Check the response status code
			if httpResp.StatusCode != http.StatusOK {
				subAppAction.SetError("Failed to update labels on Plex",
					fmt.Sprintf("Plex server returned status code %d", httpResp.StatusCode),
					map[string]any{
						"URL":        URL,
						"StatusCode": httpResp.StatusCode,
					})
				continue
			}
		}
	}
}
