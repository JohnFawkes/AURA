package api

import (
	"aura/internal/logging"
	"encoding/json"
	"fmt"
	"net/url"
)

func (s *SonarrApp) TestConnection(app Config_SonarrRadarrApp) (bool, logging.StandardError) {
	return SR_TestConnection(app)
}

func (r *RadarrApp) TestConnection(app Config_SonarrRadarrApp) (bool, logging.StandardError) {
	return SR_TestConnection(app)
}

//////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////

func SR_CallTestConnection(app Config_SonarrRadarrApp) (bool, logging.StandardError) {
	// Make sure all of the required information is set
	Err := SR_MakeSureAllInfoIsSet(app)
	if Err.Message != "" {
		return false, Err
	}

	// Get the appropriate interface
	interfaceSR, Err := SR_GetSonarrRadarrInterface(app)
	if Err.Message != "" {
		return false, Err
	}

	return interfaceSR.TestConnection(app)
}

//////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////

func SR_TestConnection(app Config_SonarrRadarrApp) (bool, logging.StandardError) {
	logging.LOG.Trace(fmt.Sprintf("Checking %s (%s) connection", app.Type, app.Library))

	u, _ := url.Parse(app.URL)
	u.Path = "/api/v3/system/status"
	urlStr := u.String()

	apiHeader := map[string]string{
		"X-Api-Key": app.APIKey,
	}

	response, body, Err := MakeHTTPRequest(urlStr, "GET", apiHeader, 60, nil, app.Type)
	if Err.Message != "" {
		logging.LOG.Error(fmt.Sprintf("%v", Err))
		return false, Err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		Err.Message = fmt.Sprintf("Failed to connect to %s (%s) instance", app.Type, app.Library)
		Err.HelpText = fmt.Sprintf("Ensure the %s instance is running and accessible at the configured URL with the correct API Key.", app.Type)
		Err.Details = map[string]any{
			"statusCode": response.StatusCode,
			"request":    urlStr,
			"response":   string(body),
		}
		return false, Err
	}

	type ConnectionInfo struct {
		Version string `json:"version"`
	}

	var connInfo ConnectionInfo
	err := json.Unmarshal(body, &connInfo)
	if err != nil {
		Err.Message = fmt.Sprintf("Failed to parse %s (%s) response", app.Type, app.Library)
		Err.HelpText = fmt.Sprintf("Ensure the %s (%s) instance is returning a valid JSON response.", app.Type, app.Library)
		Err.Details = map[string]any{
			"error":   err.Error(),
			"request": urlStr,
		}
		return false, Err
	}
	logging.LOG.Info(fmt.Sprintf("%s (%s) Version: %s", app.Type, app.Library, connInfo.Version))

	return true, logging.StandardError{}
}
