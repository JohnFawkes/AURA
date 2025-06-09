package database

import (
	"aura/internal/logging"
	"aura/internal/modals"
	"aura/internal/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

func UpdateSavedSetTypesForItem(w http.ResponseWriter, r *http.Request) {
	logging.LOG.Debug(r.URL.Path)
	startTime := time.Now()

	//Get the request body
	var SaveItem modals.DBMediaItemWithPosterSets

	// Decode the request body into the struct
	err := json.NewDecoder(r.Body).Decode(&SaveItem)
	if err != nil {
		utils.SendErrorJSONResponse(w, http.StatusBadRequest, logging.ErrorLog{
			Err: err,
			Log: logging.Log{
				Message: "Failed to decode request body",
			},
		})
		return
	}

	for _, posterSet := range SaveItem.PosterSets {
		// Check if posterSet.toDelete is true
		if posterSet.ToDelete {
			logErr := deletePosterItemFromDatabase(SaveItem.MediaItemID, posterSet.PosterSet.ID)
			if logErr.Err != nil {
				utils.SendErrorJSONResponse(w, http.StatusInternalServerError, logErr)
				return
			}
		} else {
			// Update the poster set in the database
			saveItem := modals.DBSavedItem{
				MediaItemID:   SaveItem.MediaItemID,
				MediaItem:     SaveItem.MediaItem,
				PosterSetID:   posterSet.PosterSet.ID,
				PosterSet:     posterSet.PosterSet,
				SelectedTypes: posterSet.SelectedTypes,
				AutoDownload:  posterSet.AutoDownload,
			}

			logErr := UpdateItemInDatabase(saveItem)
			if logErr.Err != nil {
				utils.SendErrorJSONResponse(w, http.StatusInternalServerError, logErr)
				return
			}
		}
	}

	utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
		Status:  "success",
		Message: "Item updated successfully",
		Elapsed: utils.ElapsedTime(startTime),
		Data:    "success"})
}

func UpdateItemInDatabase(saveItem modals.DBSavedItem) logging.ErrorLog {
	// Marshal the MediaItem into JSON
	mediaItemJSONBytes, err := json.Marshal(saveItem.MediaItem)
	if err != nil {
		return logging.ErrorLog{Err: err, Log: logging.Log{
			Message: "Failed to marshal MediaItem data",
		}}
	}

	// Marshal the PosterSet into JSON
	posterSetJSONBytes, err := json.Marshal(saveItem.PosterSet)
	if err != nil {
		return logging.ErrorLog{Err: err, Log: logging.Log{
			Message: "Failed to marshal PosterSet data",
		}}
	}

	// Convert SelectedTypes (slice of strings) to a comma-separated string
	selectedTypesStr := strings.Join(saveItem.SelectedTypes, ",")

	// Update the MediaItem in the database for any media item changes
	query := `
	UPDATE SavedItems
	SET media_item = ?
	WHERE media_item_id = ?`
	_, err = db.Exec(query,
		string(mediaItemJSONBytes),
		saveItem.MediaItem.RatingKey)
	if err != nil {
		return logging.ErrorLog{Err: err, Log: logging.Log{
			Message: "Failed to update MediaItem data in database",
		}}
	}

	// Update the PosterSet in the database
	query = `
	UPDATE SavedItems
	SET media_item = ?, poster_set_id = ?, poster_set = ?, selected_types = ?, auto_download = ?, last_update = ?
	WHERE media_item_id = ? AND poster_set_id = ?`
	_, err = db.Exec(query,
		string(mediaItemJSONBytes),
		saveItem.PosterSet.ID,
		string(posterSetJSONBytes),
		selectedTypesStr,
		saveItem.AutoDownload,
		saveItem.PosterSet.DateUpdated.UTC().Format(time.RFC3339),
		saveItem.MediaItem.RatingKey,
		saveItem.PosterSet.ID)
	if err != nil {
		return logging.ErrorLog{Err: err, Log: logging.Log{
			Message: "Failed to update Item data in database",
		}}
	}
	logging.LOG.Info(fmt.Sprintf("DB - Item updated successfully for: %s - Set %s", saveItem.MediaItem.Title, saveItem.PosterSet.ID))
	return logging.ErrorLog{}
}
