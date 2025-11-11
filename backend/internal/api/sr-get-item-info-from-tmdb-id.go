package api

import (
	"aura/internal/logging"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
)

func (s *SonarrApp) GetItemInfoFromTMDBID(ctx context.Context, app Config_SonarrRadarrApp, tmdbID int) (any, logging.LogErrorInfo) {
	return SR_GetItemInfoFromTMDBID(ctx, app, tmdbID)
}

func (r *RadarrApp) GetItemInfoFromTMDBID(ctx context.Context, app Config_SonarrRadarrApp, tmdbID int) (any, logging.LogErrorInfo) {
	return SR_GetItemInfoFromTMDBID(ctx, app, tmdbID)
}

//////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////

func SR_CallGetItemInfoFromTMDBID(ctx context.Context, app Config_SonarrRadarrApp, tmdbID int) (any, logging.LogErrorInfo) {
	srInterface, Err := NewSonarrRadarrInterface(ctx, app)
	if Err.Message != "" {
		return nil, Err
	}

	return srInterface.GetItemInfoFromTMDBID(ctx, app, tmdbID)
}

//////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////

func SR_GetItemInfoFromTMDBID(ctx context.Context, app Config_SonarrRadarrApp, tmdbID int) (any, logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, fmt.Sprintf("Fetching full item info from TMDB ID %d from %s (%s)", tmdbID, app.Type, app.Library), logging.LevelInfo)
	defer logAction.Complete()

	actionGetBaseInfo := logAction.AddSubAction("Getting base item info", logging.LevelDebug)
	actionGetURL := actionGetBaseInfo.AddSubAction("Constructing API URL", logging.LevelDebug)
	urlEndpoint := ""
	switch app.Type {
	case "Sonarr":
		urlEndpoint = "series"
	case "Radarr":
		urlEndpoint = "movie"
	default:
		actionGetURL.SetError("Unsupported application type", "Ensure the application type is either 'Sonarr' or 'Radarr'", map[string]any{"appType": app.Type})
		return nil, *actionGetURL.Error
	}

	u, err := url.Parse(app.URL)
	if err != nil {
		actionGetURL.SetError("Failed to parse base URL", "Ensure the URL is valid", map[string]any{"error": err.Error()})
		return nil, *actionGetURL.Error
	}
	u.Path = path.Join(u.Path, "api/v3", urlEndpoint, "lookup")
	query := u.Query()
	query.Set("term", fmt.Sprintf("tmdb:%d", tmdbID))
	u.RawQuery = query.Encode()
	URL := u.String()
	actionGetURL.Complete()

	// Make the Auth Headers for Request
	headers := MakeAuthHeader("X-Api-Key", app.APIKey)

	// Make the API Request
	httpResp, respBody, Err := MakeHTTPRequest(ctx, URL, "GET", headers, 60, nil, app.Type)
	if Err.Message != "" {
		return nil, Err
	}

	// Check for non-200 response
	if httpResp.StatusCode != http.StatusOK {
		actionGetBaseInfo.SetError("Non-200 response from Sonarr/Radarr API", fmt.Sprintf("Status Code: %d, Response: %s", httpResp.StatusCode, string(respBody)), nil)
		return nil, *actionGetBaseInfo.Error
	}

	// Parse the response body
	var results []struct {
		ID         int `json:"id"`
		SR_TMDB_ID int `json:"tmdbId"`
	}
	Err = DecodeJSONBody(ctx, respBody, &results, "Sonarr/Radarr TMDB ID Lookup Decoding")
	if Err.Message != "" {
		return nil, Err
	}
	if len(results) == 0 {
		actionGetBaseInfo.SetError("No results from TMDB ID lookup", "The TMDB ID does not correspond to any item in the Sonarr/Radarr library", map[string]any{"tmdbID": tmdbID})
		return nil, *actionGetBaseInfo.Error
	}

	result := results[0]

	// Make sure that the TMDB ID Matches
	if result.SR_TMDB_ID != tmdbID {
		actionGetBaseInfo.SetError("TMDB ID mismatch in response", "The TMDB ID in the response does not match the requested TMDB ID", map[string]any{"requestedTMDBID": tmdbID, "responseTMDBID": result.SR_TMDB_ID})
		return nil, *actionGetBaseInfo.Error
	}

	// If ID is not there, this means the item is not in Sonarr/Radarr
	if result.ID == 0 {
		actionGetBaseInfo.SetError("Item not found in Sonarr/Radarr", "The item with the specified TMDB ID does not exist in the Sonarr/Radarr library", map[string]any{"tmdbID": tmdbID})
		return nil, *actionGetBaseInfo.Error
	}
	actionGetBaseInfo.Complete()

	// Now that we have the base info, fetch the full item details
	actionFetchFullInfo := logAction.AddSubAction("Fetching full item details", logging.LevelInfo)

	u, err = url.Parse(app.URL)
	if err != nil {
		actionFetchFullInfo.SetError("Failed to parse base URL", "Ensure the URL is valid", map[string]any{"error": err.Error()})
		return nil, *actionFetchFullInfo.Error
	}
	u.Path = path.Join(u.Path, "api/v3", urlEndpoint, fmt.Sprintf("%d", result.ID))
	URL = u.String()

	// Make the API Request for full details
	httpResp, respBody, Err = MakeHTTPRequest(ctx, URL, "GET", headers, 60, nil, app.Type)
	if Err.Message != "" {
		return nil, Err
	}

	// Check for non-200 response
	if httpResp.StatusCode != http.StatusOK {
		actionFetchFullInfo.SetError("Non-200 response from Sonarr/Radarr API", fmt.Sprintf("Status Code: %d, Response: %s", httpResp.StatusCode, string(respBody)), nil)
		return nil, *actionFetchFullInfo.Error
	}

	// Parse the full item response body
	switch app.Type {
	case "Sonarr":
		var item SR_SonarrItem
		err = json.Unmarshal(respBody, &item)
		if err != nil {
			actionFetchFullInfo.SetError("Failed to decode Sonarr item info", "Ensure the response format is correct", map[string]any{"error": err.Error()})
			return nil, *actionFetchFullInfo.Error
		}
		actionFetchFullInfo.Complete()
		return item, logging.LogErrorInfo{}
	case "Radarr":
		var item SR_RadarrItem
		err = json.Unmarshal(respBody, &item)
		if err != nil {
			actionFetchFullInfo.SetError("Failed to decode Radarr item info", "Ensure the response format is correct", map[string]any{"error": err.Error()})
			return nil, *actionFetchFullInfo.Error
		}
		actionFetchFullInfo.Complete()
		return item, logging.LogErrorInfo{}
	default:
		actionFetchFullInfo.SetError("Unsupported application type", "Ensure the application type is either 'Sonarr' or 'Radarr'", map[string]any{"appType": app.Type})
		return nil, *actionFetchFullInfo.Error
	}
}
