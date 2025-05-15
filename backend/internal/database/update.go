package database

import (
	"encoding/json"
	"fmt"
	"net/http"
	"poster-setter/internal/logging"
	"poster-setter/internal/modals"
	"poster-setter/internal/utils"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

func UpdateSavedSetTypesForItem(w http.ResponseWriter, r *http.Request) {
	logging.LOG.Debug(r.URL.Path)
	startTime := time.Now()

	// Get the ratingKey from the URL
	ratingKey := chi.URLParam(r, "ratingKey")
	if ratingKey == "" {
		utils.SendErrorJSONResponse(w, http.StatusBadRequest, logging.ErrorLog{
			Err: fmt.Errorf("missing ratingKey in request"),
			Log: logging.Log{
				Message: "Missing ratingKey in request",
			},
		})
		return
	}

	// Get the request body
	// Define a struct to match the expected JSON object
	var requestBody struct {
		SelectedTypes []string `json:"selectedTypes"`
		Autodownload  bool     `json:"autoDownload"`
	}
	// Decode the request body into the struct
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		utils.SendErrorJSONResponse(w, http.StatusBadRequest, logging.ErrorLog{
			Err: err,
			Log: logging.Log{
				Message: "Failed to decode request body",
			},
		})
		return
	}

	selectedTypes := requestBody.SelectedTypes
	autoDownload := requestBody.Autodownload
	logging.LOG.Debug(fmt.Sprintf("Updating Selected Types to: %v", selectedTypes))
	logging.LOG.Debug(fmt.Sprintf("Updating AutoDownload to: %v", autoDownload))

	// Update the selected types in the database
	logErr := UpdateSavedSetForItemInDB(ratingKey, selectedTypes, autoDownload)
	if logErr.Err != nil {
		utils.SendErrorJSONResponse(w, http.StatusInternalServerError, logErr)
		return
	}

	utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
		Status:  "success",
		Message: "Item updated successfully",
		Elapsed: utils.ElapsedTime(startTime),
		Data:    nil})
}

func UpdateSavedSetForItemInDB(ratingKey string, selectedTypes []string, autoDownload bool) logging.ErrorLog {
	// Convert SelectedTypes (slice of strings) to a comma-separated string
	selectedTypesStr := strings.Join(selectedTypes, ",")

	// Get the current time in the local timezone
	now := time.Now().In(time.Local)

	query := `
UPDATE auto_downloader
SET selected_types = ?, last_update = ?, auto_download = ?
WHERE id = ?`
	_, err := db.Exec(query, selectedTypesStr, now.UTC().Format(time.RFC3339), autoDownload, ratingKey)
	if err != nil {
		return logging.ErrorLog{Err: err, Log: logging.Log{
			Message: "Failed to update Saved Set in database",
		}}
	}
	logging.LOG.Debug(fmt.Sprintf("Saved Set updated successfully for item with ratingKey: %s", ratingKey))
	return logging.ErrorLog{}
}

func UpdateAutoDownloadItem(clientMessage modals.ClientMessage) logging.ErrorLog {
	mediaItemJSON, err := json.Marshal(clientMessage.MediaItem)
	if err != nil {
		return logging.ErrorLog{Err: err, Log: logging.Log{
			Message: "Failed to marshal MediaItem data",
		}}
	}

	setJSON, err := json.Marshal(clientMessage.Set)
	if err != nil {
		return logging.ErrorLog{Err: err, Log: logging.Log{
			Message: "Failed to marshal Set data",
		}}
	}

	// Convert SelectedTypes (slice of strings) to a comma-separated string
	selectedTypes := strings.Join(clientMessage.SelectedTypes, ",")

	// Get the current time in the local timezone
	now := time.Now().In(time.Local)

	query := `
UPDATE auto_downloader
SET media_item = ?, poster_set = ?, selected_types = ?, auto_download = ?, last_update = ?
WHERE id = ?`
	_, err = db.Exec(query, mediaItemJSON, setJSON, selectedTypes, clientMessage.AutoDownload, now.UTC().Format(time.RFC3339), clientMessage.MediaItem.RatingKey)
	if err != nil {
		return logging.ErrorLog{Err: err, Log: logging.Log{
			Message: "Failed to update data in database",
		}}
	}

	logging.LOG.Debug("Item updated successfully in the database")
	return logging.ErrorLog{}
}
