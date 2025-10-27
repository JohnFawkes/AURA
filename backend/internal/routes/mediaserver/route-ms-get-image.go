package routes_ms

import (
	"aura/internal/api"
	"aura/internal/logging"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

func GetImage(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	Err := logging.NewStandardError()

	ratingKey := chi.URLParam(r, "ratingKey")
	imageType := chi.URLParam(r, "imageType")
	if ratingKey == "" || imageType == "" {
		Err.Message = "Missing rating key or image type in URL"
		Err.HelpText = "Ensure the URL contains both rating key and image type parameters."
		Err.Details = map[string]any{
			"error":     "Rating key or image type is empty",
			"ratingKey": ratingKey,
			"imageType": imageType,
			"request":   r.URL.Path,
		}
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		return
	} else if imageType != "poster" && imageType != "backdrop" {
		Err.Message = "Invalid image type"
		Err.HelpText = "Image type must be either 'poster' or 'backdrop'."
		Err.Details = fmt.Sprintf("Received image type: %s", imageType)
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		return
	}

	mediaServer, Err := api.GetMediaServerInterface(api.Config_MediaServer{})
	if Err.Message != "" {
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
	}

	// If the image does not exist, then get it from the media server
	imageData, Err := mediaServer.FetchImageFromMediaServer(ratingKey, imageType)
	if Err.Message != "" {
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		return
	}

	// Set the content type for the response
	w.Header().Set("Content-Type", "image/jpeg")
	// Write the image data to the response
	w.WriteHeader(http.StatusOK)
	w.Write(imageData)
}
