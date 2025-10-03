package emby_jellyfin

import (
	"aura/internal/config"
	"aura/internal/logging"
	"aura/internal/modals"
	"aura/internal/utils"
	"fmt"
	"net/http"

	"github.com/goccy/go-json"
)

func FetchLibrarySectionOptions() ([]string, logging.StandardError) {

	url, Err := utils.MakeMediaServerAPIURL(fmt.Sprintf("/Users/%s/Items", config.Global.MediaServer.UserID), config.Global.MediaServer.URL)
	if Err.Message != "" {
		return nil, Err
	}

	// Make a GET request to the Emby server
	response, body, Err := utils.MakeHTTPRequest(url.String(), http.MethodGet, nil, 60, nil, "MediaServer")
	if Err.Message != "" {
		logging.LOG.Error(Err.Message)
		return nil, Err
	}
	defer response.Body.Close()

	Err = logging.NewStandardError()

	var responseSection modals.EmbyJellyLibrarySectionsResponse
	err := json.Unmarshal(body, &responseSection)
	if err != nil {
		Err.Message = "Failed to parse JSON response"
		Err.HelpText = "Ensure the Emby/Jellyfin server is returning a valid JSON response."
		Err.Details = map[string]any{
			"error": err.Error(),
		}
		return nil, Err
	}

	var options []string
	for _, item := range responseSection.Items {
		if item.CollectionType != "tvshows" && item.CollectionType != "movies" {
			continue
		}
		options = append(options, item.Name)
	}

	return options, logging.StandardError{}
}
