package mediaserver

import (
	"aura/internal/config"
	"aura/internal/logging"
	"aura/internal/modals"
	"aura/internal/server/emby_jellyfin"
	"aura/internal/server/plex"
	mediaserver_shared "aura/internal/server/shared"
	"aura/internal/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

func DownloadAndUpdate(w http.ResponseWriter, r *http.Request) {
	logging.LOG.Trace(r.URL.Path)
	startTime := time.Now()
	Err := logging.NewStandardError()

	// Parse the request body to get posterFile & mediaItem
	var requestBody struct {
		PosterFile modals.PosterFile `json:"PosterFile"`
		MediaItem  modals.MediaItem  `json:"MediaItem"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {

		Err.Message = "Failed to decode request body"
		Err.HelpText = "Ensure the request body is a valid JSON object."
		Err.Details = fmt.Sprintf("Error: %s", err.Error())
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}

	posterFile := requestBody.PosterFile
	mediaItem := requestBody.MediaItem

	// Make sure that the mediaItem has the following fields set
	// 1. MediaItem.RatingKey
	if mediaItem.RatingKey == "" {

		Err.Message = "mediaItem.RatingKey is required"
		Err.HelpText = "Ensure the mediaItem.RatingKey is provided in the request body."
		Err.Details = "mediaItem.RatingKey is required to identify the media item."
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}

	// Make sure that the posterFile has the following fields set
	// 1. PosterFile.ID
	// 2. PosterFile.Type
	if posterFile.ID == "" || posterFile.Type == "" {

		Err.Message = "PosterFile.ID and PosterFile.Type are required"
		Err.HelpText = "Ensure the PosterFile.ID and PosterFile.Type are provided in the request body."
		Err.Details = fmt.Sprintf("Received PosterFile.ID: %s, PosterFile.Type: %s", posterFile.ID, posterFile.Type)
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}

	var mediaServer mediaserver_shared.MediaServer
	switch config.Global.MediaServer.Type {
	case "Plex":
		mediaServer = &mediaserver_shared.PlexServer{}

	case "Emby", "Jellyfin":
		mediaServer = &mediaserver_shared.EmbyJellyServer{}
	default:

		Err.Message = "Unsupported media server type"
		Err.HelpText = fmt.Sprintf("The media server type '%s' is not supported.", config.Global.MediaServer.Type)
		Err.Details = fmt.Sprintf("Unsupported media server type: %s", config.Global.MediaServer.Type)
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}

	downloadFileName := GetFileDownloadName(posterFile)
	logging.LOG.Debug(fmt.Sprintf("Downloading %s", downloadFileName))

	// Respond with a success message
	// time.Sleep(150 * time.Millisecond) // Simulate a delay for the download process
	// utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
	// 	Status:  "success",
	// 	Elapsed: utils.ElapsedTime(startTime),
	// 	Data:    fmt.Sprintf("Downloaded %s successfully", downloadFileName),
	// })
	// return

	Err = mediaServer.DownloadAndUpdatePosters(mediaItem, posterFile)
	if Err.Message != "" {
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}

	deleteTempImageForNextLoad(posterFile, mediaItem.RatingKey)

	// Respond with a success message
	utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
		Status:  "success",
		Elapsed: utils.ElapsedTime(startTime),
		Data:    fmt.Sprintf("Downloaded %s successfully", downloadFileName),
	})
}

func deleteTempImageForNextLoad(file modals.PosterFile, ratingKey string) {
	if file.Type == "poster" || file.Type == "backdrop" {
		var tmpFolder string
		switch config.Global.MediaServer.Type {
		case "Plex":
			tmpFolder = plex.PlexTempImageFolder
		case "Emby", "Jellyfin":
			tmpFolder = emby_jellyfin.EmbyJellyTempImageFolder
		default:
			logging.LOG.Error(fmt.Sprintf("Unsupported media server type: %s", config.Global.MediaServer.Type))
			return
		}

		// Delete the poster and backdrop temporary image file
		fileName := fmt.Sprintf("%s_%s.jpg", ratingKey, file.Type)
		filePath := fmt.Sprintf("%s/%s", tmpFolder, fileName)
		exists := utils.CheckIfImageExists(filePath)
		if exists {
			logging.LOG.Trace(fmt.Sprintf("Deleting temporary image %s", fileName))
			err := os.Remove(filePath)
			if err != nil {
				logging.LOG.Error(fmt.Sprintf("Failed to delete temporary image %s: %s", fileName, err.Error()))
			}
		}

		otherFile := "backdrop"
		if file.Type == "backdrop" {
			otherFile = "poster"
		}
		// Delete the other temporary image file
		otherFileName := fmt.Sprintf("%s_%s.jpg", ratingKey, otherFile)
		otherFilePath := fmt.Sprintf("%s/%s", tmpFolder, otherFileName)
		exists = utils.CheckIfImageExists(otherFilePath)
		if exists {
			logging.LOG.Trace(fmt.Sprintf("Deleting temporary image %s", otherFileName))
			err := os.Remove(otherFilePath)
			if err != nil {
				logging.LOG.Error(fmt.Sprintf("Failed to delete temporary image %s: %s", otherFileName, err.Error()))
			}
		}
	}
}
