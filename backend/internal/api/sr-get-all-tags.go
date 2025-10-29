package api

import (
	"aura/internal/logging"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
)

func (s *SonarrApp) GetAllTags(ctx context.Context, app Config_SonarrRadarrApp) ([]SonarrRadarrTag, logging.LogErrorInfo) {
	return SR_GetAllTags(ctx, app)
}

func (r *RadarrApp) GetAllTags(ctx context.Context, app Config_SonarrRadarrApp) ([]SonarrRadarrTag, logging.LogErrorInfo) {
	return SR_GetAllTags(ctx, app)
}

//////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////

func SR_CallGetAllTags(ctx context.Context, app Config_SonarrRadarrApp) ([]SonarrRadarrTag, logging.LogErrorInfo) {
	srInterface, Err := NewSonarrRadarrInterface(ctx, app)
	if Err.Message != "" {
		return nil, Err
	}

	return srInterface.GetAllTags(ctx, app)
}

//////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////

func SR_GetAllTags(ctx context.Context, app Config_SonarrRadarrApp) ([]SonarrRadarrTag, logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, fmt.Sprintf("Fetching all tags from %s (%s)", app.Type, app.Library), logging.LevelInfo)
	defer logAction.Complete()

	var allTags []SonarrRadarrTag

	u, err := url.Parse(app.URL)
	if err != nil {
		logAction.SetError("Failed to parse base URL", "Ensure the URL is valid", map[string]any{"error": err.Error()})
		return nil, *logAction.Error
	}
	u.Path = path.Join(u.Path, "api/v3", "tag")
	URL := u.String()

	apiHeader := map[string]string{
		"X-Api-Key": app.APIKey,
	}

	// Make the API Request
	httpResp, respBody, Err := MakeHTTPRequest(ctx, URL, http.MethodGet, apiHeader, 60, nil, app.Type)
	if Err.Message != "" {
		return nil, Err
	}

	if httpResp.StatusCode != http.StatusOK {
		logAction.SetError("Non-200 response from Sonarr/Radarr API", fmt.Sprintf("Status Code: %d, Response: %s", httpResp.StatusCode, string(respBody)), nil)
		return nil, *logAction.Error
	}

	// Decode the response
	Err = DecodeJSONBody(ctx, respBody, &allTags, "Sonarr/Radarr Tags Decoding")
	if Err.Message != "" {
		return nil, Err
	}

	return allTags, logging.LogErrorInfo{}
}
