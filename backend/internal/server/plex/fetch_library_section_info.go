package plex

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"poster-setter/internal/config"
	"poster-setter/internal/logging"
	"poster-setter/internal/modals"
	"poster-setter/internal/utils"
)

func FetchLibrarySectionInfo(library *modals.Config_MediaServerLibrary) (bool, logging.ErrorLog) {

	// Construct the URL for the Plex server API request
	url, logErr := utils.MakeMediaServerAPIURL("library/sections", config.Global.MediaServer.URL)
	if logErr.Err != nil {
		return false, logErr
	}
	// Make a GET request to the Plex server
	response, body, logErr := utils.MakeHTTPRequest(url.String(), http.MethodGet, nil, 60, nil, "MediaServer")
	if logErr.Err != nil {
		logging.LOG.Error(logErr.Log.Message)
		return false, logErr
	}
	defer response.Body.Close()

	// Check if the response status is OK
	if response.StatusCode != http.StatusOK {
		logging.LOG.Error(fmt.Sprintf("Received status code '%d' from Plex server", response.StatusCode))
		return false, logging.ErrorLog{Err: fmt.Errorf("received status code '%d' from Plex server", response.StatusCode),
			Log: logging.Log{Message: fmt.Sprintf("Received status code '%d' from Plex server", response.StatusCode)}}
	}

	var responseSection modals.PlexResponse
	err := xml.Unmarshal(body, &responseSection)
	if err != nil {
		logging.LOG.Error("Failed to parse XML response")
		return false, logging.ErrorLog{Err: err,
			Log: logging.Log{Message: "Failed to parse XML response"},
		}
	}

	// Find the library section with the matching Name
	for _, section := range responseSection.Directory {
		if section.Title == library.Name {
			library.Type = section.Type
			library.SectionID = section.Key
			break
		}
	}
	if library.SectionID == "" {
		logging.LOG.Error(fmt.Sprintf("Library section '%s' not found", library.Name))
		return false, logging.ErrorLog{Err: fmt.Errorf("library section '%s' not found", library.Name),
			Log: logging.Log{Message: fmt.Sprintf("Library section '%s' not found", library.Name)}}
	}

	return true, logging.ErrorLog{}
}
