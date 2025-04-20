package mediaserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"poster-setter/internal/config"
	"poster-setter/internal/logging"
	"poster-setter/internal/modals"
	"poster-setter/internal/utils"
	"sync"
	"time"
)

var UpdateItemStore = struct {
	sync.Mutex
	ClientMessage modals.ClientMessage
}{
	ClientMessage: modals.ClientMessage{},
}

func GetUpdateSetFromClient(w http.ResponseWriter, r *http.Request) {
	logging.LOG.Trace(fmt.Sprintf("Getting set from client to update in %s", config.Global.MediaServer.Type))
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

	// Validate the clientMessage contains MediaItem.RatingKey
	if clientMessage.MediaItem.RatingKey == "" {
		utils.SendErrorJSONResponse(w, http.StatusBadRequest, logging.ErrorLog{
			Err: fmt.Errorf("missing MediaItem.RatingKey"),
			Log: logging.Log{Message: "Missing MediaItem.RatingKey in clientMessage",
				Elapsed: utils.ElapsedTime(startTime)}})
		return
	}

	// Store the clientMessage in the UpdateItemStore
	UpdateItemStore.Lock()
	UpdateItemStore.ClientMessage = clientMessage
	UpdateItemStore.Unlock()

	utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
		Status:  "success",
		Message: "Received message from client",
		Elapsed: utils.ElapsedTime(startTime),
	})
}
