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

	var requestBody struct {
		Item modals.DBMediaItemWithPosterSets
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		utils.SendErrorJSONResponse(w, http.StatusBadRequest, logging.ErrorLog{
			Err: err,
			Log: logging.Log{Message: "Failed to parse request body",
				Elapsed: utils.ElapsedTime(startTime)}})
		return
	}

	item := requestBody.Item

	// Get the latest item from DB incase it has been updated
	allItems, logErr := database.GetAllItemsFromDatabase()
	if logErr.Err != nil {
		logging.LOG.ErrorWithLog(logErr)
		utils.SendErrorJSONResponse(w, http.StatusInternalServerError, logErr)
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
		logErr := logging.ErrorLog{
			Err: fmt.Errorf("item with ID %s not found in database", item.MediaItemID),
			Log: logging.Log{Message: "Item not found in database", Elapsed: utils.ElapsedTime(startTime)},
		}
		logging.LOG.ErrorWithLog(logErr)
		utils.SendErrorJSONResponse(w, http.StatusNotFound, logErr)
		return
	}

	results := CheckItemForAutodownload(dbSavedItem)

	// If no warnings, send a success response
	utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
		Status:  "success",
		Message: "Force recheck completed successfully",
		Elapsed: utils.ElapsedTime(startTime),
		Data:    results,
	})
}
