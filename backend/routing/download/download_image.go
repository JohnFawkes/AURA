package routes_download

import (
	"aura/logging"
	"aura/mediaserver"
	"aura/models"
	"aura/utils"
	"aura/utils/httpx"
	"fmt"
	"net/http"
)

func DownloadImage(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Download Image", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	// Parse the request body
	var req struct {
		ImageFile models.ImageFile `json:"image_file"`
		MediaItem models.MediaItem `json:"media_item"`
	}
	Err := httpx.DecodeRequestBodyToJSON(ctx, r.Body, &req, "Download Image - Decode Request Body")
	if Err.Message != "" {
		httpx.SendResponse(w, ld, nil)
		return
	}

	actionValidate := logAction.AddSubAction("Validate Request Body Fields", logging.LevelDebug)
	// Make sure that the mediaItem has the following fields set
	// 1. MediaItem.RatingKey
	// 2. MediaItem.LibraryTitle
	if req.MediaItem.RatingKey == "" || req.MediaItem.LibraryTitle == "" {
		actionValidate.SetError("Invalid Media Item Data", "RatingKey and LibraryTitle are required in Media Item", map[string]any{
			"rating_key":    req.MediaItem.RatingKey,
			"library_title": req.MediaItem.LibraryTitle,
		})
		actionValidate.Complete()
		httpx.SendResponse(w, ld, nil)
		return
	}

	// Make sure that the ImageFile has the following fields set
	// 1. ImageFile.ID
	// 2. ImageFile.Type
	if req.ImageFile.ID == "" || req.ImageFile.Type == "" {
		actionValidate.SetError("Invalid Image File Data", "ID and Type are required in Image File", map[string]any{
			"id":   req.ImageFile.ID,
			"type": req.ImageFile.Type,
		})
		actionValidate.Complete()
		httpx.SendResponse(w, ld, nil)
		return
	}
	actionValidate.Complete()

	// Make the download and apply the image
	Err = mediaserver.DownloadApplyImageToMediaItem(ctx, &req.MediaItem, req.ImageFile)
	if Err.Message != "" {
		httpx.SendResponse(w, ld, nil)
		return
	}

	httpx.SendResponse(w, ld, fmt.Sprintf("Sucessfully downloaded %s", utils.GetFileDownloadName(req.MediaItem.Title, req.ImageFile)))
}
