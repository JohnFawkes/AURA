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
	SavedSet modals.Database_SavedSet
}{
	SavedSet: modals.Database_SavedSet{},
}

func GetUpdateSetFromClient(w http.ResponseWriter, r *http.Request) {
	logging.LOG.Trace(fmt.Sprintf("Getting set from client to update in %s", config.Global.MediaServer.Type))
	startTime := time.Now()

	var SavedSet modals.Database_SavedSet

	// Get the data from the POST request
	err := json.NewDecoder(r.Body).Decode(&SavedSet)
	if err != nil {
		utils.SendErrorJSONResponse(w, http.StatusBadRequest, logging.ErrorLog{Err: err,
			Log: logging.Log{Message: "Failed to decode JSON from client",
				Elapsed: utils.ElapsedTime(startTime)}})
		return
	}

	// Validate the SavedSet contains MediaItem.RatingKey
	if SavedSet.MediaItem.RatingKey == "" {
		utils.SendErrorJSONResponse(w, http.StatusBadRequest, logging.ErrorLog{
			Err: fmt.Errorf("missing MediaItem.RatingKey"),
			Log: logging.Log{Message: "Missing MediaItem.RatingKey in SavedSet",
				Elapsed: utils.ElapsedTime(startTime)}})
		return
	}

	// Store the SavedSet in the UpdateItemStore
	UpdateItemStore.Lock()
	UpdateItemStore.SavedSet = SavedSet
	UpdateItemStore.Unlock()

	utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
		Status:  "success",
		Message: "Received message from client",
		Elapsed: utils.ElapsedTime(startTime),
	})
}
