package database

import (
	"aura/internal/logging"
	"aura/internal/modals"
	"encoding/json"
	"strings"
	"time"
)

func SaveItemInDB(saveItem modals.DBSavedItem) logging.StandardError {
	Err := logging.NewStandardError()
	// Check if the item already exists in the database
	// Check Media Item Rating Key and Poster Set ID
	// If it exists, update the item
	itemExists, Err := CheckIfItemExistsInDatabase(saveItem.MediaItemID, saveItem.PosterSetID)
	if Err.Message != "" {
		return Err
	}

	if itemExists {
		logging.LOG.Trace("Item already exists in database, updating it")
		// Update the media item in the database
		Err = UpdateItemInDatabase(saveItem)
		if Err.Message != "" {
			return Err
		}
	} else {
		logging.LOG.Trace("Item does not exist in database, inserting it")
		// Insert the item into the database
		Err = InsertItemIntoDatabase(saveItem)
		if Err.Message != "" {
			return Err
		}
	}

	return logging.StandardError{}
}

func InsertItemIntoDatabase(saveItem modals.DBSavedItem) logging.StandardError {
	Err := logging.NewStandardError()

	// Mark the MediaItem as existing in the database, since we are inserting it now
	saveItem.MediaItem.ExistInDatabase = true

	// Marshal the MediaItem into JSON
	mediaItemJSON, err := json.Marshal(saveItem.MediaItem)
	if err != nil {
		Err.Message = "Failed to marshal MediaItem data"
		Err.HelpText = "Ensure the MediaItem struct is correctly defined and contains valid data."
		Err.Details = map[string]any{
			"error": err.Error(),
			"json":  string(mediaItemJSON),
		}
		return Err
	}

	// Marshal the PosterSet into JSON
	posterSetJSON, err := json.Marshal(saveItem.PosterSet)
	if err != nil {
		Err.Message = "Failed to marshal PosterSet data"
		Err.HelpText = "Ensure the PosterSet struct is correctly defined and contains valid data."
		Err.Details = map[string]any{
			"error": err.Error(),
			"json":  string(posterSetJSON),
		}
		return Err
	}

	// Convert SelectedTypes (slice of strings) to a comma-separated string
	selectedTypesStr := strings.Join(saveItem.SelectedTypes, ",")

	query := `
		INSERT INTO SavedItems (media_item_id, media_item, poster_set_id, poster_set, selected_types, auto_download, last_update)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	_, err = db.Exec(query,
		saveItem.MediaItem.RatingKey,
		string(mediaItemJSON),
		saveItem.PosterSet.ID,
		string(posterSetJSON),
		selectedTypesStr,
		saveItem.AutoDownload,
		saveItem.PosterSet.DateUpdated.UTC().Format(time.RFC3339),
	)
	if err != nil {
		Err.Message = "Failed to insert item into database"
		Err.HelpText = "Ensure the database connection is established and the query is correct."
		Err.Details = map[string]any{
			"error":          err.Error(),
			"query":          query,
			"media_item_id":  saveItem.MediaItem.RatingKey,
			"media_item":     string(mediaItemJSON),
			"poster_set_id":  saveItem.PosterSet.ID,
			"poster_set":     string(posterSetJSON),
			"selected_types": selectedTypesStr,
			"auto_download":  saveItem.AutoDownload,
			"last_update":    saveItem.PosterSet.DateUpdated.UTC().Format(time.RFC3339),
		}
		return Err
	}
	logging.LOG.Info("Item inserted successfully into the database")
	return logging.StandardError{}
}
