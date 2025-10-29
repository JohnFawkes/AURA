package api

import (
	"aura/internal/logging"
	"context"
	"fmt"
	"net/url"
	"path"
)

func (p *PlexServer) InitializeMediaServerConnection(ctx context.Context, msConfig Config_MediaServer) (string, logging.LogErrorInfo) {
	return Plex_GetMediaServerStatus(ctx, msConfig)
}

func (e *EmbyJellyServer) InitializeMediaServerConnection(ctx context.Context, msConfig Config_MediaServer) (string, logging.LogErrorInfo) {
	userID, Err := EJ_InitializeMediaServerConnection(ctx, &msConfig)
	if Err.Message != "" {
		return "", Err
	}
	_, Err = EmbyJelly_GetMediaServerStatus(ctx, msConfig)
	if Err.Message != "" {
		return "", Err
	}
	return userID, logging.LogErrorInfo{}
}

//////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////

func CallInitializeMediaServerConnection(ctx context.Context, msConfig Config_MediaServer) (string, logging.LogErrorInfo) {
	mediaServer, msConfig, logErr := NewMediaServerInterface(ctx, msConfig)
	if mediaServer == nil || logErr.Message != "" {
		return "", logErr
	}
	return mediaServer.InitializeMediaServerConnection(ctx, msConfig)
}

//////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////

func EJ_InitializeMediaServerConnection(ctx context.Context, msConfig *Config_MediaServer) (string, logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, fmt.Sprintf("Initializing %s Media Server Connection", msConfig.Type), logging.LevelDebug)
	defer logAction.Complete()

	// Build the API URL
	u, err := url.Parse(msConfig.URL)
	if err != nil {
		logAction.SetError("Failed to parse base URL", "Ensure the URL is valid", map[string]any{"error": err.Error()})
		return "", *logAction.Error
	}
	u.Path = path.Join(u.Path, "Users")
	URL := u.String()

	// Make header map
	headers := make(map[string]string)
	headers["X-Emby-Token"] = msConfig.Token

	// Make the HTTP request to Emby/Jellyfin
	httpResp, respBody, logErr := MakeHTTPRequest(ctx, URL, "GET", headers, 60, nil, "")
	if logErr.Message != "" {
		return "", logErr
	}
	defer httpResp.Body.Close()

	// Check the response status code
	if httpResp.StatusCode != 200 {
		logAction.SetError("Emby/Jellyfin server returned non-200 status", fmt.Sprintf("Status Code: %d", httpResp.StatusCode), nil)
		return "", *logAction.Error
	}

	// Decode the response body
	var responseSection EmbyJellyUserIDResponse
	logErr = DecodeJSONBody(ctx, respBody, &responseSection, "EmbyJellyUserIDResponse")
	if logErr.Message != "" {
		return "", logErr
	}

	// Find the first Admin user ID
	for _, user := range responseSection {
		if user.Policy.IsAdministrator {
			Global_Config.MediaServer.UserID = user.ID
			msConfig.UserID = user.ID
			logAction.AppendResult("admin user id", user.ID)
			return user.ID, logging.LogErrorInfo{}
		}
	}

	// No Admin user found
	logAction.SetError("No administrator user found on Emby/Jellyfin server", "Ensure that at least one user has administrator privileges", nil)
	return "", *logAction.Error
}
