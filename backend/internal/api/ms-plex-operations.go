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

// Plex_GetAllImages retrieves all images for a Plex item by ratingKey.
func Plex_GetAllImages(ctx context.Context, itemRatingKey string, imageType string) ([]PlexGetAllImagesMetadata, logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, fmt.Sprintf("Getting all '%s' for Item with Rating Key '%s' from Plex", imageType, itemRatingKey), logging.LevelDebug)
	defer logAction.Complete()

	var plexImages []PlexGetAllImagesMetadata

	// Adjust imageType for URL construction
	if imageType == "backdrop" {
		imageType = "arts"
	} else {
		imageType = "posters"
	}

	// Make the URL
	u, err := url.Parse(Global_Config.MediaServer.URL)
	if err != nil {
		logAction.SetError("Failed to parse Plex base URL", err.Error(), nil)
		return plexImages, *logAction.Error
	}
	u.Path = path.Join(u.Path, "library", "metadata", itemRatingKey, imageType)
	URL := u.String()

	// Make Auth Headers for Request
	headers := MakeAuthHeader("X-Plex-Token", Global_Config.MediaServer.Token)

	// Make the API request to Plex
	httpResp, body, logErr := MakeHTTPRequest(ctx, URL, http.MethodGet, headers, 60, nil, "Plex")
	if logErr.Message != "" {
		return plexImages, logErr
	}
	defer httpResp.Body.Close()

	// Check the response status code
	if httpResp.StatusCode != http.StatusOK {
		logAction.SetError("Failed to get images from Plex",
			fmt.Sprintf("Plex server returned status code %d", httpResp.StatusCode),
			map[string]any{"URL": URL, "StatusCode": httpResp.StatusCode})
		return plexImages, *logAction.Error
	}

	// Parse the response body into a PlexGetAllImagesWrapper struct
	var plexImagesWrapper PlexGetAllImagesWrapper
	logErr = DecodeJSONBody(ctx, body, &plexImagesWrapper, "PlexGetAllImagesWrapper")
	if logErr.Message != "" {
		return plexImages, logErr
	}

	plexImages = plexImagesWrapper.MediaContainer.Metadata

	// Remove any images that do not have a provider of "local"
	// If there are no local images, return an empty slice
	localImages := []PlexGetAllImagesMetadata{}
	for _, image := range plexImages {
		if image.Provider == "local" {
			localImages = append(localImages, image)
		}
	}

	return localImages, logging.LogErrorInfo{}
}

// Plex_GetNewImage attempts to retrieve a new image for a Plex item by comparing previous and current images.
// It tries up to 3 times, refreshing the Plex item between attempts if needed.
// Returns the new image's ratingKey (URL or path) if found, or a StandardError if not.
func Plex_GetNewImage(ctx context.Context, itemRatingKey string, imageType string, previousImages []PlexGetAllImagesMetadata) (string, logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, fmt.Sprintf("Getting new '%s' for Item with Rating Key '%s' from Plex", imageType, itemRatingKey), logging.LevelDebug)
	defer logAction.Complete()

	var lastErrorMsg string
	var lastErrorDetail map[string]any

	// Retry logic for the entire process (up to 3 attempts)
	for attempt := 1; attempt <= 3; attempt++ {
		attemptAction := logAction.AddSubAction(fmt.Sprintf("Attempt %d to fetch new image", attempt), logging.LevelDebug)
		// Get all current images from Plex
		currentImages, logErr := Plex_GetAllImages(ctx, itemRatingKey, imageType)
		if logErr.Message != "" {
			attemptAction.AppendWarning(fmt.Sprintf("attempt_%d", attempt), map[string]any{"error": logErr.Message})
			lastErrorMsg = logErr.Message
			lastErrorDetail = logErr.Detail
		} else {
			// Compare current images with previous images to find a new one
			for _, currentImage := range currentImages {
				if currentImage.Provider != "local" {
					continue
				}
				isNew := true
				for _, previousImage := range previousImages {
					if currentImage.RatingKey == previousImage.RatingKey {
						isNew = false
						break
					}
				}
				if isNew {
					attemptAction.AppendResult(fmt.Sprintf("attempt_%d", attempt), map[string]any{"new_image_rating_key": currentImage.RatingKey})
					return currentImage.RatingKey, logging.LogErrorInfo{}
				}
			}
			attemptAction.AppendWarning(fmt.Sprintf("attempt_%d", attempt), map[string]any{
				"error": "No new image found; all images match previous ones.",
			})
			lastErrorMsg = "No new image found; all images match previous ones."
			lastErrorDetail = nil
		}
		// If we reach here, the attempt failed; refresh the item before retrying
		if attempt < 3 {
			numberOfSeconds := 1
			time.Sleep(time.Duration(numberOfSeconds) * time.Second) // Wait before retrying
			refreshErr := Plex_RefreshItem(ctx, itemRatingKey)
			if refreshErr.Message != "" {
				attemptAction.AppendWarning("refresh_error", map[string]any{"error": refreshErr.Message})
				lastErrorMsg = refreshErr.Message
				lastErrorDetail = refreshErr.Detail
			}
		}
	}

	// All attempts failed, set error only now
	finalError := logging.LogErrorInfo{Message: lastErrorMsg, Detail: lastErrorDetail}
	return "", finalError
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

func Plex_HandleLabels(item MediaItem, selectedTypes []string) {
	// If the Global Config Media Server is not Plex, exit
	if Global_Config.MediaServer.Type != "Plex" {
		return
	}

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
				if app.AddLabelTagForSelectedTypes && len(selectedTypes) > 0 {
					for _, label := range selectedTypes {
						additionParams += "&"
						switch label {
						case "poster":
							label = "aura-poster"
						case "backdrop":
							label = "aura-backdrop"
						case "seasonPoster":
							label = "aura-season-poster"
						case "specialSeasonPoster":
							label = "aura-special-season-poster"
						case "titlecard":
							label = "aura-titlecard"
						}
						additionParams += fmt.Sprintf("label%%5B%d%%5D.tag.tag=%s", len(app.Add), url.QueryEscape(label))
						labelsToAdd += "," + label
					}
				}
			}
			if labelsToAdd != "" {
				subAppAction.AppendResult("labels_to_add", strings.Split(labelsToAdd, ","))
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
