package routes_ms

import (
	"aura/internal/api"
	"aura/internal/logging"
	"net/http"
)

func GetImage(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Get Image From Media Server", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	actionGetQueryParams := logAction.AddSubAction("Get Query Params", logging.LevelDebug)
	ratingKey := r.URL.Query().Get("ratingKey")
	imageType := r.URL.Query().Get("imageType")
	if ratingKey == "" || imageType == "" {
		actionGetQueryParams.SetError("Missing Query Parameters", "One or more required query parameters are missing",
			map[string]any{
				"ratingKey": ratingKey,
				"imageType": imageType,
			})
		api.Util_Response_SendJSON(w, ld, nil)
		return
	} else if imageType != "poster" && imageType != "backdrop" {
		actionGetQueryParams.SetError("Invalid Query Parameters", "Image type must be either 'poster' or 'backdrop'",
			map[string]any{
				"ratingKey": ratingKey,
				"imageType": imageType,
			})
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}
	actionGetQueryParams.Complete()

	// If the image does not exist, then get it from the media server
	imageData, Err := api.CallFetchImageFromMediaServer(ctx, ratingKey, imageType)
	if Err.Message != "" {
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}

	// Set the content type for the response
	w.Header().Set("Content-Type", "image/jpeg")
	// Write the image data to the response
	w.WriteHeader(http.StatusOK)
	w.Write(imageData)
}
