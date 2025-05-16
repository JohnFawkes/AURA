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

func GetAllItemsFromDatabase() ([]modals.Database_SavedSet, logging.ErrorLog) {

	mediaQuery := "SELECT id, media_item FROM Media_Item"
	mediaRows, err := db.Query(mediaQuery)
	if err != nil {
		return nil, logging.ErrorLog{Err: err, Log: logging.Log{
			Message: "Failed to query Media_Item table",
		}}
	}
	defer mediaRows.Close()

	var savedSets []modals.Database_SavedSet
	for mediaRows.Next() {
		var mediaItemID string
		var mediaItemJSON string
		err := mediaRows.Scan(&mediaItemID, &mediaItemJSON)
		if err != nil {
			return nil, logging.ErrorLog{Err: err, Log: logging.Log{
				Message: "Failed to scan Media_Item row",
			}}
		}

		savedSet := modals.Database_SavedSet{
			ID:            mediaItemID,
			MediaItemJSON: mediaItemJSON,
		}

		// Unmarshal the JSON data into modals.MediaItem struct
		var mediaItem modals.MediaItem
		err = json.Unmarshal([]byte(mediaItemJSON), &mediaItem)
		if err != nil {
			return nil, logging.ErrorLog{Err: err, Log: logging.Log{
				Message: "Failed to unmarshal MediaItem data",
			}}
		}
		savedSet.MediaItem = mediaItem

		posterQuery := `
        SELECT id, media_item_id, poster_set, selected_types, auto_download, last_update
        FROM Poster_Sets
        WHERE media_item_id = ?`
		posterRows, err := db.Query(posterQuery, mediaItemID)
		if err != nil {
			return nil, logging.ErrorLog{Err: err, Log: logging.Log{
				Message: "Failed to query Poster_Sets table",
			}}
		}

		var sets []modals.Database_Set

		if posterRows.Next() {
			var setID, mediaItemID, setJSON, selectedTypesStr string
			var autoDownload bool
			var lastUpdate string
			err := posterRows.Scan(&setID, &mediaItemID, &setJSON, &selectedTypesStr, &autoDownload, &lastUpdate)
			if err != nil {
				return nil, logging.ErrorLog{Err: err, Log: logging.Log{
					Message: "Failed to scan Poster_Sets row",
				}}
			}

			// Convert the comma-separated string into []string.
			var selectedTypes []string
			if selectedTypesStr != "" {
				selectedTypes = strings.Split(selectedTypesStr, ",")
			}

			// Unmarshal the JSON data into modals.PosterSet struct
			var set modals.PosterSet
			err = json.Unmarshal([]byte(setJSON), &set)
			if err != nil {
				return nil, logging.ErrorLog{Err: err, Log: logging.Log{
					Message: "Failed to unmarshal PosterSet data",
				}}
			}

			dbSet := modals.Database_Set{
				ID:            setID,
				MediaItemID:   mediaItemID,
				Set:           set,
				SetJSON:       setJSON,
				SelectedTypes: selectedTypes,
				AutoDownload:  autoDownload,
				LastUpdate:    lastUpdate,
			}
			sets = append(sets, dbSet)
		}

		posterRows.Close()

		// If no poster sets were found, print a message
		if len(sets) == 0 {
			continue
		} else {
			savedSet.Sets = sets
			savedSets = append(savedSets, savedSet)
		}
	}
	if err = mediaRows.Err(); err != nil {
		return nil, logging.ErrorLog{Err: err, Log: logging.Log{
			Message: "Error iterating Media_Item rows",
		}}
	}
	if err = mediaRows.Close(); err != nil {
		return nil, logging.ErrorLog{Err: err, Log: logging.Log{
			Message: "Error closing Media_Item rows",
		}}
	}
	return savedSets, logging.ErrorLog{}

}
