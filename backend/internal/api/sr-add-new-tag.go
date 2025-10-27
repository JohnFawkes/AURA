package api

import (
	"aura/internal/logging"
	"encoding/json"
	"fmt"
	"net/url"
	"path"
)

func (s *SonarrApp) AddNewTags(app Config_SonarrRadarrApp, tags []string) ([]SonarrRadarrTag, logging.StandardError) {
	return SR_AddNewTags(app, tags)
}

func (r *RadarrApp) AddNewTags(app Config_SonarrRadarrApp, tags []string) ([]SonarrRadarrTag, logging.StandardError) {
	return SR_AddNewTags(app, tags)
}

//////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////

func SR_CallAddNewTags(app Config_SonarrRadarrApp, tags []string) ([]SonarrRadarrTag, logging.StandardError) {
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

	return interfaceSR.AddNewTags(app, tags)
}

//////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////

func SR_AddNewTags(app Config_SonarrRadarrApp, tags []string) ([]SonarrRadarrTag, logging.StandardError) {
	Err := logging.NewStandardError()
	var addedTags []SonarrRadarrTag

	logging.LOG.Trace(fmt.Sprintf("Adding new tags [%v] to %s (%s)", tags, app.Type, app.Library))

	u, _ := url.Parse(app.URL)
	u.Path = path.Join(u.Path, "api/v3", "tag")
	url := u.String()

	apiHeader := map[string]string{
		"X-Api-Key": app.APIKey,
	}

	for _, tag := range tags {
		tagData := map[string]string{
			"label": tag,
		}
		tagBody, _ := json.Marshal(tagData)

		response, body, Err := MakeHTTPRequest(url, "POST", apiHeader, 60, tagBody, app.Type)
		if Err.Message != "" {
			return nil, Err
		}
		defer response.Body.Close()

		if response.StatusCode != 201 {
			Err.Message = fmt.Sprintf("Failed to add tag '%s' to %s (%s)", tag, app.Type, app.Library)
			Err.HelpText = fmt.Sprintf("Ensure the %s instance is running and accessible at the configured URL with the correct API Key.", app.Type)
			Err.Details = map[string]any{
				"statusCode": response.StatusCode,
				"request":    url,
				"response":   string(body),
			}
			return nil, Err
		}

		var newTag SonarrRadarrTag
		err := json.Unmarshal(body, &newTag)
		if err != nil {
			Err.Message = "Failed to parse response when adding new tag"
			Err.HelpText = "Ensure the Sonarr/Radarr instance is running the correct version and is accessible."
			Err.Details = map[string]any{
				"error":        err.Error(),
				"responseBody": string(body),
			}
			return nil, Err
		}

		addedTags = append(addedTags, newTag)
		logging.LOG.Info(fmt.Sprintf("Added new tag '%s' (ID: %d) to %s (%s)", newTag.Label, newTag.ID, app.Type, app.Library))
	}

	return addedTags, Err
}
