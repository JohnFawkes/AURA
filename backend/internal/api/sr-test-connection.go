package api

import (
	"aura/internal/logging"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
)

func (s *SonarrApp) TestConnection(ctx context.Context, app Config_SonarrRadarrApp) (bool, logging.LogErrorInfo) {
	return SR_TestConnection(ctx, app)
}

func (r *RadarrApp) TestConnection(ctx context.Context, app Config_SonarrRadarrApp) (bool, logging.LogErrorInfo) {
	return SR_TestConnection(ctx, app)
}

//////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////

func SR_CallTestConnection(ctx context.Context, app Config_SonarrRadarrApp) (bool, logging.LogErrorInfo) {
	srInterface, Err := NewSonarrRadarrInterface(ctx, app)
	if Err.Message != "" {
		return false, Err
	}

	return srInterface.TestConnection(ctx, app)
}

//////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////

func SR_TestConnection(ctx context.Context, app Config_SonarrRadarrApp) (bool, logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, fmt.Sprintf("Testing %s (%s) Connection", app.Type, app.Library), logging.LevelTrace)
	defer logAction.Complete()

	// Make the URL
	u, err := url.Parse(app.URL)
	if err != nil {
		logAction.SetError("Failed to parse base URL", "Ensure the URL is valid", map[string]any{
			"error": err.Error(),
			"url":   app.URL,
		})
		return false, *logAction.Error
	}
	u.Path = path.Join(u.Path, "api/v3", "system/status")
	URL := u.String()

	// Make the Auth Headers for Request
	headers := MakeAuthHeader("X-Api-Key", app.APIKey)

	// Make the request
	httpResp, respBody, Err := MakeHTTPRequest(ctx, URL, http.MethodGet, headers, 60, nil, app.Type)
	if Err.Message != "" {
		return false, Err
	}

	// Check for valid status code
	if httpResp.StatusCode != http.StatusOK {
		logAction.SetError("Received non-200 status code", "Check the URL and API key are correct", map[string]any{"status_code": httpResp.StatusCode})
		return false, *logAction.Error
	}

	// Decode the response
	type ConnectionInfo struct {
		Version string `json:"version"`
	}
	var connInfo ConnectionInfo
	Err = DecodeJSONBody(ctx, respBody, &connInfo, "ConnectionInfo")
	if Err.Message != "" {
		logAction.SetError("Failed to decode response body", Err.Message, nil)
		return false, *logAction.Error
	}

	logAction.Result = map[string]any{
		"version": connInfo.Version,
	}

	return true, logging.LogErrorInfo{}
}
