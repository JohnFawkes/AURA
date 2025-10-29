package routes_ms

import (
	"aura/internal/api"
	"aura/internal/logging"
	"fmt"
	"net/http"
)

func DownloadAndUpdate(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Download and Update", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	// Parse the request body to get posterFile & mediaItem
	var requestBody struct {
		PosterFile api.PosterFile `json:"PosterFile"`
		MediaItem  api.MediaItem  `json:"MediaItem"`
	}
	Err := api.DecodeRequestBodyJSON(ctx, r.Body, &requestBody, "requestBody")
	if Err.Message != "" {
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}

	posterFile := requestBody.PosterFile
	mediaItem := requestBody.MediaItem
	actionValidate := logAction.AddSubAction("Validate Request Body Fields", logging.LevelDebug)
	// Make sure that the mediaItem has the following fields set
	// 1. MediaItem.RatingKey
	if mediaItem.RatingKey == "" {
		actionValidate.SetError("Media Item Rating Key is required",
			"Ensure that the Media Item in the request body has a Rating Key",
			map[string]any{
				"MediaItem": mediaItem,
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

	// Response with a success message
	// time.Sleep(2 * time.Second) // Simulate download time
	// api.Util_Response_SendJSON(w, ld, fmt.Sprintf("Downloaded %s successfully", downloadFileName))

	Err = api.CallDownloadAndUpdatePosters(ctx, mediaItem, posterFile)
	if Err.Message != "" {
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}

	api.DeleteTempImageForNextLoad(ctx, posterFile, mediaItem.RatingKey)

	api.Util_Response_SendJSON(w, ld, fmt.Sprintf("Downloaded and updated %s successfully", downloadFileName))
}
