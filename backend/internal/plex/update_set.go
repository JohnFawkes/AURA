package plex

import (
	"encoding/json"
	"fmt"
	"net/http"
	"poster-setter/internal/database"
	"poster-setter/internal/logging"
	"poster-setter/internal/modals"
	"poster-setter/internal/utils"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
)

var updateItemStore = struct {
	sync.Mutex
	clientMessage modals.ClientMessage
}{
	clientMessage: modals.ClientMessage{},
}

func GetUpdateSetFromClient(w http.ResponseWriter, r *http.Request) {
	logging.LOG.Trace("Getting Item from frontend to update in Plex")
	startTime := time.Now()

	var clientMessage modals.ClientMessage

	// Get the data from the POST request
	err := json.NewDecoder(r.Body).Decode(&clientMessage)
	if err != nil {
		utils.SendErrorJSONResponse(w, http.StatusBadRequest, logging.ErrorLog{Err: err,
			Log: logging.Log{Message: "Failed to decode JSON from client",
				Elapsed: utils.ElapsedTime(startTime)}})
		return
	}

	// Validate the clientMessage contains Plex.RatingKey
	if clientMessage.Plex.RatingKey == "" {
		utils.SendErrorJSONResponse(w, http.StatusBadRequest, logging.ErrorLog{
			Log: logging.Log{Message: "Missing Plex.RatingKey in clientMessage",
				Elapsed: utils.ElapsedTime(startTime)}})
		return
	}

	// Store the clientMessage in the updateItemStore
	updateItemStore.Lock()
	updateItemStore.clientMessage = clientMessage
	updateItemStore.Unlock()

	utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
		Status:  "success",
		Message: "Received message from client",
		Elapsed: utils.ElapsedTime(startTime),
	})
}

func UpdateSet(w http.ResponseWriter, r *http.Request) {
	logging.LOG.Trace("Updating Item in Plex with Set Poster")
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
			Log: logging.Log{Message: "Streaming unsupported",
				Elapsed: utils.ElapsedTime(startTime)}})
		return
	}

	// Get the initial clientMessage from the updateItemStore
	updateItemStore.Lock()
	defer updateItemStore.Unlock()

	// Check if the ratingKey matches the stored clientMessage.Plex.RatingKey
	if updateItemStore.clientMessage.Plex.RatingKey != ratingKey {
		utils.SendErrorJSONResponse(w, http.StatusBadRequest, logging.ErrorLog{
			Log: logging.Log{Message: "RatingKey does not match stored clientMessage",
				Elapsed: utils.ElapsedTime(startTime)}})
		return
	}

	// Get the clientMessage from the updateItemStore
	clientMessage := updateItemStore.clientMessage

	// Make sure the clientMessage has at least the following fields set
	// For Plex:
	// 1. Plex.RatingKey
	// 2. Plex.Year
	// For Set:
	// 1. Set.ID
	// 2. Set.Files (at least one file)
	// For SelectedTypes:
	// 1. SelectedTypes (at least one type)
	if clientMessage.Plex.RatingKey == "" || clientMessage.Plex.Year == 0 {
		utils.SendErrorJSONResponse(w, http.StatusBadRequest, logging.ErrorLog{
			Log: logging.Log{Message: "Missing required fields in message.Plex",
				Elapsed: utils.ElapsedTime(startTime)}})
		return
	} else if clientMessage.Set.ID == "" || len(clientMessage.Set.Files) == 0 {
		utils.SendErrorJSONResponse(w, http.StatusBadRequest, logging.ErrorLog{
			Log: logging.Log{Message: "Missing required fields in message.Set",
				Elapsed: utils.ElapsedTime(startTime)}})
		return
	} else if len(clientMessage.SelectedTypes) == 0 {
		utils.SendErrorJSONResponse(w, http.StatusBadRequest, logging.ErrorLog{
			Log: logging.Log{Message: "Missing required fields in message.SelectedTypes",
				Elapsed: utils.ElapsedTime(startTime)}})
		return
	}

	utils.SendSSEResponse(w, flusher, utils.SSEMessage{
		Response: utils.JSONResponse{
			Status:  "success",
			Message: "Starting update process",
			Elapsed: utils.ElapsedTime(startTime),
		},
		Progress: utils.SSEProgress{
			Value:    1,
			Text:     "Starting update process",
			NextStep: "Downloading files",
		},
	})

	clientMessage.Set.Files = FilterAndSortFiles(clientMessage.Set.Files, clientMessage.SelectedTypes)

	// Download the selected types from the Set.Files
	progressValue := 1
	progressStep := 80 / len(clientMessage.Set.Files)

	for _, file := range clientMessage.Set.Files {
		progressValue += progressStep
		utils.SendSSEResponse(w, flusher, utils.SSEMessage{
			Response: utils.JSONResponse{
				Status:  "success",
				Message: "Downloading images",
				Elapsed: utils.ElapsedTime(startTime),
			},
			Progress: utils.SSEProgress{
				Value:    progressValue,
				Text:     "Downloading images",
				NextStep: fmt.Sprintf("Downloading %s", getFileDownloadName(file)),
			},
		})
		logging.LOG.Debug(fmt.Sprintf("Downloading %s", getFileDownloadName(file)))

		// Download and update the set
		logErr := DownloadAndUpdateSet(clientMessage.Plex, file)
		if logErr.Err != nil {
			logging.LOG.ErrorWithLog(logErr)
			continue
		}
	}

	if clientMessage.AutoDownload {
		logging.LOG.Trace("AutoDownload is set to true, saving to database")
		logErr := database.SaveClientMessage(clientMessage)
		if logErr.Err != nil {
			utils.SendErrorJSONResponse(w, http.StatusInternalServerError, logErr)
			return
		}
		logging.LOG.Info(fmt.Sprintf("Added posters for '%s' to the database", clientMessage.Plex.Title))
	} else {
		logging.LOG.Trace("AutoDownload is set to false, not saving to database")
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
