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
		Err := logging.NewStandardError()
		Err.Message = "Failed to decode request body"
		Err.HelpText = "Ensure the request body is a valid JSON object matching the expected structure."
		Err.Details = fmt.Sprintf("Request Body: %s", r.Body)
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}

	for _, posterSet := range SaveItem.PosterSets {
		// Check if posterSet.toDelete is true
		if posterSet.ToDelete {
			Err := deletePosterItemFromDatabase(SaveItem.MediaItemID, posterSet.PosterSet.ID)
			if Err.Message != "" {
				utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
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

			Err := UpdateItemInDatabase(saveItem)
			if Err.Message != "" {
				utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
				return
			}
		}
	}

	utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
		Status:  "success",
		Elapsed: utils.ElapsedTime(startTime),
		Data:    "success"})
}

func UpdateItemInDatabase(saveItem modals.DBSavedItem) logging.StandardError {
	Err := logging.NewStandardError()

	// Set the MediaItem to exist in the database, in case it was not set
	saveItem.MediaItem.ExistInDatabase = true

	// Marshal the MediaItem into JSON
	mediaItemJSONBytes, err := json.Marshal(saveItem.MediaItem)
	if err != nil {
		Err.Message = "Failed to marshal MediaItem data"
		Err.HelpText = "Ensure the MediaItem struct is correctly defined and contains valid data."
		Err.Details = fmt.Sprintf("MediaItem: %+v", saveItem.MediaItem)
		return Err
	}

	// Marshal the PosterSet into JSON
	posterSetJSONBytes, err := json.Marshal(saveItem.PosterSet)
	if err != nil {
		Err.Message = "Failed to marshal PosterSet data"
		Err.HelpText = "Ensure the PosterSet struct is correctly defined and contains valid data."
		Err.Details = fmt.Sprintf("PosterSet: %+v", saveItem.PosterSet)
		return Err
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
		Err.Message = "Failed to update MediaItem data in database"
		Err.HelpText = "Ensure the database connection is established and the query is correct."
		Err.Details = fmt.Sprintf("MediaItemID: %s, MediaItem: %+v", saveItem.MediaItem.RatingKey, saveItem.MediaItem)
		return Err
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
		Err.Message = "Failed to update PosterSet data in database"
		Err.HelpText = "Ensure the database connection is established and the query is correct."
		Err.Details = fmt.Sprintf("PosterSetID: %s, PosterSet: %+v, MediaItemID: %s", saveItem.PosterSet.ID, saveItem.PosterSet, saveItem.MediaItem.RatingKey)
		return Err
	}
	logging.LOG.Info(fmt.Sprintf("DB - Item updated successfully for: %s - Set %s", saveItem.MediaItem.Title, saveItem.PosterSet.ID))
	return logging.StandardError{}
}
