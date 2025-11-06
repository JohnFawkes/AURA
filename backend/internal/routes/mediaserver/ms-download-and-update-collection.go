package routes_ms

import (
	"aura/internal/api"
	"aura/internal/logging"
	"fmt"
	"net/http"
)

func DownloadAndUpdateCollection(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Download and Update Collection Image", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	// Parse the request body to get posterFile & collectionItem
	var requestBody struct {
		PosterFile     api.PosterFile     `json:"PosterFile"`
		CollectionItem api.CollectionItem `json:"CollectionItem"`
	}
	Err := api.DecodeRequestBodyJSON(ctx, r.Body, &requestBody, "requestBody")
	if Err.Message != "" {
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}

	posterFile := requestBody.PosterFile
	collectionItem := requestBody.CollectionItem
	actionValidate := logAction.AddSubAction("Validate Request Body Fields", logging.LevelDebug)
	// Make sure that the collectionItem has the following fields set
	// 1. CollectionItem.RatingKey
	// 2. CollectionItem.Title
	if collectionItem.RatingKey == "" || collectionItem.Title == "" {
		actionValidate.SetError("Collection Item Rating Key is required",
			"Ensure that the Collection Item in the request body has a Rating Key",
			map[string]any{
				"CollectionItem": collectionItem,
			})
	}

	// Make sure that the posterFile has the following fields set
	// 1. PosterFile.ID
	// 2. PosterFile.Type
	if posterFile.ID == "" || posterFile.Type == "" {
		actionValidate.SetError("Poster File ID and Type are required",
			"Ensure that the Poster File in the request body has an ID and Type",
			map[string]any{
				"PosterFile": posterFile,
			})
	}
	actionValidate.Complete()

	downloadFileName := api.MediaServer_GetFileDownloadName(posterFile)
	downloadFileName = collectionItem.Title + " - " + downloadFileName

	// // Response with a success message
	// logging.LOGGER.Info().Timestamp().Str("title", collectionItem.Title).Msgf("Downloading %s", downloadFileName)
	// api.Util_Response_SendJSON(w, ld, fmt.Sprintf("Downloaded %s successfully", downloadFileName))
	// return

	Err = api.CallDownloadAndUpdateCollectionImage(ctx, collectionItem, posterFile)
	if Err.Message != "" {
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}

	api.Util_Response_SendJSON(w, ld, fmt.Sprintf("Downloaded %s successfully", downloadFileName))
}
