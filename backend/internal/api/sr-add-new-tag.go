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

func (s *SonarrApp) AddNewTags(ctx context.Context, app Config_SonarrRadarrApp, tags []string) ([]SonarrRadarrTag, logging.LogErrorInfo) {
	return SR_AddNewTags(ctx, app, tags)
}

func (r *RadarrApp) AddNewTags(ctx context.Context, app Config_SonarrRadarrApp, tags []string) ([]SonarrRadarrTag, logging.LogErrorInfo) {
	return SR_AddNewTags(ctx, app, tags)
}

//////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////

func SR_CallAddNewTags(ctx context.Context, app Config_SonarrRadarrApp, tags []string) ([]SonarrRadarrTag, logging.LogErrorInfo) {
	srInterface, Err := NewSonarrRadarrInterface(ctx, app)
	if Err.Message != "" {
		return nil, Err
	}

	return srInterface.AddNewTags(ctx, app, tags)
}

//////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////

func SR_AddNewTags(ctx context.Context, app Config_SonarrRadarrApp, tags []string) ([]SonarrRadarrTag, logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, fmt.Sprintf("Adding new tags to %s (%s)", app.Type, app.Library), logging.LevelInfo)
	defer logAction.Complete()

	var addedTags []SonarrRadarrTag

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

	for _, tag := range tags {
		actionLog := logAction.AddSubAction(fmt.Sprintf("Adding new tag '%s' to %s (%s)", tag, app.Type, app.Library), logging.LevelInfo)
		tagData := map[string]string{
			"label": tag,
		}
		tagBody, _ := json.Marshal(tagData)

		// Make the API Request
		httpResp, respBody, Err := MakeHTTPRequest(ctx, URL, http.MethodPost, apiHeader, 60, tagBody, app.Type)
		if Err.Message != "" {
			actionLog.SetError(
				fmt.Sprintf("Failed to add tag '%s' to %s (%s)", tag, app.Type, app.Library),
				fmt.Sprintf("Ensure the %s instance is running and accessible at the configured URL with the correct API Key.", app.Type),
				map[string]any{
					"tag":      tag,
					"response": respBody,
					"error":    Err,
				})
			return nil, Err
		}
		defer httpResp.Body.Close()

		// Check the response status code
		if httpResp.StatusCode != http.StatusCreated {
			actionLog.SetError(
				fmt.Sprintf("Failed to add tag '%s' to %s (%s)", tag, app.Type, app.Library),
				fmt.Sprintf("The %s server returned status code %d when adding the tag.", app.Type, httpResp.StatusCode),
				map[string]any{
					"tag":        tag,
					"StatusCode": httpResp.StatusCode,
					"response":   respBody,
				})
			return nil, *actionLog.Error
		}

		var newTag SonarrRadarrTag
		Err = DecodeJSONBody(ctx, respBody, &newTag, "New Sonarr/Radarr Tag Decoding")
		if Err.Message != "" {
			return nil, Err
		}

		addedTags = append(addedTags, newTag)
		actionLog.AppendResult("tag_id", newTag.ID)
		actionLog.AppendResult("tag_label", newTag.Label)
		actionLog.Complete()
	}

	return []SonarrRadarrTag{}, logging.LogErrorInfo{}
}
