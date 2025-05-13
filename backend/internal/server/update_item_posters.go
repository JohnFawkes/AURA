package mediaserver

import (
	"fmt"
	"net/http"
	"poster-setter/internal/config"
	"poster-setter/internal/database"
	"poster-setter/internal/logging"
	"poster-setter/internal/utils"
	"time"

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

	// Get the initial clientMessage from the updateItemStore
	UpdateItemStore.Lock()

	defer UpdateItemStore.Unlock()

	// Check if the ratingKey matches the stored ClientMessage.MediaItem.RatingKey
	if UpdateItemStore.ClientMessage.MediaItem.RatingKey != ratingKey {
		utils.SendErrorJSONResponse(w, http.StatusBadRequest, logging.ErrorLog{
			Err: fmt.Errorf("ratingKey does not match stored ClientMessage"),
			Log: logging.Log{Message: "RatingKey does not match stored ClientMessage",
				Elapsed: utils.ElapsedTime(startTime)}})
		return
	}

	// Get the clientMessage from the UpdateItemStore
	clientMessage := UpdateItemStore.ClientMessage

	// Make sure the clientMessage has at least the following fields set
	// For MediaItem:
	// 1. MediaItem.RatingKey
	// 2. MediaItem.Year
	// For Set:
	// 1. Set.ID
	// 2. Set.Files (at least one file)
	// For SelectedTypes:
	// 1. SelectedTypes (at least one type)
	if clientMessage.MediaItem.RatingKey == "" || clientMessage.MediaItem.Year == 0 {
		utils.SendErrorJSONResponse(w, http.StatusBadRequest, logging.ErrorLog{
			Err: fmt.Errorf("missing required fields in message.MediaItem"),
			Log: logging.Log{Message: "Missing required fields in message.MediaItem",
				Elapsed: utils.ElapsedTime(startTime)}})
		return
	} else if clientMessage.Set.ID == "" || len(clientMessage.Set.Files) == 0 {
		utils.SendErrorJSONResponse(w, http.StatusBadRequest, logging.ErrorLog{
			Err: fmt.Errorf("missing required fields in message.Set"),
			Log: logging.Log{Message: "Missing required fields in message.Set",
				Elapsed: utils.ElapsedTime(startTime)}})
		return
	} else if len(clientMessage.SelectedTypes) == 0 {
		utils.SendErrorJSONResponse(w, http.StatusBadRequest, logging.ErrorLog{
			Err: fmt.Errorf("missing required fields in message.SelectedTypes"),
			Log: logging.Log{Message: "Missing required fields in message.SelectedTypes",
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

	// Send the initial clientMessage to the client
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

	clientMessage.Set.Files = mediaserver_shared.FilterAndSortFiles(clientMessage.Set.Files, clientMessage.SelectedTypes)

	// Download the selected types from the Set.Files
	progressValue := 2
	progressStep := 95 / len(clientMessage.Set.Files)

	for _, file := range clientMessage.Set.Files {
		progressValue += progressStep
		downloadFileName := getFileDownloadName(file)
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
		logErr := mediaServer.DownloadAndUpdatePosters(clientMessage.MediaItem, file)
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

	}

	logErr := database.SaveClientMessage(clientMessage)
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
