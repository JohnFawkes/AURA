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

	// Parse the request body to get posterFile & mediaItem
	var requestBody struct {
		PosterFile modals.PosterFile `json:"posterFile"`
		MediaItem  modals.MediaItem  `json:"mediaItem"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		utils.SendErrorJSONResponse(w, http.StatusBadRequest, logging.ErrorLog{
			Err: err,
			Log: logging.Log{Message: "Failed to parse request body",
				Elapsed: utils.ElapsedTime(startTime)}})
		return
	}

	posterFile := requestBody.PosterFile
	mediaItem := requestBody.MediaItem

	// Make sure that the mediaItem has the following fields set
	// 1. MediaItem.RatingKey
	if mediaItem.RatingKey == "" {
		utils.SendErrorJSONResponse(w, http.StatusBadRequest, logging.ErrorLog{
			Err: fmt.Errorf("mediaItem.RatingKey is required"),
			Log: logging.Log{Message: "MediaItem.RatingKey is required",
				Elapsed: utils.ElapsedTime(startTime)}})
		return
	}

	// Make sure that the posterFile has the following fields set
	// 1. PosterFile.ID
	// 2. PosterFile.Type
	if posterFile.ID == "" || posterFile.Type == "" {
		utils.SendErrorJSONResponse(w, http.StatusBadRequest, logging.ErrorLog{
			Err: fmt.Errorf("posterFile.ID and posterFile.Type are required"),
			Log: logging.Log{Message: "PosterFile.ID and PosterFile.Type are required",
				Elapsed: utils.ElapsedTime(startTime)}})
		return
	}

	var mediaServer mediaserver_shared.MediaServer
	switch config.Global.MediaServer.Type {
	case "Plex":
		mediaServer = &mediaserver_shared.PlexServer{}

	case "Emby", "Jellyfin":
		mediaServer = &mediaserver_shared.EmbyJellyServer{}
	default:
		logErr := logging.ErrorLog{
			Err: fmt.Errorf("unsupported media server type: %s", config.Global.MediaServer.Type),
			Log: logging.Log{Message: fmt.Sprintf("Unsupported media server type: %s", config.Global.MediaServer.Type),
				Elapsed: utils.ElapsedTime(startTime),
			},
		}
		utils.SendErrorJSONResponse(w, http.StatusBadRequest, logErr)
		return
	}

	downloadFileName := GetFileDownloadName(posterFile)
	logging.LOG.Debug(fmt.Sprintf("Downloading %s", downloadFileName))

	logErr := mediaServer.DownloadAndUpdatePosters(mediaItem, posterFile)
	if logErr.Err != nil {
		utils.SendErrorJSONResponse(w, http.StatusInternalServerError, logErr)
		return
	}

	deleteTempImageForNextLoad(posterFile, mediaItem.RatingKey)

	// Respond with a success message
	utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
		Status:  "success",
		Message: "Downloaded and Updated successfully",
		Elapsed: utils.ElapsedTime(startTime),
		Data:    "downloaded some file successfully",
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
