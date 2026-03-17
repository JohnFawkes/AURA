package sonarr_radarr

import (
	"aura/config"
	"aura/logging"
	"aura/utils/httpx"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"path"
)

func (s *SonarrApp) AddNewTags(ctx context.Context, app config.Config_SonarrRadarrApp, newTags []string) (tags []SonarrRadarrTag, Err logging.LogErrorInfo) {
	return srAddNewTags(ctx, app, newTags)
}

func (r *RadarrApp) AddNewTags(ctx context.Context, app config.Config_SonarrRadarrApp, newTags []string) (tags []SonarrRadarrTag, Err logging.LogErrorInfo) {
	return srAddNewTags(ctx, app, newTags)
}

func AddNewTags(ctx context.Context, app config.Config_SonarrRadarrApp, newTags []string) (tags []SonarrRadarrTag, Err logging.LogErrorInfo) {
	interfaceSR, Err := NewSonarrRadarrInterface(ctx, app)
	if Err.Message != "" {
		return nil, Err
	}
	return interfaceSR.AddNewTags(ctx, app, newTags)
}

func srAddNewTags(ctx context.Context, app config.Config_SonarrRadarrApp, newTags []string) (addedTags []SonarrRadarrTag, Err logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, fmt.Sprintf("Adding New Tags to %s | %s", app.Type, app.Library), logging.LevelInfo)
	defer logAction.Complete()

	addedTags = []SonarrRadarrTag{}
	Err = logging.LogErrorInfo{}

	// Construct the URL
	u, err := url.Parse(app.URL)
	if err != nil {
		logAction.SetError(fmt.Sprintf("Invalid %s URL", app.Type),
			"Make sure that the URL is properly formatted",
			map[string]any{
				"url":   app.URL,
				"error": err.Error(),
			})
		return addedTags, *logAction.Error
	}
	u.Path = path.Join(u.Path, "api", "v3", "tag")
	URL := u.String()

	// Loop through and add each new tag
	for _, newTag := range newTags {
		actionLog := logAction.AddSubAction(fmt.Sprintf("Adding Tag '%s' to %s | %s", newTag, app.Type, app.Library), logging.LevelInfo)
		tagData := map[string]string{
			"label": newTag,
		}
		tagBody, _ := json.Marshal(tagData)

		// Make the request to Sonarr/Radarr
		_, respBody, reqErr := makeRequest(ctx, app, URL, "POST", tagBody)
		if reqErr.Message != "" {
			actionLog.SetError(fmt.Sprintf("Failed to Add Tag '%s'", newTag),
				fmt.Sprintf("Make sure the %s server is reachable and the API Token is correct", app.Type),
				map[string]any{
					"error": reqErr,
				})
			continue
		}

		// Decode the response body
		var createdTag SonarrRadarrTag
		Err = httpx.DecodeResponseToJSON(ctx, respBody, &createdTag, fmt.Sprintf("Decoding Add Tag '%s' Response", newTag))
		if Err.Message != "" {
			actionLog.SetError(fmt.Sprintf("Failed to Decode Response for Added Tag '%s'", newTag),
				"Ensure the response from Sonarr/Radarr is valid",
				map[string]any{
					"response": string(respBody),
				})
			continue
		}

		// Append to the tags list
		addedTags = append(addedTags, createdTag)
		actionLog.AppendResult("tag_id", createdTag.ID)
		actionLog.AppendResult("tag_label", createdTag.Label)
		actionLog.Complete()
	}
	return addedTags, Err
}
