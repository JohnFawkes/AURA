package emby_jellyfin

import (
	"fmt"
	"net/http"
	"poster-setter/internal/config"
	"poster-setter/internal/logging"
	"poster-setter/internal/modals"
	"poster-setter/internal/utils"

	"github.com/goccy/go-json"
)

func FetchLibrarySectionInfo(library *modals.Config_MediaServerLibrary) (bool, logging.ErrorLog) {

	url, logErr := utils.MakeMediaServerAPIURL(fmt.Sprintf("/Users/%s/Items", config.Global.MediaServer.UserID), config.Global.MediaServer.URL)
	if logErr.Err != nil {
		return false, logErr
	}

	// Make a GET request to the Emby server
	response, body, logErr := utils.MakeHTTPRequest(url.String(), http.MethodGet, nil, 60, nil, "MediaServer")
	if logErr.Err != nil {
		logging.LOG.Error(logErr.Log.Message)
		return false, logErr
	}
	defer response.Body.Close()

	// Check if the response status is OK
	if response.StatusCode != http.StatusOK {
		return false, logging.ErrorLog{Err: fmt.Errorf("received status code '%d' from %s server", response.StatusCode, config.Global.MediaServer.Type),
			Log: logging.Log{Message: fmt.Sprintf("Received status code '%d' from %s server", response.StatusCode, config.Global.MediaServer.Type)}}
	}

	var responseSection modals.EmbyJellyLibrarySectionsResponse
	err := json.Unmarshal(body, &responseSection)
	if err != nil {
		logging.LOG.Error("Failed to parse JSON response")
		return false, logging.ErrorLog{Err: err, Log: logging.Log{Message: "Failed to parse JSON response"}}
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
		return false, logging.ErrorLog{Err: fmt.Errorf("library section '%s' not found", library.Name),
			Log: logging.Log{Message: fmt.Sprintf("Library section '%s' not found", library.Name)}}
	}

	return true, logging.ErrorLog{}
}
