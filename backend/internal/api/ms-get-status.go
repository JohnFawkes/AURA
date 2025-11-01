package api

import (
	"aura/internal/logging"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
)

func (p *PlexServer) GetMediaServerStatus(ctx context.Context, msConfig Config_MediaServer) (string, logging.LogErrorInfo) {
	return Plex_GetMediaServerStatus(ctx, msConfig)
}

func (e *EmbyJellyServer) GetMediaServerStatus(ctx context.Context, msConfig Config_MediaServer) (string, logging.LogErrorInfo) {
	return EmbyJelly_GetMediaServerStatus(ctx, msConfig)
}

//////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////

func CallGetMediaServerStatus(ctx context.Context, msConfig Config_MediaServer) (string, logging.LogErrorInfo) {
	mediaServer, _, err := NewMediaServerInterface(ctx, msConfig)
	if err.Message != "" || mediaServer == nil {
		return "", err
	}
	return mediaServer.GetMediaServerStatus(ctx, msConfig)
}

//////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////

func Plex_GetMediaServerStatus(ctx context.Context, msConfig Config_MediaServer) (string, logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Checking Plex Connection", logging.LevelDebug)
	defer logAction.Complete()

	// Construct the URL for the Plex server API request
	u, err := url.Parse(msConfig.URL)
	if err != nil {
		logAction.SetError("Failed to parse base URL", "Ensure the URL is valid", map[string]any{"error": err.Error()})
		return "", *logAction.Error
	}
	u.Path = path.Join(u.Path, "/")
	URL := u.String()

	// Add the token as a header
	headers := MakeAuthHeader("X-Plex-Token", msConfig.Token)

	// Make the HTTP request to Plex
	httpResp, respBody, logErr := MakeHTTPRequest(ctx, URL, http.MethodGet, headers, 60, nil, "Plex")
	if logErr.Message != "" {
		return "", logErr
	}
	defer httpResp.Body.Close()

	// Check the response status code
	if httpResp.StatusCode != 200 {
		logAction.SetError("Plex server returned non-200 status", fmt.Sprintf("Status Code: %d", httpResp.StatusCode), nil)
		return "", *logAction.Error
	}

	// Decode the response body
	var plexResponse PlexConnectionInfoWrapper
	logErr = DecodeJSONBody(ctx, respBody, &plexResponse, "PlexConnectionInfoWrapper")
	if logErr.Message != "" {
		return "", logErr
	}

	// Get the server version
	serverVersion := plexResponse.MediaContainer.Version

	if serverVersion == "" {
		logAction.SetError("Failed to retrieve Plex server version",
			"Ensure that the Plex server is running and accessible",
			map[string]any{
				"URL": URL,
			})
		return "", *logAction.Error
	}

	logAction.AppendResult("server_version", serverVersion)

	return serverVersion, logging.LogErrorInfo{}
}

//////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////

func EmbyJelly_GetMediaServerStatus(ctx context.Context, msConfig Config_MediaServer) (string, logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, fmt.Sprintf("Checking %s Connection", msConfig.Type), logging.LevelDebug)
	defer logAction.Complete()

	// Construct the URL for the Emby/Jellyfin server API request
	u, err := url.Parse(msConfig.URL)
	if err != nil {
		logAction.SetError("Failed to parse base URL", "Ensure the URL is valid", map[string]any{"error": err.Error()})
		return "", *logAction.Error
	}
	u.Path = path.Join(u.Path, "/System/Info")
	URL := u.String()

	// Make headers for authentication
	headers := MakeAuthHeader("X-Emby-Token", msConfig.Token)

	httpResp, respBody, logErr := MakeHTTPRequest(ctx, URL, http.MethodGet, headers, 60, nil, msConfig.Type)
	if logErr.Message != "" {
		return "", logErr
	}
	defer httpResp.Body.Close()

	// Decode the response body
	var statusResponse struct {
		Version string `json:"Version"`
	}
	logErr = DecodeJSONBody(ctx, respBody, &statusResponse, "EmbyJellyfinStatusResponse")
	if logErr.Message != "" {
		return "", logErr
	}

	if statusResponse.Version == "" {
		logAction.SetError("Failed to retrieve Emby/Jellyfin server version",
			"Ensure that the Emby/Jellyfin server is running and accessible",
			map[string]any{
				"URL": URL,
			})
		return "", *logAction.Error
	}

	logAction.AppendResult("server_version", statusResponse.Version)

	return statusResponse.Version, logging.LogErrorInfo{}
}
