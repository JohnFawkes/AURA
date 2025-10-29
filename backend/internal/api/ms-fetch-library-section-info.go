package api

import (
	"aura/internal/logging"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
)

func (p *PlexServer) FetchLibrarySectionInfo(ctx context.Context, library *Config_MediaServerLibrary) (bool, logging.LogErrorInfo) {
	return Plex_FetchLibrarySectionInfo(ctx, library)
}

func (e *EmbyJellyServer) FetchLibrarySectionInfo(ctx context.Context, library *Config_MediaServerLibrary) (bool, logging.LogErrorInfo) {
	return EJ_FetchLibrarySectionInfo(ctx, library)
}

//////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////

func CallFetchLibrarySectionInfo(ctx context.Context) ([]LibrarySection, logging.LogErrorInfo) {
	mediaServer, _, Err := NewMediaServerInterface(ctx, Config_MediaServer{})
	if Err.Message != "" {
		return nil, Err
	}

	var allSections []LibrarySection
	for _, library := range Global_Config.MediaServer.Libraries {
		found, Err := mediaServer.FetchLibrarySectionInfo(ctx, &library)
		if Err.Message != "" {
			continue
		}
		if !found {
			continue
		}

		var section LibrarySection
		section.ID = library.SectionID
		section.Type = library.Type
		section.Title = library.Name
		section.Path = library.Path
		allSections = append(allSections, section)
	}
	return allSections, logging.LogErrorInfo{}
}

//////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////

func Plex_FetchLibrarySectionInfo(ctx context.Context, library *Config_MediaServerLibrary) (bool, logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, fmt.Sprintf("Fetch %s Library Section Info", library.Name), logging.LevelDebug)
	defer logAction.Complete()

	// Construct the URL for the Plex server API request
	u, err := url.Parse(Global_Config.MediaServer.URL)
	if err != nil {
		logAction.SetError("Failed to parse base URL", "Ensure the URL is valid", map[string]any{"error": err.Error()})
		return false, *logAction.Error
	}
	u.Path = path.Join(u.Path, "library", "sections", "all")
	URL := u.String()

	// Make the API request to Plex
	httpResp, respBody, logErr := MakeHTTPRequest(ctx, URL, http.MethodGet, nil, 60, nil, "Plex")
	if logErr.Message != "" {
		return false, logErr
	}
	defer httpResp.Body.Close()

	// Parse the response body into a PlexLibrarySectionsWrapper struct
	var plexResponse PlexLibrarySectionsWrapper
	logErr = DecodeJSONBody(ctx, respBody, &plexResponse, "PlexLibrarySectionsWrapper")
	if logErr.Message != "" {
		return false, logErr
	}

	// Find the library section with the matching Name
	found := false
	for _, section := range plexResponse.MediaContainer.Directory {
		if section.Type != "movie" && section.Type != "show" {
			continue
		}
		if section.Title == library.Name {
			library.Type = section.Type
			library.SectionID = section.Key
			library.Path = section.Location[0].Path
			found = true
			break
		}
	}
	if !found {
		logAction.SetError("Library section not found", fmt.Sprintf("Ensure the library section '%s' exists on the Plex server.", library.Name), map[string]any{
			"URL": URL,
		})
		return false, *logAction.Error
	}

	return true, logging.LogErrorInfo{}
}

//////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////

func EJ_FetchLibrarySectionInfo(ctx context.Context, library *Config_MediaServerLibrary) (bool, logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, fmt.Sprintf("Fetch %s Library Section Info", library.Name), logging.LevelDebug)
	defer logAction.Complete()

	// Construct the URL for the Emby/Jellyfin server API request
	u, err := url.Parse(Global_Config.MediaServer.URL)
	if err != nil {
		logAction.SetError("Failed to parse base URL", "Ensure the URL is valid", map[string]any{"error": err.Error()})
		return false, *logAction.Error
	}
	u.Path = path.Join(u.Path, "Users", Global_Config.MediaServer.UserID, "Items")
	URL := u.String()

	// Make a GET request to the Emby server
	httpResp, respBody, logErr := MakeHTTPRequest(ctx, URL, http.MethodGet, nil, 60, nil, "MediaServer")
	if logErr.Message != "" {
		return false, logErr
	}
	defer httpResp.Body.Close()

	// Parse the response body into an EmbyJellyLibrarySectionsResponse struct
	var responseSection EmbyJellyLibrarySectionsResponse
	logErr = DecodeJSONBody(ctx, respBody, &responseSection, "EmbyJellyLibrarySectionsResponse")
	if logErr.Message != "" {
		return false, logErr
	}

	found := false
	for _, item := range responseSection.Items {
		if item.Name == library.Name {
			library.Type = map[string]string{
				"movies":  "movie",
				"tvshows": "show",
			}[item.CollectionType]
			library.SectionID = item.ID
			found = true
			break
		}
	}
	if !found {
		logAction.SetError("Library section not found", fmt.Sprintf("Ensure the library section '%s' exists on the media server.", library.Name), nil)
		return false, *logAction.Error
	}

	return true, logging.LogErrorInfo{}
}
