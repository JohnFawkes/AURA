package download

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

func ForceRecheckItem(w http.ResponseWriter, r *http.Request) {
	logging.LOG.Trace(r.URL.Path)
	startTime := time.Now()
	Err := logging.NewStandardError()

	var requestBody struct {
		Item modals.DBMediaItemWithPosterSets
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {

		Err.Message = "Failed to decode request body"
		Err.HelpText = "Ensure the request body is a valid JSON object matching the expected structure."
		Err.Details = fmt.Sprintf("Request Body: %s", r.Body)
		logging.LOG.ErrorWithLog(Err)
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}

	item := requestBody.Item

	// Get the latest item from DB incase it has been updated
	allItems, Err := database.GetAllItemsFromDatabase()
	if Err.Message != "" {
		logging.LOG.ErrorWithLog(Err)
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}

	// Find the item in the database
	var dbSavedItem modals.DBMediaItemWithPosterSets
	for _, dbItem := range allItems {
		if dbItem.MediaItemID == item.MediaItemID {
			dbSavedItem = dbItem
			break
		}
	}
	if dbSavedItem.MediaItemID == "" {

		Err.Message = "Item not found in database"
		Err.HelpText = "Ensure the item exists in the database before attempting a force recheck."
		Err.Details = fmt.Sprintf("Item ID: %s", item.MediaItemID)
		logging.LOG.ErrorWithLog(Err)
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}

	results := CheckItemForAutodownload(dbSavedItem)

	// If no warnings, send a success response
	utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
		Status:  "success",
		Elapsed: utils.ElapsedTime(startTime),
		Data:    results,
	})
}
