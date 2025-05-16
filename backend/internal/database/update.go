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
)

func UpdateSavedSetTypesForItem(w http.ResponseWriter, r *http.Request) {
	logging.LOG.Debug(r.URL.Path)
	startTime := time.Now()

	// Get the request body
	// Define a struct to match the expected JSON object
	var savedSet modals.Database_SavedSet
	// Decode the request body into the struct
	err := json.NewDecoder(r.Body).Decode(&savedSet)
	if err != nil {
		utils.SendErrorJSONResponse(w, http.StatusBadRequest, logging.ErrorLog{
			Err: err,
			Log: logging.Log{
				Message: "Failed to decode request body",
			},
		})
		return
	}

	for _, set := range savedSet.Sets {
		// Check if the set is marked for deletion
		if set.ToDelete {
			logErr := DeletePosterSetFromDatabaseByID(set.ID)
			if logErr.Err != nil {
				logging.LOG.Error(fmt.Sprintf("Failed to delete PosterSet '%s': %v", set.ID, logErr.Err))
			}

		} else { // Update the MediaItem in the database
			logging.LOG.Debug(fmt.Sprintf("Updating Poster Set '%s'", set.ID))
			logErr := UpdatePosterSetInDatabase(set.Set, savedSet.MediaItem.RatingKey, set.SelectedTypes, set.AutoDownload)
			if logErr.Err != nil {
				utils.SendErrorJSONResponse(w, http.StatusInternalServerError, logErr)
			}
		}
	}

	utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
		Status:  "success",
		Message: "Item updated successfully",
		Elapsed: utils.ElapsedTime(startTime),
		Data:    nil})
}

func UpdateMediaItemInDatabase(savedSet modals.Database_SavedSet) logging.ErrorLog {
	// Marshal the MediaItem into JSON
	mediaItemJSONBytes, err := json.Marshal(savedSet.MediaItem)
	if err != nil {
		return logging.ErrorLog{Err: err, Log: logging.Log{
			Message: "Failed to marshal MediaItem data",
		}}
	}
	savedSet.MediaItemJSON = string(mediaItemJSONBytes)

	query := `
UPDATE Media_Item
SET media_item = ?
WHERE id = ?`
	_, err = db.Exec(query, savedSet.MediaItemJSON, savedSet.MediaItem.RatingKey)
	if err != nil {
		return logging.ErrorLog{Err: err, Log: logging.Log{
			Message: "Failed to update MediaItem data in database",
		}}
	}
	logging.LOG.Info(fmt.Sprintf("DB - MediaItem updated successfully for item: %s", savedSet.MediaItem.Title))
	return logging.ErrorLog{}
}

func UpdatePosterSetInDatabase(posterSet modals.PosterSet, ratingKey string, selectedTypes []string, autoDownload bool) logging.ErrorLog {
	// Marshal the Set into JSON
	setJSONBytes, err := json.Marshal(posterSet)
	if err != nil {
		return logging.ErrorLog{Err: err, Log: logging.Log{
			Message: "Failed to marshal Set data",
		}}
	}

	// Convert SelectedTypes (slice of strings) to a comma-separated string
	selectedTypesStr := strings.Join(selectedTypes, ",")

	// Get the current time in the local timezone
	now := time.Now().In(time.Local)

	query := `
UPDATE Poster_Sets
SET media_item_id = ?, poster_set = ?, selected_types = ?, auto_download = ?, last_update = ?
WHERE id = ?`
	_, err = db.Exec(query,
		ratingKey,
		string(setJSONBytes),
		selectedTypesStr,
		autoDownload,
		now.UTC().Format(time.RFC3339),
		posterSet.ID)
	if err != nil {
		return logging.ErrorLog{Err: err, Log: logging.Log{
			Message: "Failed to update Set data in database",
		}}
	}
	logging.LOG.Debug(fmt.Sprintf("DB - PosterSet '%s' updated successfully for item '%s'", posterSet.ID, ratingKey))
	return logging.ErrorLog{}
}
