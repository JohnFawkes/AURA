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
	var posterImageType string
	var posterType string
	switch file.Type {
	case "poster":
		posterImageType = "Poster"
		posterType = "Primary"
	case "backdrop":
		posterImageType = "Backdrop"
		posterType = "Backdrop"
	}

	ctx, logAction := logging.AddSubActionToContext(ctx, fmt.Sprintf("Downloading and Updating %s in %s", posterImageType, Global_Config.MediaServer.Type), logging.LevelInfo)
	defer logAction.Complete()

	// Get the Image from MediUX
	// Mediux_GetImage will handle checking the temp folder and caching based on config
	formatDate := file.Modified.Format("20060102150405")
	imageData, _, Err := Mediux_GetImage(ctx, file.ID, formatDate, MediuxImageQualityOriginal)
	if Err.Message != "" {
		return Err
	}

	// If the posterType is Backdrop, we need to set the index to 0 to replace the current backdrop
	// First we will get the list of current images
	if posterType == "Backdrop" {
		currentImages, logErr := EJ_GetCurrentImages(ctx, collectionItem.Title, collectionItem.RatingKey, "Current")
		if logErr.Message != "" {
			return logErr
		}

		// Upload the new image
		logErr = EJ_UploadImage(ctx, collectionItem.Title, collectionItem.RatingKey, file, imageData)
		if logErr.Message != "" {
			return logErr
		}

		if len(currentImages) != 0 {
			// Get the list of images again to find the new one
			newImages, logErr := EJ_GetCurrentImages(ctx, collectionItem.Title, collectionItem.RatingKey, "New")
			if logErr.Message != "" {
				return logErr
			}

			// Find the new image by comparing currentImages and newImages
			newImageItem := EJ_FindNewImage(currentImages, newImages, posterType)
			if newImageItem.ImageTag == "" && newImageItem.ImagePath == "" {
				logAction.SetError("Failed to find new image tag after upload",
					"Ensure the image was uploaded successfully",
					map[string]any{
						"currentImages": currentImages,
						"newImages":     newImages,
					})
				return *logAction.Error
			}

			// Now we change the image index to 0, if it's not already 0
			if newImageItem.ImageIndex != 0 {
				logErr = EJ_ChangeImageIndex(ctx, collectionItem.Title, collectionItem.RatingKey, newImageItem)
				if logErr.Message != "" {
					return logErr
				}
			}
		}
	} else {
		// For Primary images, just upload the image
		logErr := EJ_UploadImage(ctx, collectionItem.Title, collectionItem.RatingKey, file, imageData)
		if logErr.Message != "" {
			return logErr
		}
	}

	logAction.Complete()
	return logging.LogErrorInfo{}
}
