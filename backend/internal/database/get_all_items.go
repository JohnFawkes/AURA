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

	items, Err := GetAllItemsFromDatabase()
	if Err.Message != "" {
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}

	utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
		Status:  "success",
		Elapsed: utils.ElapsedTime(startTime),
		Data:    items,
	})
}

func GetAllItemsFromDatabase() ([]modals.DBMediaItemWithPosterSets, logging.StandardError) {
	Err := logging.NewStandardError()

	// Query all rows from SavedItems.
	query := `
SELECT media_item_id, media_item, poster_set_id, poster_set, selected_types, auto_download, last_update
FROM SavedItems
ORDER BY media_item_id`
	rows, err := db.Query(query)
	if err != nil {
		Err.Message = "Failed to query all items from database"
		Err.HelpText = "Ensure the database connection is established and the query is correct."
		Err.Details = "Query: " + query
		return nil, Err
	}
	defer rows.Close()

	var (
		result    []modals.DBMediaItemWithPosterSets
		currentID string
		current   *modals.DBMediaItemWithPosterSets
	)

	for rows.Next() {
		var savedItem modals.DBSavedItem
		var selectedTypesStr string
		if err := rows.Scan(
			&savedItem.MediaItemID,
			&savedItem.MediaItemJSON,
			&savedItem.PosterSetID,
			&savedItem.PosterSetJSON,
			&selectedTypesStr,
			&savedItem.AutoDownload,
			&savedItem.LastDownloaded,
		); err != nil {
			Err.Message = "Failed to scan row from SavedItems"
			Err.HelpText = "Check schema/data types."
			Err.Details = "Query: " + query
			return nil, Err
		}

		// Start a new group when media_item_id changes
		if savedItem.MediaItemID != currentID {
			if current != nil {
				result = append(result, *current)
			}
			currentID = savedItem.MediaItemID

			var mediaItem modals.MediaItem
			if Err = UnmarshalMediaItem(savedItem.MediaItemJSON, &mediaItem); Err.Message != "" {
				return nil, Err
			}

			current = &modals.DBMediaItemWithPosterSets{
				MediaItemID:   savedItem.MediaItemID,
				MediaItem:     mediaItem,
				MediaItemJSON: savedItem.MediaItemJSON,
				PosterSets:    make([]modals.DBPosterSetDetail, 0, 4),
			}
		}

		var posterSet modals.PosterSet
		if Err = UnmarshalPosterSet(savedItem.PosterSetJSON, &posterSet); Err.Message != "" {
			return nil, Err
		}

		if selectedTypesStr != "" {
			savedItem.SelectedTypes = strings.Split(selectedTypesStr, ",")
		} else {
			savedItem.SelectedTypes = nil
		}

		psDetail := modals.DBPosterSetDetail{
			PosterSetID:    savedItem.PosterSetID,
			PosterSet:      posterSet,
			PosterSetJSON:  savedItem.PosterSetJSON,
			LastDownloaded: savedItem.LastDownloaded,
			SelectedTypes:  savedItem.SelectedTypes,
			AutoDownload:   savedItem.AutoDownload,
		}
		current.PosterSets = append(current.PosterSets, psDetail)
	}

	if current != nil {
		result = append(result, *current)
	}

	return result, logging.StandardError{}
}
