package mediaserver

import (
	"fmt"
	"net/http"
	"os"
	"poster-setter/internal/config"
	"poster-setter/internal/database"
	"poster-setter/internal/logging"
	"poster-setter/internal/modals"
	"poster-setter/internal/utils"
	"time"

	"poster-setter/internal/server/emby_jellyfin"
	"poster-setter/internal/server/plex"
	mediaserver_shared "poster-setter/internal/server/shared"

	"github.com/go-chi/chi/v5"
)

func UpdateItemPosters(w http.ResponseWriter, r *http.Request) {
	logging.LOG.Trace(r.URL.Path)
	startTime := time.Now()

	// Get the ratingKey from the URL
	ratingKey := chi.URLParam(r, "ratingKey")

	// Set the correct headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Set the response writer to flush the headers immediately
	flusher, ok := w.(http.Flusher)
	if !ok {
		utils.SendErrorJSONResponse(w, http.StatusInternalServerError, logging.ErrorLog{
			Err: fmt.Errorf("failed to set response writer to flush"),
			Log: logging.Log{Message: "Streaming unsupported",
				Elapsed: utils.ElapsedTime(startTime)}})
		return
	}

	// Get the initial SavedSet from the updateItemStore
	UpdateItemStore.Lock()

	defer UpdateItemStore.Unlock()

	// Check if the ratingKey matches the stored SavedSet.MediaItem.RatingKey
	//if UpdateItemStore.SavedSet.MediaItem.RatingKey != ratingKey {
	if UpdateItemStore.SavedSet.MediaItem.RatingKey != ratingKey {
		utils.SendErrorJSONResponse(w, http.StatusBadRequest, logging.ErrorLog{
			Err: fmt.Errorf("ratingKey does not match stored SavedSet"),
			Log: logging.Log{Message: "RatingKey does not match stored SavedSet",
				Elapsed: utils.ElapsedTime(startTime)}})
		return
	}

	// Get the SavedSet from the UpdateItemStore
	savedSet := UpdateItemStore.SavedSet

	// Make sure the SavedSet has at least the following fields set
	// For MediaItem:
	// 1. MediaItem.RatingKey
	// 2. MediaItem.Year
	// For Set:
	// 1. Set.ID
	// 2. Set.Files (at least one file)
	// For SelectedTypes:
	// 1. SelectedTypes (at least one type)
	if savedSet.MediaItem.RatingKey == "" || savedSet.MediaItem.Year == 0 {
		utils.SendErrorJSONResponse(w, http.StatusBadRequest, logging.ErrorLog{
			Err: fmt.Errorf("missing required fields in message.MediaItem"),
			Log: logging.Log{Message: "Missing required fields in message.MediaItem",
				Elapsed: utils.ElapsedTime(startTime)}})
		return
	}
	for _, set := range savedSet.Sets {
		if set.ID == "" || len(set.Set.Files) == 0 {
			utils.SendErrorJSONResponse(w, http.StatusBadRequest, logging.ErrorLog{
				Err: fmt.Errorf("missing required fields in message.Set"),
				Log: logging.Log{Message: "Missing required fields in message.Set",
					Elapsed: utils.ElapsedTime(startTime)}})
			return
		} else if len(set.SelectedTypes) == 0 {
			utils.SendErrorJSONResponse(w, http.StatusBadRequest, logging.ErrorLog{
				Err: fmt.Errorf("missing required fields in message.SelectedTypes"),
				Log: logging.Log{Message: "Missing required fields in message.SelectedTypes",
					Elapsed: utils.ElapsedTime(startTime)}})
			return
		}
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

	// Send the initial SavedSet to the client
	utils.SendSSEResponse(w, flusher, utils.SSEMessage{
		Response: utils.JSONResponse{
			Status:  "success",
			Message: "Starting update process",
			Elapsed: utils.ElapsedTime(startTime),
		},
		Progress: utils.SSEProgress{
			Value:    2,
			Text:     "Starting update process",
			NextStep: "Downloading files",
		},
	})

	savedSet.Sets[0].Set.Files = mediaserver_shared.FilterAndSortFiles(savedSet.Sets[0].Set.Files, savedSet.Sets[0].SelectedTypes)

	// Download the selected types from the Set.Files
	progressValue := 2
	progressStep := 95 / len(savedSet.Sets[0].Set.Files)

	for _, file := range savedSet.Sets[0].Set.Files {
		progressValue += progressStep
		downloadFileName := GetFileDownloadName(file)
		utils.SendSSEResponse(w, flusher, utils.SSEMessage{
			Response: utils.JSONResponse{
				Status:  "success",
				Message: "Downloading images",
				Elapsed: utils.ElapsedTime(startTime),
			},
			Progress: utils.SSEProgress{
				Value:    progressValue,
				Text:     "Downloading images",
				NextStep: fmt.Sprintf("Downloading %s", downloadFileName),
			},
		})
		logging.LOG.Debug(fmt.Sprintf("Downloading %s", downloadFileName))
		// Download the image from the media server
		logErr := mediaServer.DownloadAndUpdatePosters(savedSet.MediaItem, file)
		if logErr.Err != nil {
			utils.SendSSEResponse(w, flusher, utils.SSEMessage{
				Response: utils.JSONResponse{
					Status:  "warning",
					Message: "Downloading images",
					Elapsed: utils.ElapsedTime(startTime),
				},
				Progress: utils.SSEProgress{
					Value:    progressValue,
					Text:     downloadFileName,
					NextStep: logErr.Log.Message,
				},
			})
			logging.LOG.ErrorWithLog(logErr)
			continue
		}
		deleteTempImageForNextLoad(file, savedSet.MediaItem.RatingKey)
	}

	logErr := database.SaveSavedSet(savedSet)
	if logErr.Err != nil {
		utils.SendErrorJSONResponse(w, http.StatusInternalServerError, logErr)
		return
	}

	utils.SendSSEResponse(w, flusher, utils.SSEMessage{
		Response: utils.JSONResponse{
			Status:  "complete",
			Message: "Update process completed",
			Elapsed: utils.ElapsedTime(startTime),
		},
		Progress: utils.SSEProgress{
			Value:    100,
			Text:     "Update process completed",
			NextStep: "",
		},
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

		// Delete the temporary image file
		fileName := fmt.Sprintf("%s_%s.jpg", ratingKey, file.Type)
		filePath := fmt.Sprintf("%s/%s", tmpFolder, fileName)
		exists := utils.CheckIfImageExists(filePath)
		if exists {
			logging.LOG.Trace(fmt.Sprintf("Deleting temporary image %s", fileName))
			err := os.Remove(filePath)
			if err != nil {
				logging.LOG.Error(fmt.Sprintf("Failed to delete temporary image %s: %s", fileName, err.Error()))
			}
		} else {
			logging.LOG.Trace(fmt.Sprintf("Temporary image %s does not exist", fileName))
		}
	}
}
