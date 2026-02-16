package routes_download

import (
	"aura/logging"
	"aura/mediaserver"
	"aura/models"
	"aura/utils/httpx"
	"net/http"
)

func DownloadCollectionImage(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Download Collection Image", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	// Parse the request body
	var req struct {
		ImageFile      models.ImageFile      `json:"image_file"`
		CollectionItem models.CollectionItem `json:"collection_item"`
	}
	Err := httpx.DecodeRequestBodyToJSON(ctx, r.Body, &req, "Download Collection Image - Decode Request Body")
	if Err.Message != "" {
		httpx.SendResponse(w, ld, nil)
		return
	}

	actionValidate := logAction.AddSubAction("Validate Request Body Fields", logging.LevelDebug)
	// Make sure that the collectionItem has the following fields set
	// 1. CollectionItem.RatingKey
	// 2. CollectionItem.LibraryTitle
	if req.CollectionItem.RatingKey == "" || req.CollectionItem.LibraryTitle == "" {
		actionValidate.SetError("Invalid Collection Item Data", "RatingKey and LibraryTitle are required in Collection Item", map[string]any{
			"rating_key":    req.CollectionItem.RatingKey,
			"library_title": req.CollectionItem.LibraryTitle,
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
	Err = mediaserver.ApplyCollectionImage(ctx, &req.CollectionItem, req.ImageFile)
	if Err.Message != "" {
		httpx.SendResponse(w, ld, nil)
		return
	}

	httpx.SendResponse(w, ld, nil)
}
