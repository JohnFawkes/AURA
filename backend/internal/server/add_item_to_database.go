package mediaserver

import (
	"aura/internal/database"
	"aura/internal/logging"
	"aura/internal/modals"
	"aura/internal/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func AddItemToDatabase(w http.ResponseWriter, r *http.Request) {
	logging.LOG.Trace(r.URL.Path)
	startTime := time.Now()

	// Parse the request body to get posterFile & mediaItem
	var SaveItem modals.DBSavedItem

	if err := json.NewDecoder(r.Body).Decode(&SaveItem); err != nil {
		utils.SendErrorJSONResponse(w, http.StatusBadRequest, logging.ErrorLog{
			Err: err,
			Log: logging.Log{Message: "Failed to parse request body",
				Elapsed: utils.ElapsedTime(startTime)}})
		return
	}

	// Validate the request body
	if SaveItem.MediaItemID == "" || SaveItem.PosterSet.ID == "" {
		utils.SendErrorJSONResponse(w, http.StatusBadRequest, logging.ErrorLog{
			Err: fmt.Errorf("invalid request body"),
			Log: logging.Log{Message: "Invalid request body",
				Elapsed: utils.ElapsedTime(startTime)}})
		return
	}

	logErr := database.SaveItemInDB(SaveItem)
	if logErr.Err != nil {
		utils.SendErrorJSONResponse(w, http.StatusInternalServerError, logErr)
		return
	}

	// Respond with a success message
	utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
		Status:  "success",
		Message: "Downloaded and Updated successfully",
		Elapsed: utils.ElapsedTime(startTime),
		Data:    "downloaded some file successfully",
	})
}
