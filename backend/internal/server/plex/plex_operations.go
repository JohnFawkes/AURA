package plex

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"poster-setter/internal/config"
	"poster-setter/internal/logging"
	"poster-setter/internal/modals"
	"poster-setter/internal/utils"
)

func refreshPlexItem(ratingKey string) logging.ErrorLog {
	logging.LOG.Trace(fmt.Sprintf("Refreshing Plex item with rating key: %s", ratingKey))

	url := fmt.Sprintf("%s/library/metadata/%s/refresh", config.Global.MediaServer.URL, ratingKey)
	response, _, logErr := utils.MakeHTTPRequest(url, "PUT", nil, 60, nil, "MediaServer")
	if logErr.Err != nil {
		return logErr
	}

	// Check if the response is successful
	if response.StatusCode != http.StatusOK {
		return logging.ErrorLog{Err: fmt.Errorf("received status code '%d' from Plex server", response.StatusCode),
			Log: logging.Log{Message: fmt.Sprintf("Failed to refresh Plex item with rating key: %s", ratingKey)},
		}
	}

	return logging.ErrorLog{}
}

func getPosters(ratingKey string) (string, logging.ErrorLog) {
	logging.LOG.Trace(fmt.Sprintf("Getting posters for rating key: %s", ratingKey))
	posterURL := fmt.Sprintf("%s/library/metadata/%s/posters", config.Global.MediaServer.URL, ratingKey)
	response, body, logErr := utils.MakeHTTPRequest(posterURL, "GET", nil, 60, nil, "MediaServer")
	if logErr.Err != nil {
		return "", logErr
	}
	defer response.Body.Close()

	logging.LOG.Trace(fmt.Sprintf("Response from Plex server: %s", string(body)))

	// Check if the response is successful
	if response.StatusCode != http.StatusOK {
		return "", logging.ErrorLog{Err: fmt.Errorf("received status code '%d' from Plex server", response.StatusCode),
			Log: logging.Log{Message: fmt.Sprintf("Failed to get posters for rating key: %s", ratingKey)},
		}
	}

	// Parse the response body into a PlexPhotosResponse struct
	var plexPosters modals.PlexPhotosResponse
	err := xml.Unmarshal(body, &plexPosters)
	if err != nil {
		return "", logging.ErrorLog{Err: err,
			Log: logging.Log{Message: "Failed to parse XML response"},
		}
	}

	// Check if the response contains any posters
	if len(plexPosters.Photos) == 0 {
		return "", logging.ErrorLog{Err: fmt.Errorf("no posters found for rating key: %s", ratingKey),
			Log: logging.Log{Message: "No posters found for the item"},
		}
	}

	// Look for the first poster with a provider of "local"
	var selectedPoster modals.PlexPhotoItem
	for _, poster := range plexPosters.Photos {
		if poster.Provider == "local" {
			selectedPoster = poster
			break
		}
	}

	// Make sure that poster.RatingKey is not empty
	if selectedPoster.RatingKey == "" {
		return "", logging.ErrorLog{Err: fmt.Errorf("no local poster found for rating key: %s", ratingKey),
			Log: logging.Log{Message: fmt.Sprintf("no local poster found for rating key: %s", ratingKey)},
		}
	}

	logging.LOG.Trace(fmt.Sprintf("Poster RatingKey: %s", selectedPoster.RatingKey))
	return selectedPoster.RatingKey, logging.ErrorLog{}
}

func setPoster(ratingKey string, posterKey string, posterType string) logging.ErrorLog {
	// If the posterType is "backdrop", set the URL to art
	// Else set it to posters

	if posterType == "backdrop" {
		posterType = "arts"
	} else {
		posterType = "posters"
	}
	logging.LOG.Trace(fmt.Sprintf("Setting %s for rating key: %s", posterType, ratingKey))

	// Use net/url to escape the rating key
	escapedPosterKey := url.QueryEscape(posterKey)

	// Construct the URL for setting the poster
	url := fmt.Sprintf("%s/library/metadata/%s/%s?url=%s", config.Global.MediaServer.URL, ratingKey, posterType, escapedPosterKey)

	response, body, logErr := utils.MakeHTTPRequest(url, "PUT", nil, 60, nil, "MediaServer")
	if logErr.Err != nil {
		return logErr
	}
	defer response.Body.Close()

	logging.LOG.Trace(fmt.Sprintf("Response from Plex server: %s", string(body)))

	// Check if the response is successful
	if response.StatusCode != http.StatusOK {
		return logging.ErrorLog{Err: fmt.Errorf("received status code '%d' from Plex server", response.StatusCode),
			Log: logging.Log{Message: fmt.Sprintf("Failed to set poster for rating key: %s", ratingKey)},
		}
	}

	return logging.ErrorLog{}
}
