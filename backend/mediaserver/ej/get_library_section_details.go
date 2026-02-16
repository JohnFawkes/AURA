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
		if item.CollectionType != "movies" && item.CollectionType != "tvshows" {
			continue
		}
		if item.Name == library.Title {
			library.ID = item.ID
			library.Type = map[string]string{
				"movies":  "movie",
				"tvshows": "show",
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

	return found, logging.LogErrorInfo{}
}
