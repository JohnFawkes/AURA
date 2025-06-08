package download

import (
	"aura/internal/logging"
	"aura/internal/modals"
	"aura/internal/utils"
	"encoding/json"
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
	warningMessages := CheckItemForAutodownload(item)

	if len(warningMessages) > 0 {
		utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
			Status:  "warning",
			Message: "Force recheck completed with warnings",
			Elapsed: utils.ElapsedTime(startTime),
			Data:    warningMessages,
		})
		return
	}

	// If no warnings, send a success response

	utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
		Status:  "success",
		Message: "Force recheck completed successfully",
		Elapsed: utils.ElapsedTime(startTime),
		Data:    "success",
	})
}
