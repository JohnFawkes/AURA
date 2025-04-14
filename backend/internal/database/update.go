package database

import (
	"encoding/json"
	"poster-setter/internal/logging"
	"poster-setter/internal/modals"
	"strings"
	"time"
)

func UpdateAutoDownloadItem(clientMessage modals.ClientMessage) logging.ErrorLog {
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
UPDATE auto_downloader
SET plex = ?, poster_set = ?, selected_types = ?, auto_download = ?, last_update = ?
WHERE id = ?`
	_, err = db.Exec(query, plexJSON, setJSON, selectedTypes, clientMessage.AutoDownload, now.UTC().Format(time.RFC3339), clientMessage.Plex.RatingKey)
	if err != nil {
		return logging.ErrorLog{Err: err, Log: logging.Log{
			Message: "Failed to update data in database",
		}}
	}

	logging.LOG.Debug("Item updated successfully in the database")
	return logging.ErrorLog{}
}
