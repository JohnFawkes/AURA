package database

import (
	"encoding/json"
	"poster-setter/internal/logging"
	"poster-setter/internal/modals"
	"strings"
	"time"
)

func SaveClientMessage(clientMessage modals.ClientMessage) logging.ErrorLog {

	// Check if the item is already in the database
	exists, errLog := CheckIfAlreadyInDatabase(clientMessage.Plex.RatingKey)
	if errLog.Err != nil {
		return errLog
	}

	// If it exists, update it
	if exists {
		logging.LOG.Trace("Item already exists in the database, updating it")
		return UpdateAutoDownloadItem(clientMessage)
	}

	logging.LOG.Trace("Item does not exist in the database, inserting it")
	// If it doesn't exist, insert it
	plexJSON, err := json.Marshal(clientMessage.Plex)
	if err != nil {
		return logging.ErrorLog{Err: err, Log: logging.Log{
			Message: "Failed to marshal Plex data",
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
INSERT INTO auto_downloader (id, plex, poster_set, selected_types, auto_download, last_update)
VALUES (?, ?, ?, ?, ?, ?)
`
	_, err = db.Exec(query, clientMessage.Plex.RatingKey, string(plexJSON), string(setJSON), selectedTypes, clientMessage.AutoDownload, now.UTC().Format(time.RFC3339))
	if err != nil {
		return logging.ErrorLog{Err: err, Log: logging.Log{
			Message: "Failed to insert data into database",
		}}
	}

	return logging.ErrorLog{}
}
