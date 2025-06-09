package database

import (
	"aura/internal/logging"
	"aura/internal/modals"
	"encoding/json"
	"strings"
	"time"
)

func SaveItemInDB(saveItem modals.DBSavedItem) logging.ErrorLog {

	// Check if the item already exists in the database
	// Check Media Item Rating Key and Poster Set ID
	// If it exists, update the item
	itemExists, errLog := CheckIfItemExistsInDatabase(saveItem.MediaItemID, saveItem.PosterSetID)
	if errLog.Err != nil {
		return errLog
	}

	if itemExists {
		logging.LOG.Trace("Item already exists in database, updating it")
		// Update the media item in the database
		errLog = UpdateItemInDatabase(saveItem)
		if errLog.Err != nil {
			return errLog
		}
	} else {
		logging.LOG.Trace("Item does not exist in database, inserting it")
		// Insert the item into the database
		errLog = InsertItemIntoDatabase(saveItem)
		if errLog.Err != nil {
			return errLog
		}
	}

	return logging.ErrorLog{}
}

func InsertItemIntoDatabase(saveItem modals.DBSavedItem) logging.ErrorLog {
	// Marshal the MediaItem into JSON
	mediaItemJSON, err := json.Marshal(saveItem.MediaItem)
	if err != nil {
		return logging.ErrorLog{Err: err, Log: logging.Log{
			Message: "Failed to marshal MediaItem data",
		}}
	}

	// Marshal the PosterSet into JSON
	posterSetJSON, err := json.Marshal(saveItem.PosterSet)
	if err != nil {
		return logging.ErrorLog{Err: err, Log: logging.Log{
			Message: "Failed to marshal PosterSet data",
		}}
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
		return logging.ErrorLog{Err: err, Log: logging.Log{
			Message: "Failed to insert Item into database",
		}}
	}
	logging.LOG.Info("Item inserted successfully into the database")
	return logging.ErrorLog{}
}
