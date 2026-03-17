package ej

import (
	"aura/config"
	"aura/logging"
	"aura/models"
	"aura/utils/httpx"
	"context"
	"fmt"
	"net/url"
	"path"
)

func (e *EJ) GetLibrarySectionDetails(ctx context.Context, library *models.LibrarySection) (found bool, Err logging.LogErrorInfo) {
	serverType := config.Current.MediaServer.Type

	ctx, logAction := logging.AddSubActionToContext(ctx, fmt.Sprintf("Fetching Details for Library Section: %s from %s Media Server", library.Title, serverType), logging.LevelDebug)
	defer logAction.Complete()

	found = false
	Err = logging.LogErrorInfo{}

	// Construct the URL for the EJ server API request
	u, err := url.Parse(config.Current.MediaServer.URL)
	if err != nil {
		logAction.SetError("Failed to parse base URL", "Ensure the URL is valid", map[string]any{"error": err.Error()})
		return found, *logAction.Error
	}
	u.Path = path.Join(u.Path, "Users", config.Current.MediaServer.UserID, "Items")
	URL := u.String()

	// Make the HTTP Request to EJ
	resp, respBody, Err := makeRequest(ctx, config.Current.MediaServer, URL, "GET", nil)
	if Err.Message != "" {
		logAction.SetErrorFromInfo(Err)
		return found, *logAction.Error
	}
	defer resp.Body.Close()

	// Decode the Response
	var ejResp EmbyJellyLibrarySectionsResponse
	Err = httpx.DecodeResponseToJSON(ctx, respBody, &ejResp, fmt.Sprintf("%s Library Sections Response", serverType))
	if Err.Message != "" {
		return found, *logAction.Error
	}

	for _, item := range ejResp.Items {
		if item.CollectionType == "" {
			item.CollectionType = "mixed"
		}
		if item.CollectionType != "movies" && item.CollectionType != "tvshows" && item.CollectionType != "mixed" {
			continue
		}
		if item.Name == library.Title {
			library.ID = item.ID
			library.Type = map[string]string{
				"movies":  "movie",
				"tvshows": "show",
				"mixed":   "mixed",
			}[item.CollectionType]
			found = true
			break
		}
	}
	if !found {
		logAction.SetError("Library section not found", fmt.Sprintf("Ensure the library section '%s' exists on the %s server.", library.Title, serverType), map[string]any{
			"LibrarySectionTitle": library.Title,
		})
	}

	// Get the path for the library section
	u, err = url.Parse(config.Current.MediaServer.URL)
	if err != nil {
		logAction.SetError("Failed to parse base URL for section path", "Ensure the URL is valid", map[string]any{"error": err.Error()})
		return found, *logAction.Error
	}
	u.Path = path.Join(u.Path, "Library", "VirtualFolders")
	URL = u.String()

	// Make the HTTP Request to EJ for section path
	resp, respBody, Err = makeRequest(ctx, config.Current.MediaServer, URL, "GET", nil)
	if Err.Message != "" {
		logAction.SetErrorFromInfo(Err)
		return found, *logAction.Error
	}
	defer resp.Body.Close()

	// Decode the Response for section path
	var pathResps []EmbyJellyVirtualFoldersResponse
	Err = httpx.DecodeResponseToJSON(ctx, respBody, &pathResps, fmt.Sprintf("%s Library Section Path Response", serverType))
	if Err.Message != "" {
		return found, *logAction.Error
	}

	for _, folder := range pathResps {
		if folder.Name == library.Title {
			// Prefer top-level Locations (present in Jellyfin/Emby virtual folders response)
			if len(folder.Locations) > 0 {
				library.Path = folder.Locations[0]
			} else if len(folder.PathInfos) > 0 {
				// Backward/alternate payload support
				library.Path = folder.PathInfos[0].Path
			}
			break
		}
	}
	if library.Path != "" {
		// Update the library section in the config with the path info
		for i, lib := range config.Current.MediaServer.Libraries {
			if lib.Title == library.Title {
				config.Current.MediaServer.Libraries[i].Path = library.Path
				break
			}
		}
	}

	return found, logging.LogErrorInfo{}
}
