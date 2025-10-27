package api

import (
	"aura/internal/logging"
	"encoding/json"
	"fmt"
	"net/url"
	"path"
)

func (s *SonarrApp) GetAllTags(app Config_SonarrRadarrApp) ([]SonarrRadarrTag, logging.StandardError) {
	return SR_GetAllTags(app)
}

func (r *RadarrApp) GetAllTags(app Config_SonarrRadarrApp) ([]SonarrRadarrTag, logging.StandardError) {
	return SR_GetAllTags(app)
}

//////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////

func SR_CallGetAllTags(app Config_SonarrRadarrApp) ([]SonarrRadarrTag, logging.StandardError) {
	// Make sure all of the required information is set
	Err := SR_MakeSureAllInfoIsSet(app)
	if Err.Message != "" {
		return nil, Err
	}

	// Get the appropriate interface
	interfaceSR, Err := SR_GetSonarrRadarrInterface(app)
	if Err.Message != "" {
		return nil, Err
	}

	return interfaceSR.GetAllTags(app)
}

//////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////

func SR_GetAllTags(app Config_SonarrRadarrApp) ([]SonarrRadarrTag, logging.StandardError) {
	Err := logging.NewStandardError()
	var tags []SonarrRadarrTag

	logging.LOG.Trace(fmt.Sprintf("Getting all tags from %s (%s)", app.Type, app.Library))

	u, _ := url.Parse(app.URL)
	u.Path = path.Join(u.Path, "api/v3", "tag")
	url := u.String()

	apiHeader := map[string]string{
		"X-Api-Key": app.APIKey,
	}

	response, body, Err := MakeHTTPRequest(url, "GET", apiHeader, 60, nil, app.Type)
	if Err.Message != "" {
		return nil, Err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		Err.Message = fmt.Sprintf("Failed to get tags from %s (%s) instance", app.Type, app.Library)
		Err.HelpText = fmt.Sprintf("Ensure the %s instance is running and accessible at the configured URL with the correct API Key.", app.Type)
		Err.Details = map[string]any{
			"statusCode": response.StatusCode,
			"request":    url,
			"response":   string(body),
		}
		return nil, Err
	}

	// Parse the response body
	var parsedBody []SonarrRadarrTag
	if err := json.Unmarshal(body, &parsedBody); err != nil {
		Err.Message = fmt.Sprintf("Failed to parse tags from %s (%s) response", app.Type, app.Library)
		Err.HelpText = "Ensure the Sonarr/Radarr instance is running the correct version and is accessible."
		Err.Details = map[string]any{
			"error":    err.Error(),
			"response": string(body),
		}
		return nil, Err
	}

	tags = parsedBody

	return tags, Err
}
