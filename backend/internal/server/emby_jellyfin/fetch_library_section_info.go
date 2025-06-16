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

func FetchLibrarySectionInfo(library *modals.Config_MediaServerLibrary) (bool, logging.StandardError) {

	url, Err := utils.MakeMediaServerAPIURL(fmt.Sprintf("/Users/%s/Items", config.Global.MediaServer.UserID), config.Global.MediaServer.URL)
	if Err.Message != "" {
		return false, Err
	}

	// Make a GET request to the Emby server
	response, body, Err := utils.MakeHTTPRequest(url.String(), http.MethodGet, nil, 60, nil, "MediaServer")
	if Err.Message != "" {
		logging.LOG.Error(Err.Message)
		return false, Err
	}
	defer response.Body.Close()

	Err = logging.NewStandardError()

	var responseSection modals.EmbyJellyLibrarySectionsResponse
	err := json.Unmarshal(body, &responseSection)
	if err != nil {

		Err.Message = "Failed to parse JSON response"
		Err.HelpText = "Ensure the Emby/Jellyfin server is returning a valid JSON response."
		Err.Details = fmt.Sprintf("Error: %s", err.Error())
		return false, Err
	}

	found := false
	for _, item := range responseSection.Items {
		if item.Name == library.Name {
			library.Type = map[string]string{
				"movies":  "movie",
				"tvshows": "show",
			}[item.CollectionType]
			library.SectionID = item.ID
			found = true
			break
		}
	}

	if !found {

		Err.Message = "Library section not found"
		Err.HelpText = fmt.Sprintf("Ensure the library section '%s' exists on the media server.", library.Name)
		Err.Details = fmt.Sprintf("No section with name '%s' found in the media server response.", library.Name)
		return false, Err
	}

	return true, logging.StandardError{}
}
