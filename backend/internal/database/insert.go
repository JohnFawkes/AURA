package database

import (
	"aura/internal/logging"
	"aura/internal/modals"
	"encoding/json"
	"strings"
	"time"
)

func SaveSavedSet(savedSet modals.Database_SavedSet) logging.ErrorLog {

	// Check if the media item is already in the database
	existsMediaItem, errLog := CheckIfMediaItemAlreadyInDatabase(savedSet.MediaItem.RatingKey)
	if errLog.Err != nil {
		return errLog
	}

	if existsMediaItem {
		logging.LOG.Trace("Media item already exists in the database, updating it")
		// Update the media item in the database
		errLog = UpdateMediaItemInDatabase(savedSet)
		if errLog.Err != nil {
			return errLog
		}
	} else {
		logging.LOG.Trace("Media item does not exist in the database, inserting it")
		// Insert the media item into the database
		errLog = InsertMediaItemIntoDatabase(savedSet)
		if errLog.Err != nil {
			return errLog
		}
	}

	// Check if the set is already in the database
	existsSet, errLog := CheckIfPosterSetAlreadyInDatabase(savedSet.Sets[0].ID)
	if errLog.Err != nil {
		return errLog
	}

	// If it exists, update it
	if existsSet {
		logging.LOG.Trace("Poster Set already exists in the database, updating it")
		errLog = UpdatePosterSetInDatabase(savedSet.Sets[0].Set, savedSet.MediaItem.RatingKey, savedSet.Sets[0].SelectedTypes, savedSet.Sets[0].AutoDownload)
		if errLog.Err != nil {
			return errLog
		}
	} else {
		logging.LOG.Trace("Poster Set does not exist in the database, inserting it")
		// Insert the set into the database
		errLog = InsertPosterSetIntoDatabase(savedSet)
		if errLog.Err != nil {
			return errLog
		}
	}

	return logging.ErrorLog{}
}

func InsertMediaItemIntoDatabase(savedSet modals.Database_SavedSet) logging.ErrorLog {
	// Marshal the MediaItem into JSON
	mediaItemJSON, err := json.Marshal(savedSet.MediaItem)
	if err != nil {
		return logging.ErrorLog{Err: err, Log: logging.Log{
			Message: "Failed to marshal MediaItem data",
		}}
	}

	query := `
	INSERT INTO Media_Item (id, media_item)
	VALUES (?, ?)
`
	_, err = db.Exec(query,
		savedSet.MediaItem.RatingKey,
		string(mediaItemJSON))
	if err != nil {
		return logging.ErrorLog{Err: err, Log: logging.Log{
			Message: "Failed to insert media item into database",
		}}
	}
	logging.LOG.Info("Media item inserted successfully into the database")
	return logging.ErrorLog{}
}

func InsertPosterSetIntoDatabase(savedSet modals.Database_SavedSet) logging.ErrorLog {
	// Marshal the Set into JSON
	setJSON, err := json.Marshal(savedSet.Sets[0].Set)
	if err != nil {
		return logging.ErrorLog{Err: err, Log: logging.Log{
			Message: "Failed to marshal Set data",
		}}
	}

	selectedTypes := strings.Join(savedSet.Sets[0].SelectedTypes, ",")

	// Get the current time in the local timezone
	now := time.Now().In(time.Local)

	setQuery := `
	INSERT INTO Poster_Sets (id, media_item_id, poster_set, selected_types, auto_download, last_update)
	VALUES (?, ?, ?, ?, ?, ?)
`
	_, err = db.Exec(setQuery,
		savedSet.Sets[0].ID,
		savedSet.MediaItem.RatingKey,
		string(setJSON),
		selectedTypes,
		savedSet.Sets[0].AutoDownload,
		now.UTC().Format(time.RFC3339))
	if err != nil {
		return logging.ErrorLog{Err: err, Log: logging.Log{
			Message: "Failed to insert data into database",
		}}
	}
	logging.LOG.Info("Poster set inserted successfully into the database")
	return logging.ErrorLog{}
}
