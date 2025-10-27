package routes_ms

import (
	"aura/internal/api"
	"aura/internal/logging"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func DownloadAndUpdate(w http.ResponseWriter, r *http.Request) {
	logging.LOG.Trace(r.URL.Path)
	startTime := time.Now()
	Err := logging.NewStandardError()

	// Parse the request body to get posterFile & mediaItem
	var requestBody struct {
		PosterFile api.PosterFile `json:"PosterFile"`
		MediaItem  api.MediaItem  `json:"MediaItem"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		Err.Message = "Failed to decode request body"
		Err.HelpText = "Ensure the request body is a valid JSON object."
		Err.Details = map[string]any{
			"error": err.Error(),
			"body":  r.Body,
		}
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		return
	}

	posterFile := requestBody.PosterFile
	mediaItem := requestBody.MediaItem

	// Make sure that the mediaItem has the following fields set
	// 1. MediaItem.RatingKey
	if mediaItem.RatingKey == "" {
		Err.Message = "mediaItem.RatingKey is required"
		Err.HelpText = "Ensure the mediaItem.RatingKey is provided in the request body."
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		return
	}

	// Make sure that the posterFile has the following fields set
	// 1. PosterFile.ID
	// 2. PosterFile.Type
	if posterFile.ID == "" || posterFile.Type == "" {
		Err.Message = "PosterFile.ID and PosterFile.Type are required"
		Err.HelpText = "Ensure the PosterFile.ID and PosterFile.Type are provided in the request body."
		Err.Details = map[string]any{
			"PosterFile.ID":   posterFile.ID,
			"PosterFile.Type": posterFile.Type,
		}
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		return
	}

	downloadFileName := api.MediaServer_GetFileDownloadName(posterFile)
	logging.LOG.Debug(fmt.Sprintf("Downloading %s", downloadFileName))

	// // Respond with a success message
	// time.Sleep(2 * time.Second) // Simulate download time
	// Util_Response_SendJson(w, http.StatusOK, JSONResponse{
	// 	Status:  "success",
	// 	Elapsed: Util_ElapsedTime(startTime),
	// 	Data:    fmt.Sprintf("Downloaded %s successfully", downloadFileName),
	// })
	// return

	Err = api.CallDownloadAndUpdatePosters(mediaItem, posterFile)
	if Err.Message != "" {
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		return
	}

	api.DeleteTempImageForNextLoad(posterFile, mediaItem.RatingKey)

	// Respond with a success message
	api.Util_Response_SendJson(w, http.StatusOK, api.JSONResponse{
		Status:  "success",
		Elapsed: api.Util_ElapsedTime(startTime),
		Data:    fmt.Sprintf("Downloaded %s successfully", downloadFileName),
	})
}
