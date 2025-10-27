package api

import (
	"aura/internal/logging"
	"encoding/json"
	"fmt"
	"net/url"
	"path"
)

func (s *SonarrApp) GetItemInfoFromTMDBID(app Config_SonarrRadarrApp, tmdbID int) (any, logging.StandardError) {
	return SR_GetItemInfoFromTMDBID(app, tmdbID)
}

func (r *RadarrApp) GetItemInfoFromTMDBID(app Config_SonarrRadarrApp, tmdbID int) (any, logging.StandardError) {
	return SR_GetItemInfoFromTMDBID(app, tmdbID)
}

//////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////

func SR_CallGetItemInfoFromTMDBID(app Config_SonarrRadarrApp, tmdbID int) (any, logging.StandardError) {
	// Make sure all of the required information is set
	Err := SR_MakeSureAllInfoIsSet(app)
	if Err.Message != "" {
		return 0, Err
	}

	// Get the appropriate interface
	interfaceSR, Err := SR_GetSonarrRadarrInterface(app)
	if Err.Message != "" {
		return 0, Err
	}

	return interfaceSR.GetItemInfoFromTMDBID(app, tmdbID)
}

//////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////

func SR_GetItemInfoFromTMDBID(app Config_SonarrRadarrApp, tmdbID int) (any, logging.StandardError) {
	Err := logging.NewStandardError()

	urlEndPoint := ""
	switch app.Type {
	case "Sonarr":
		urlEndPoint = "series"
	case "Radarr":
		urlEndPoint = "movie"
	default:
		Err.Message = "Unsupported Sonarr/Radarr Type"
		Err.HelpText = "Ensure the Sonarr/Radarr Type is set to either 'Sonarr' or 'Radarr' in the configuration."
		return nil, Err
	}
	logging.LOG.Debug(fmt.Sprintf("Getting item info from %s (%s) for TMDB ID: %d", app.Type, app.Library, tmdbID))
	// Construct the URL to search for the series by TMDB ID
	u, _ := url.Parse(app.URL)
	u.Path = path.Join(u.Path, "api/v3", urlEndPoint, "lookup")
	query := u.Query()
	query.Set("term", fmt.Sprintf("tmdb:%d", tmdbID))
	u.RawQuery = query.Encode()
	tmdbURL := u.String()

	apiHeader := map[string]string{
		"X-Api-Key": app.APIKey,
	}

	response, body, Err := MakeHTTPRequest(tmdbURL, "GET", apiHeader, 60, nil, app.Type)
	if Err.Message != "" {
		return nil, Err
	}
	defer response.Body.Close()

	// Check to see if Status is OK
	if response.StatusCode != 200 {
		Err.Message = fmt.Sprintf("%s returned non-200 status code: %d", app.Type, response.StatusCode)
		Err.HelpText = fmt.Sprintf("Check the %s URL and API key in the configuration.", app.Type)
		Err.Details = map[string]any{
			"status_code":   response.StatusCode,
			"response_body": string(body),
			"app":           app,
		}
		return nil, Err
	}

	// Parse the response body
	var results []struct {
		ID         int `json:"id"`
		SR_TMDB_ID int `json:"tmdbId"`
	}
	err := json.Unmarshal(body, &results)
	if err != nil {
		Err.Message = fmt.Sprintf("Failed to parse %s response", app.Type)
		Err.HelpText = fmt.Sprintf("Ensure the %s instance is running the correct version and is accessible.", app.Type)
		Err.Details = map[string]any{
			"error":         err.Error(),
			"response_body": string(body),
		}
		return nil, Err
	}
	result := results[0]

	// Make sure that the TMDB ID Matches
	if result.SR_TMDB_ID != tmdbID {
		Err.Message = "TMDB ID mismatch"
		Err.HelpText = fmt.Sprintf("The TMDB ID returned by %s does not match the requested TMDB ID.", app.Type)
		Err.Details = map[string]any{
			"requested_tmdb_id": tmdbID,
			"returned_tmdb_id":  result.SR_TMDB_ID,
		}
		return nil, Err
	}

	// If ID is not there, this means the item is not in Sonarr/Radarr
	if result.ID == 0 {
		Err.Message = fmt.Sprintf("Item with TMDB ID '%d' not found in %s", tmdbID, app.Type)
		Err.HelpText = fmt.Sprintf("The item with the specified TMDB ID does not exist in %s.", app.Type)
		return nil, Err
	}

	logging.LOG.Trace(fmt.Sprintf("Found %s item with ID: %d for TMDB ID: %d", app.Type, result.ID, tmdbID))

	// Now that we have the item, get the full item information
	itemURL := ""
	u, _ = url.Parse(app.URL)
	u.Path = path.Join(u.Path, "api/v3", urlEndPoint, fmt.Sprintf("%d", result.ID))
	itemURL = u.String()

	response, body, Err = MakeHTTPRequest(itemURL, "GET", apiHeader, 60, nil, app.Type)
	if Err.Message != "" {
		return nil, Err
	}
	defer response.Body.Close()

	// Check to see if Status is OK
	if response.StatusCode != 200 {
		Err.Message = fmt.Sprintf("%s returned non-200 status code when fetching item: %d", app.Type, response.StatusCode)
		Err.HelpText = fmt.Sprintf("Check the %s URL and API key in the configuration.", app.Type)
		Err.Details = map[string]any{
			"status_code":   response.StatusCode,
			"response_body": string(body),
			"app":           app,
		}
		return nil, Err
	}

	// Parse the full item response body
	switch app.Type {
	case "Sonarr":
		var item SR_SonarrItem
		err = json.Unmarshal(body, &item)
		if err != nil {
			Err.Message = "Failed to parse Sonarr item response"
			Err.HelpText = "Ensure the Sonarr instance is running the correct version and is accessible."
			Err.Details = map[string]any{
				"error":         err.Error(),
				"response_body": string(body),
			}
			logging.LOG.Error(fmt.Sprintf("Error details: %+v", Err.Details))
			return nil, Err
		}
		return item, logging.StandardError{}
	case "Radarr":
		var item SR_RadarrItem
		err = json.Unmarshal(body, &item)
		if err != nil {
			Err.Message = "Failed to parse Radarr item response"
			Err.HelpText = "Ensure the Radarr instance is running the correct version and is accessible."
			Err.Details = map[string]any{
				"error":         err.Error(),
				"response_body": string(body),
			}
			logging.LOG.Error(fmt.Sprintf("Error details: %+v", Err.Details))
			return nil, Err
		}
		return item, logging.StandardError{}
	}

	// If we reach here, something went wrong
	Err.Message = "Unsupported Sonarr/Radarr Type"
	Err.HelpText = "Ensure the Sonarr/Radarr Type is set to either 'Sonarr' or 'Radarr' in the configuration."
	Err.Details = map[string]any{
		"app": app,
	}
	return SR_SonarrItem{}, logging.StandardError{}
}
