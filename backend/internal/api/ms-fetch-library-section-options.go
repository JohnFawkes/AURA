package api

import (
	"aura/internal/logging"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
)

func (p *PlexServer) FetchLibrarySectionOptions(ctx context.Context, msConfig Config_MediaServer) ([]string, logging.LogErrorInfo) {
	return Plex_FetchLibrarySectionOptions(ctx, msConfig)
}

func (e *EmbyJellyServer) FetchLibrarySectionOptions(ctx context.Context, msConfig Config_MediaServer) ([]string, logging.LogErrorInfo) {
	return EJ_FetchLibrarySectionOptions(ctx, msConfig)
}

/////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////

func CallFetchLibrarySectionOptions(ctx context.Context, msConfig Config_MediaServer) ([]string, logging.LogErrorInfo) {
	mediaServer, msConfig, Err := NewMediaServerInterface(ctx, msConfig)
	if Err.Message != "" {
		return nil, Err
	}

	return mediaServer.FetchLibrarySectionOptions(ctx, msConfig)
}

//////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////

func Plex_FetchLibrarySectionOptions(ctx context.Context, msConfig Config_MediaServer) ([]string, logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Fetching Library Section Options from Plex", logging.LevelDebug)
	defer logAction.Complete()

	// Construct the URL
	u, err := url.Parse(msConfig.URL)
	if err != nil {
		logAction.SetError("Failed to parse base URL", "Ensure the URL is valid", map[string]any{"error": err.Error()})
		return nil, *logAction.Error
	}
	u.Path = path.Join(u.Path, "library", "sections", "all")
	URL := u.String()

	// Make Auth Headers for Request
	headers := MakeAuthHeader("X-Plex-Token", msConfig.Token)

	// Make a GET request to the Plex server
	httpResp, respBody, logErr := MakeHTTPRequest(ctx, URL, http.MethodGet, headers, 60, nil, "Plex")
	if logErr.Message != "" {
		return nil, logErr
	}
	defer httpResp.Body.Close()

	// Parse the response body into a PlexLibrarySectionsWrapper struct
	var plexResponse PlexLibrarySectionsWrapper
	logErr = DecodeJSONBody(ctx, respBody, &plexResponse, "PlexLibrarySectionsWrapper")
	if logErr.Message != "" {
		return nil, logErr
	}

	var sectionOptions []string
	for _, section := range plexResponse.MediaContainer.Directory {
		sectionOptions = append(sectionOptions, section.Title)
	}

	return sectionOptions, logging.LogErrorInfo{}
}

//////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////

func EJ_FetchLibrarySectionOptions(ctx context.Context, msConfig Config_MediaServer) ([]string, logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, fmt.Sprintf("Fetching Library Section Options from %s", msConfig.Type), logging.LevelDebug)
	defer logAction.Complete()

	// Construct the URL for the Emby/Jellyfin server API request
	u, err := url.Parse(msConfig.URL)
	if err != nil {
		logAction.SetError("Failed to parse Emby/Jellyfin base URL", err.Error(), nil)
		return nil, *logAction.Error
	}
	u.Path = path.Join(u.Path, "Users", msConfig.UserID, "Items")
	URL := u.String()

	// Make Auth Headers for Request
	headers := MakeAuthHeader("X-Emby-Token", msConfig.Token)

	// Make a GET request to the Emby/Jellyfin server
	httpResp, respBody, logErr := MakeHTTPRequest(ctx, URL, http.MethodGet, headers, 60, nil, msConfig.Type)
	if logErr.Message != "" {
		return nil, logErr
	}
	defer httpResp.Body.Close()

	// Parse the response body into an EmbyJellyLibrarySectionsResponse struct
	var ejResponse EmbyJellyLibrarySectionsResponse
	logErr = DecodeJSONBody(ctx, respBody, &ejResponse, "EmbyJellyLibrarySectionsResponse")
	if logErr.Message != "" {
		return nil, logErr
	}

	var sectionOptions []string
	for _, item := range ejResponse.Items {
		if item.CollectionType != "tvshows" && item.CollectionType != "movies" {
			continue
		}
		sectionOptions = append(sectionOptions, item.Name)
	}

	return sectionOptions, logging.LogErrorInfo{}
}
