package api

import (
	"aura/internal/logging"
	"context"
	"fmt"
)

type Interface_MediaServer interface {

	// Initialize the Media Server connection
	InitializeMediaServerConnection(ctx context.Context, msConfig Config_MediaServer) (string, logging.LogErrorInfo)

	// Get Status of the Media Server
	GetMediaServerStatus(ctx context.Context, msConfig Config_MediaServer) (string, logging.LogErrorInfo)

	// Get the library section info
	FetchLibrarySectionInfo(ctx context.Context, library *Config_MediaServerLibrary) (bool, logging.LogErrorInfo)

	// Get the library section options
	FetchLibrarySectionOptions(ctx context.Context, msConfig Config_MediaServer) ([]string, logging.LogErrorInfo)

	// Get the library section items
	FetchLibrarySectionItems(ctx context.Context, section LibrarySection, sectionStartIndex string) ([]MediaItem, int, logging.LogErrorInfo)

	// Get an item's content by Rating Key/ID
	FetchItemContent(ctx context.Context, ratingKey string, sectionTitle string) (MediaItem, logging.LogErrorInfo)

	// Get an image from the media server
	FetchImageFromMediaServer(ctx context.Context, ratingKey string, imageType string) ([]byte, logging.LogErrorInfo)

	// Use the set to update the item on the media server
	DownloadAndUpdatePosters(ctx context.Context, mediaItem MediaItem, file PosterFile) logging.LogErrorInfo

	// Use the TMDB ID, type, title and library section to search for the item on the media server
	//SearchForItemAndGetRatingKey(tmdbID, itemType, itemTitle, librarySection string) (string, logging.StandardError)

	// Get the movie collection items
	FetchMovieCollectionItems(ctx context.Context, librarySection LibrarySection) ([]CollectionItem, logging.LogErrorInfo)

	// Get the movie collection item children
	FetchMovieCollectionItemChildren(ctx context.Context, collectionItem *CollectionItem) logging.LogErrorInfo

	// Download and update the collection image
	DownloadAndUpdateCollectionImage(ctx context.Context, collectionItem CollectionItem, file PosterFile) logging.LogErrorInfo
}

type PlexServer struct{}
type EmbyJellyServer struct{}

func NewMediaServerInterface(ctx context.Context, msConfig Config_MediaServer) (Interface_MediaServer, Config_MediaServer, logging.LogErrorInfo) {
	if msConfig.Type == "" && msConfig.URL == "" && msConfig.Token == "" {
		msConfig = Global_Config.MediaServer
	}
	var interfaceMS Interface_MediaServer
	switch msConfig.Type {
	case "Plex":
		interfaceMS = &PlexServer{}
	case "Emby", "Jellyfin":
		interfaceMS = &EmbyJellyServer{}
	default:
		_, logAction := logging.AddSubActionToContext(ctx, "Creating Media Server Interface", logging.LevelTrace)
		logAction.SetError(fmt.Sprintf("Unsupported Media Server Type: '%s'", msConfig.Type),
			"Ensure the Media Server Type is set to either 'Plex', 'Emby' or 'Jellyfin' in the config file",
			nil)
		return nil, msConfig, *logAction.Error
	}

	return interfaceMS, msConfig, logging.LogErrorInfo{}
}
