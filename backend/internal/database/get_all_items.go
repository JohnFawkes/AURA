package database

import (
	"aura/internal/logging"
	"aura/internal/modals"
	"aura/internal/utils"
	"net/http"
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

func GetAllItemsFromDatabase() ([]modals.DBMediaItemWithPosterSets, logging.ErrorLog) {
	// Query all rows from SavedItems.
	query := `
    SELECT media_item_id, media_item, poster_set_id, poster_set, selected_types, auto_download, last_update
    FROM SavedItems`
	rows, err := db.Query(query)
	if err != nil {
		return nil, logging.ErrorLog{Err: err, Log: logging.Log{Message: "Failed to query all item from database"}}
	}
	defer rows.Close()

	// Map to group poster sets by media_item_id.
	mediaMap := map[string]*modals.DBMediaItemWithPosterSets{}

	for rows.Next() {
		var savedItem modals.DBSavedItem
		var selectedTypesStr string
		err := rows.Scan(
			&savedItem.MediaItemID,
			&savedItem.MediaItemJSON,
			&savedItem.PosterSetID,
			&savedItem.PosterSetJSON,
			&selectedTypesStr,
			&savedItem.AutoDownload,
			&savedItem.LastDownloaded,
		)
		if err != nil {
			return nil, logging.ErrorLog{Err: err, Log: logging.Log{Message: "Failed to scan row"}}
		}

		// Unmarshal MediaItem and PosterSet from JSON if necessary.
		var mediaItem modals.MediaItem
		var posterSet modals.PosterSet
		if err := UnmarshalMediaItem(savedItem.MediaItemJSON, &mediaItem); err.Err != nil {
			return nil, err
		}
		if err := UnmarshalPosterSet(savedItem.PosterSetJSON, &posterSet); err.Err != nil {
			return nil, err
		}

		// Convert selectedTypesStr to a slice of strings.
		selectedTypes := strings.Split(selectedTypesStr, ",")
		savedItem.SelectedTypes = selectedTypes

		psDetail := modals.DBPosterSetDetail{
			PosterSetID:    savedItem.PosterSetID,
			PosterSet:      posterSet,
			PosterSetJSON:  savedItem.PosterSetJSON,
			LastDownloaded: savedItem.LastDownloaded,
			SelectedTypes:  savedItem.SelectedTypes,
			AutoDownload:   savedItem.AutoDownload,
		}

		// Group by MediaItemID.
		if existing, ok := mediaMap[savedItem.MediaItemID]; ok {
			existing.PosterSets = append(existing.PosterSets, psDetail)
		} else {
			mediaMap[savedItem.MediaItemID] = &modals.DBMediaItemWithPosterSets{
				MediaItemID:   savedItem.MediaItemID,
				MediaItem:     mediaItem,
				MediaItemJSON: savedItem.MediaItemJSON,
				PosterSets:    []modals.DBPosterSetDetail{psDetail},
			}
		}
	}

	// Convert the map to a slice.
	result := []modals.DBMediaItemWithPosterSets{}
	for _, v := range mediaMap {
		result = append(result, *v)
	}

	return result, logging.ErrorLog{}
}
