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
	Err := logging.NewStandardError()

	// Parse the request body to get posterFile & mediaItem
	var SaveItem modals.DBSavedItem

	if err := json.NewDecoder(r.Body).Decode(&SaveItem); err != nil {
		Err.Message = "Failed to decode request body"
		Err.HelpText = "Ensure the request body is a valid JSON object."
		Err.Details = fmt.Sprintf("Error: %s", err.Error())
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}

	// Validate the request body
	if SaveItem.MediaItemID == "" || SaveItem.PosterSet.ID == "" {
		Err.Message = "Missing required fields"
		Err.HelpText = "Ensure the request body contains both MediaItemID and PosterSet.ID."
		Err.Details = fmt.Sprintf("Received MediaItemID: %s, PosterSet.ID: %s", SaveItem.MediaItemID, SaveItem.PosterSet.ID)
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}

	Err = database.SaveItemInDB(SaveItem)
	if Err.Message != "" {
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}

	// Respond with a success message
	utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
		Status:  "success",
		Elapsed: utils.ElapsedTime(startTime),
		Data:    fmt.Sprintf("Item with ID %s added successfully", SaveItem.MediaItemID),
	})
}
