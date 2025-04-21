package database

import (
	"encoding/json"
	"net/http"
	"poster-setter/internal/logging"
	"poster-setter/internal/modals"
	"poster-setter/internal/utils"
	"strings"
	"time"
)

func GetAllItems(w http.ResponseWriter, r *http.Request) {
	logging.LOG.Debug(r.URL.Path)
	startTime := time.Now()

	items, logErr := GetAllItemsFromDatabase()
	if logErr.Err != nil {
		utils.SendErrorJSONResponse(w, http.StatusInternalServerError, logErr)
		return
	}

	utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
		Status:  "success",
		Message: "Fetched all items",
		Elapsed: utils.ElapsedTime(startTime),
		Data:    items,
	})
}

func GetAllItemsFromDatabase() ([]modals.ClientMessage, logging.ErrorLog) {
	query := `
SELECT media_item, poster_set, selected_types, auto_download, last_update FROM auto_downloader`
	rows, err := db.Query(query)
	if err != nil {
		return nil, logging.ErrorLog{Err: err, Log: logging.Log{
			Message: "Failed to query database",
		}}
	}
	defer rows.Close()
	var items []modals.ClientMessage
	for rows.Next() {
		var mediaItemJSON, setJSON, selectedTypes string
		var autoDownload bool
		var lastUpdate string
		err := rows.Scan(&mediaItemJSON, &setJSON, &selectedTypes, &autoDownload, &lastUpdate)
		if err != nil {
			return nil, logging.ErrorLog{Err: err, Log: logging.Log{
				Message: "Failed to scan row",
			}}
		}

		// Unmarshal the JSON data
		var mediaItem modals.MediaItem
		err = json.Unmarshal([]byte(mediaItemJSON), &mediaItem)
		if err != nil {
			return nil, logging.ErrorLog{Err: err, Log: logging.Log{
				Message: "Failed to unmarshal MediaItem data",
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
			MediaItem:     mediaItem,
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
