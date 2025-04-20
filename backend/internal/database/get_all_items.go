package database

import (
	"encoding/json"
	"poster-setter/internal/logging"
	"poster-setter/internal/modals"
	"strings"
)

func GetAllItemsFromDatabase() ([]modals.ClientMessage, logging.ErrorLog) {
	query := `
SELECT plex, poster_set, selected_types, auto_download, last_update FROM auto_downloader`
	rows, err := db.Query(query)
	if err != nil {
		return nil, logging.ErrorLog{Err: err, Log: logging.Log{
			Message: "Failed to query database",
		}}
	}
	defer rows.Close()
	var items []modals.ClientMessage
	for rows.Next() {
		var plexJSON, setJSON, selectedTypes string
		var autoDownload bool
		var lastUpdate string
		err := rows.Scan(&plexJSON, &setJSON, &selectedTypes, &autoDownload, &lastUpdate)
		if err != nil {
			return nil, logging.ErrorLog{Err: err, Log: logging.Log{
				Message: "Failed to scan row",
			}}
		}

		// Unmarshal the JSON data
		var plex modals.MediaItem
		err = json.Unmarshal([]byte(plexJSON), &plex)
		if err != nil {
			return nil, logging.ErrorLog{Err: err, Log: logging.Log{
				Message: "Failed to unmarshal Plex data",
			}}
		}

		var set modals.PosterSet
		err = json.Unmarshal([]byte(setJSON), &set)
		if err != nil {
			return nil, logging.ErrorLog{Err: err, Log: logging.Log{
				Message: "Failed to unmarshal Set data",
			}}
		}

		selectedTypesSlice := strings.Split(selectedTypes, ",")

		clientMessage := modals.ClientMessage{
			MediaItem:     plex,
			Set:           set,
			LastUpdate:    lastUpdate,
			SelectedTypes: selectedTypesSlice,
			AutoDownload:  autoDownload,
		}
		items = append(items, clientMessage)
	}
	if err = rows.Err(); err != nil {
		return nil, logging.ErrorLog{Err: err, Log: logging.Log{
			Message: "Error iterating rows",
		}}
	}
	return items, logging.ErrorLog{}
}
