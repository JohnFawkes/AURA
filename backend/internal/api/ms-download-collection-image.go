package api

import (
	"aura/internal/logging"
	"context"
	"fmt"
)

func (*PlexServer) DownloadAndUpdateCollectionImage(ctx context.Context, collectionItem CollectionItem, file PosterFile) logging.LogErrorInfo {
	return Plex_DownloadAndUpdateCollectionImage(ctx, collectionItem, file)
}

func (e *EmbyJellyServer) DownloadAndUpdateCollectionImage(ctx context.Context, collectionItem CollectionItem, file PosterFile) logging.LogErrorInfo {
	return EmbyJelly_DownloadAndUpdateCollectionImage(ctx, collectionItem, file)
}

//////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////

func CallDownloadAndUpdateCollectionImage(ctx context.Context, collectionItem CollectionItem, file PosterFile) logging.LogErrorInfo {
	mediaServer, _, Err := NewMediaServerInterface(ctx, Config_MediaServer{})
	if Err.Message != "" {
		return Err
	}

	return mediaServer.DownloadAndUpdateCollectionImage(ctx, collectionItem, file)
}

//////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////

func Plex_DownloadAndUpdateCollectionImage(ctx context.Context, collectionItem CollectionItem, file PosterFile) logging.LogErrorInfo {
	var posterImageType string
	switch file.Type {
	case "poster":
		posterImageType = "Poster"
	case "backdrop":
		posterImageType = "Backdrop"
	default:
		return logging.LogErrorInfo{
			Message: "Unsupported file type for Plex collection image",
			Help:    "Ensure that the file type is either 'poster' or 'backdrop'",
			Detail: map[string]any{
				"file_type": file.Type,
			},
		}
	}

	ctx, logAction := logging.AddSubActionToContext(ctx, fmt.Sprintf("Downloading Collection %s Image in Plex", posterImageType), logging.LevelInfo)
	defer logAction.Complete()

	Err := Plex_UpdateCollectionImageViaMediuxURL(ctx, collectionItem, file)
	if Err.Message != "" {
		return Err
	}
	return logging.LogErrorInfo{}

}

//////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////

func EmbyJelly_DownloadAndUpdateCollectionImage(ctx context.Context, collectionItem CollectionItem, file PosterFile) logging.LogErrorInfo {
	return logging.LogErrorInfo{}
}
