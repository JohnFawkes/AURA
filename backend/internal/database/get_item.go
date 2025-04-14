package database

import (
	"encoding/json"
	"poster-setter/internal/logging"
	"poster-setter/internal/modals"
	"strings"
)

func GetItemFromDatabase(ratingKey string) (modals.ClientMessage, logging.ErrorLog) {
	query := `
SELECT plex, poster_set, selected_types, auto_download FROM auto_downloader WHERE id = ?`
	var plexJSON, setJSON, selectedTypes string
	var autoDownload bool
	err := db.QueryRow(query, ratingKey).Scan(&plexJSON, &setJSON, &selectedTypes, &autoDownload)
	if err != nil {
		return modals.ClientMessage{}, logging.ErrorLog{Err: err, Log: logging.Log{
			Message: "Failed to query database",
		}}
	}

	// Unmarshal the JSON data
	var plex modals.MediaItem
	err = json.Unmarshal([]byte(plexJSON), &plex)
	if err != nil {
		return modals.ClientMessage{}, logging.ErrorLog{Err: err, Log: logging.Log{
			Message: "Failed to unmarshal Plex data",
		}}
	}

	var set modals.PosterSet
	err = json.Unmarshal([]byte(setJSON), &set)
	if err != nil {
		return modals.ClientMessage{}, logging.ErrorLog{Err: err, Log: logging.Log{
			Message: "Failed to unmarshal Set data",
		}}
	}

	// Convert the comma-separated string back to a slice of strings
	selectedTypesSlice := strings.Split(selectedTypes, ",")

	clientMessage := modals.ClientMessage{
		Plex:          plex,
		Set:           set,
		SelectedTypes: selectedTypesSlice,
		AutoDownload:  autoDownload,
	}

	return clientMessage, logging.ErrorLog{}
}
