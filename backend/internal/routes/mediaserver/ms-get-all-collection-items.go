package routes_ms

import (
	"aura/internal/api"
	"aura/internal/logging"
	"context"
	"fmt"
	"net/http"
)

func GetCollectionItems(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Get Collection Items", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	// Get the Media Server Type
	var collectionItems []api.CollectionItem
	switch api.Global_Config.MediaServer.Type {
	case "Plex":
		collectionItems = GetPlexCollectionItems(ctx)
	case "Emby", "Jellyfin":
		collectionItems = GetEmbyJellyCollectionItems(ctx)
	default:
		logAction.SetError("Unsupported Media Server Type", "Check configuration for Media Server Type", map[string]any{"type": api.Global_Config.MediaServer.Type})
		ld.Complete()
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}

	api.Util_Response_SendJSON(w, ld, collectionItems)
}

func GetPlexCollectionItems(ctx context.Context) []api.CollectionItem {
	ctx, ld := logging.AddSubActionToContext(ctx, "Get Plex Collection Items", logging.LevelInfo)
	defer ld.Complete()

	// Get all sections from the media server
	allSections, logErr := api.CallFetchLibrarySectionInfo(ctx)
	if logErr.Message != "" {
		return nil
	}

	// Loop through all libraries in the Sections
	var collectionItems []api.CollectionItem
	for _, section := range allSections {
		if section.Type != "movie" {
			continue
		}

		sectionCollectionItems, Err := api.CallFetchMovieCollectionItems(ctx, section)
		if Err.Message != "" {
			return nil
		}
		collectionItems = append(collectionItems, sectionCollectionItems...)
	}

	return collectionItems
}

func GetEmbyJellyCollectionItems(ctx context.Context) []api.CollectionItem {
	ctx, ld := logging.AddSubActionToContext(ctx, fmt.Sprintf("Getting Collection Items from %s", api.Global_Config.MediaServer.Type), logging.LevelInfo)
	defer ld.Complete()

	// Get the Media Server Collection Section
	section, logErr := api.EJ_FetchLibraryCollectionSectionID(ctx, api.Global_Config.MediaServer)
	if logErr.Message != "" {
		return nil
	}

	// Fetch Collection Items from the Collection Section
	collectionItems, logErr := api.EJ_FetchMovieCollectionItems(ctx, section)
	if logErr.Message != "" {
		return nil
	}

	return collectionItems
}
