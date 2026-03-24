package sonarr_radarr

import (
	"aura/config"
	"aura/logging"
	"aura/utils/httpx"
	"context"
	"fmt"
	"net/url"
	"path"
)

func (s *SonarrApp) GetAllTags(ctx context.Context, app config.Config_SonarrRadarrApp) (tags []SonarrRadarrTag, Err logging.LogErrorInfo) {
	return srGetAllTags(ctx, app)
}

func (r *RadarrApp) GetAllTags(ctx context.Context, app config.Config_SonarrRadarrApp) (tags []SonarrRadarrTag, Err logging.LogErrorInfo) {
	return srGetAllTags(ctx, app)
}

func GetAllTags(ctx context.Context, app config.Config_SonarrRadarrApp) (tags []SonarrRadarrTag, Err logging.LogErrorInfo) {
	interfaceSR, Err := NewSonarrRadarrInterface(ctx, app)
	if Err.Message != "" {
		return nil, Err
	}
	return interfaceSR.GetAllTags(ctx, app)
}

func srGetAllTags(ctx context.Context, app config.Config_SonarrRadarrApp) (tags []SonarrRadarrTag, Err logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, fmt.Sprintf("Getting all %s Tags | %s", app.Type, app.Library), logging.LevelInfo)
	defer logAction.Complete()

	tags = []SonarrRadarrTag{}
	Err = logging.LogErrorInfo{}

	u, err := url.Parse(app.URL)
	if err != nil {
		logAction.SetError(fmt.Sprintf("Invalid %s URL", app.Type),
			"Make sure that the URL is properly formatted",
			map[string]any{
				"url":   app.URL,
				"error": err.Error(),
			})
		return nil, *logAction.Error
	}
	u.Path = path.Join(u.Path, "api", "v3", "tag")
	URL := u.String()

	// Make the request to Sonarr/Radarr
	httpResp, respBody, Err := makeRequest(ctx, app, URL, "GET", nil)
	if Err.Message != "" {
		logAction.SetErrorFromInfo(Err)
		return nil, *logAction.Error
	}
	defer httpResp.Body.Close()

	// Decode the response body
	Err = httpx.DecodeResponseToJSON(ctx, respBody, &tags, fmt.Sprintf("Decoding %s Tags Response", app.Type))
	if Err.Message != "" {
		return nil, Err
	}

	logAction.AppendResult("tag_count", len(tags))
	return tags, Err
}
