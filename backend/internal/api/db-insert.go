package api

import (
	"aura/internal/logging"
	"context"
	"encoding/json"
	"slices"
	"strings"
	"time"
)

func DB_InsertAllInfoIntoTables(ctx context.Context, item DBMediaItemWithPosterSets) logging.LogErrorInfo {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Adding Item into Database", logging.LevelInfo)
	defer logAction.Complete()

	// Return Success for testing
	// return logging.LogErrorInfo{}

	// Start a DB transaction
	actionBeginTx := logAction.AddSubAction("Start DB Transaction", logging.LevelTrace)
	tx, err := db.Begin()
	if err != nil {
		actionBeginTx.SetError("Failed to begin DB transaction", err.Error(), map[string]any{
			"item": item,
		})
		return *actionBeginTx.Error
	}
	actionBeginTx.Complete()

	// Remove any "ignore" poster sets for this TMDB_ID and LibraryTitle from the database
	actionRemoveIgnorePosterSets := logAction.AddSubAction("Remove 'Ignore' From Poster Sets Table", logging.LevelTrace)
	_, err = tx.Exec(`
        DELETE FROM PosterSets
        WHERE PosterSetID = ? AND TMDB_ID = ? AND LibraryTitle = ?
    `, "ignore", item.MediaItem.TMDB_ID, item.MediaItem.LibraryTitle)
	if err != nil {
		tx.Rollback()
		actionRemoveIgnorePosterSets.SetError("Failed to remove 'ignore' poster sets", err.Error(), map[string]any{
			"item": item,
		})
		return *actionRemoveIgnorePosterSets.Error
	}
	actionRemoveIgnorePosterSets.Complete()

	// Remove any ignore entries from the SavedItems table
	actionRemoveIgnoreEntries := logAction.AddSubAction("Remove 'Ignore' Entries From Saved Items Table", logging.LevelTrace)
	_, err = tx.Exec(`
		DELETE FROM SavedItems
		WHERE PosterSetID = ? AND TMDB_ID = ? AND LibraryTitle = ?
	`, "ignore", item.MediaItem.TMDB_ID, item.MediaItem.LibraryTitle)
	if err != nil {
		tx.Rollback()
		actionRemoveIgnoreEntries.SetError("Failed to remove 'ignore' entries from SavedItems", err.Error(), map[string]any{
			"item": item,
		})
		return *actionRemoveIgnoreEntries.Error
	}
	actionRemoveIgnoreEntries.Complete()

	// Check if the media item already exists
	isUpdate := false
	var exists int
	actionCheckExistence := logAction.AddSubAction("Check If Media Item Exists in DB", logging.LevelTrace)
	err = tx.QueryRow(`
    SELECT COUNT(*) FROM MediaItems WHERE TMDB_ID = ? AND LibraryTitle = ?
`, item.MediaItem.TMDB_ID, item.MediaItem.LibraryTitle).Scan(&exists)
	if err != nil {
		tx.Rollback()
		actionCheckExistence.SetError("Failed to check if media item exists", err.Error(), map[string]any{
			"item": item,
		})
		return *actionCheckExistence.Error
	}
	if exists > 0 {
		isUpdate = true
	}
	actionCheckExistence.Complete()

	// Insert into MediaItems table
	mediaItemJSON, _ := json.Marshal(item.MediaItem)
	guidsJSON, _ := json.Marshal(item.MediaItem.Guids)
	movieJSON, _ := json.Marshal(item.MediaItem.Movie)
	seriesJSON, _ := json.Marshal(item.MediaItem.Series)
	actionAddToMediaItemsTable := logAction.AddSubAction("Insert/Update Media Item in MediaItems Table", logging.LevelTrace)
	_, err = tx.Exec(`
    INSERT INTO MediaItems (
        TMDB_ID, LibraryTitle, RatingKey, Type, Title, Year, Thumb, ContentRating, Summary, UpdatedAt, AddedAt, ReleasedAt, Guids_JSON, Movie_JSON, Series_JSON, Full_JSON
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    ON CONFLICT(TMDB_ID, LibraryTitle) DO UPDATE SET
        RatingKey = excluded.RatingKey,
        Type = excluded.Type,
        Title = excluded.Title,
        Year = excluded.Year,
        Thumb = excluded.Thumb,
        ContentRating = excluded.ContentRating,
        Summary = excluded.Summary,
        UpdatedAt = excluded.UpdatedAt,
        AddedAt = excluded.AddedAt,
        ReleasedAt = excluded.ReleasedAt,
        Guids_JSON = excluded.Guids_JSON,
        Movie_JSON = excluded.Movie_JSON,
        Series_JSON = excluded.Series_JSON,
        Full_JSON = excluded.Full_JSON
`, item.MediaItem.TMDB_ID, item.MediaItem.LibraryTitle, item.MediaItem.RatingKey,
		item.MediaItem.Type, item.MediaItem.Title, item.MediaItem.Year, item.MediaItem.Thumb,
		item.MediaItem.ContentRating, item.MediaItem.Summary, item.MediaItem.UpdatedAt, item.MediaItem.AddedAt,
		item.MediaItem.ReleasedAt, string(guidsJSON), string(movieJSON), string(seriesJSON), string(mediaItemJSON))
	if err != nil {
		tx.Rollback()
		actionAddToMediaItemsTable.SetError("Failed to insert/update media item", err.Error(), map[string]any{
			"item": item,
		})
		return *actionAddToMediaItemsTable.Error
	}
	actionAddToMediaItemsTable.Complete()

	// Find the other PosterSet Items for this TMDB_ID and LibraryTitle
	// We need to do this to remove any Selected Types from this set
	// If there is none left, then delete that entry entirely
	actionFindExistingPosterSets := logAction.AddSubAction("Find Existing Poster Sets for Media Item", logging.LevelTrace)
	rows, err := tx.Query(`
		SELECT PosterSetID, LibraryTitle, SelectedTypes FROM PosterSets
		WHERE TMDB_ID = ? AND LibraryTitle = ?
	`, item.MediaItem.TMDB_ID, item.MediaItem.LibraryTitle)
	if err != nil {
		tx.Rollback()
		actionFindExistingPosterSets.SetError("Failed to query existing poster sets", err.Error(), map[string]any{
			"item": item,
		})
		return *actionFindExistingPosterSets.Error
	}
	defer rows.Close()
	actionFindExistingPosterSets.Complete()

	existingPosterSets := make(map[string][]string) // Map of PosterSetID to SelectedTypes
	for rows.Next() {
		var posterSetID string
		var selectedTypesStr string
		if err := rows.Scan(&posterSetID, &item.MediaItem.LibraryTitle, &selectedTypesStr); err != nil {
			tx.Rollback()
			logAction.SetError("Failed to scan existing poster sets", err.Error(), map[string]any{
				"item": item,
			})
			return *logAction.Error
		}
		selectedTypes := strings.Split(selectedTypesStr, ",")
		existingPosterSets[posterSetID] = selectedTypes
	}
	if err := rows.Err(); err != nil {
		tx.Rollback()
		logAction.SetError("Failed to iterate over existing poster sets", err.Error(), map[string]any{
			"item": item,
		})
		return *logAction.Error
	}

	// Remove SelectedTypes that are no longer present in the new PosterSets
	for existingID, existingTypes := range existingPosterSets {
		found := false
		for _, newPosterSet := range item.PosterSets {
			if newPosterSet.PosterSet.ID == existingID {
				found = true
				break
			}
		}
		if !found {
			// Remove the SelectedTypes from this PosterSet
			newTypes := []string{}
			for _, t := range existingTypes {
				keep := true
				for _, newPosterSet := range item.PosterSets {
					if slices.Contains(newPosterSet.SelectedTypes, t) {
						keep = false
					}
					if !keep {
						break
					}
				}
				if keep {
					newTypes = append(newTypes, t)
				}
			}
			if len(newTypes) == 0 {
				// No SelectedTypes left, delete the PosterSet entry
				actionDeletePosterSet := logAction.AddSubAction("Delete Poster Set with No SelectedTypes", logging.LevelTrace)
				_, err = tx.Exec(`
					DELETE FROM PosterSets
					WHERE PosterSetID = ? AND TMDB_ID = ? AND LibraryTitle = ?
				`, existingID, item.MediaItem.TMDB_ID, item.MediaItem.LibraryTitle)
				if err != nil {
					tx.Rollback()
					actionDeletePosterSet.SetError("Failed to delete poster set with no selected types", err.Error(), map[string]any{
						"item": item,
					})
					return *actionDeletePosterSet.Error
				}
				actionDeletePosterSet.Complete()
				actionDeletePosterSet.AppendResult("action", "deleted poster set, no selected types left")
				actionDeletePosterSet.AppendResult("poster_set_id", existingID)
				actionDeletePosterSet.AppendResult("tmdb_id", item.MediaItem.TMDB_ID)
				actionDeletePosterSet.AppendResult("library_title", item.MediaItem.LibraryTitle)
				actionDeletePosterSet.AppendResult("title", item.MediaItem.Title)
				actionDeletePosterSet.AppendResult("selected_types", existingTypes)

				// Delete from SavedItems table as well
				actionDeleteSavedItem := logAction.AddSubAction("Delete Saved Item for Poster Set with No SelectedTypes", logging.LevelTrace)
				_, err = tx.Exec(`
					DELETE FROM SavedItems
					WHERE PosterSetID = ? AND TMDB_ID = ? AND LibraryTitle = ?
				`, existingID, item.MediaItem.TMDB_ID, item.MediaItem.LibraryTitle)
				if err != nil {
					tx.Rollback()
					actionDeleteSavedItem.SetError("Failed to delete saved item for poster set with no selected types", err.Error(), map[string]any{
						"item": item,
					})
					return *actionDeleteSavedItem.Error
				}
				actionDeleteSavedItem.Complete()
				actionDeleteSavedItem.AppendResult("action", "deleted saved item, no selected types left")
				actionDeleteSavedItem.AppendResult("poster_set_id", existingID)
				actionDeleteSavedItem.AppendResult("title", item.MediaItem.Title)
				actionDeleteSavedItem.AppendResult("tmdb_id", item.MediaItem.TMDB_ID)
				actionDeleteSavedItem.AppendResult("library_title", item.MediaItem.LibraryTitle)

			} else {
				// Update the PosterSet with the new SelectedTypes
				actionUpdatePosterSet := logAction.AddSubAction("Update Poster Set with New SelectedTypes", logging.LevelTrace)
				_, err = tx.Exec(`
					UPDATE PosterSets
					SET SelectedTypes = ?
					WHERE PosterSetID = ? AND TMDB_ID = ? AND LibraryTitle = ?
				`, strings.Join(newTypes, ","), existingID, item.MediaItem.TMDB_ID, item.MediaItem.LibraryTitle)
				if err != nil {
					tx.Rollback()
					actionUpdatePosterSet.SetError("Failed to update poster set with new selected types", err.Error(), map[string]any{
						"item": item,
					})
					return *actionUpdatePosterSet.Error
				}
				actionUpdatePosterSet.Complete()
				actionUpdatePosterSet.AppendResult("action", "updated poster set with new selected types")
				actionUpdatePosterSet.AppendResult("poster_set_id", existingID)
				actionUpdatePosterSet.AppendResult("tmdb_id", item.MediaItem.TMDB_ID)
				actionUpdatePosterSet.AppendResult("library_title", item.MediaItem.LibraryTitle)
				actionUpdatePosterSet.AppendResult("title", item.MediaItem.Title)
				actionUpdatePosterSet.AppendResult("selected_types", newTypes)
			}
		}
	}

	// Insert each of the PosterSet items
	for _, posterSet := range item.PosterSets {
		if posterSet.ToDelete {
			// Skip poster sets marked for deletion
			continue
		}
		posterSetJSON, _ := json.Marshal(posterSet.PosterSet)
		now := time.Now().UTC().Format(time.RFC3339)
		actionAddToPosterSetsTable := logAction.AddSubAction("Insert/Update Poster Set in PosterSets Table", logging.LevelTrace)
		_, err = tx.Exec(`
        INSERT INTO PosterSets (
            PosterSetID, TMDB_ID, LibraryTitle, PosterSetUser, PosterSet_JSON, LastDownloaded, SelectedTypes, AutoDownload
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
        ON CONFLICT(PosterSetID, TMDB_ID, LibraryTitle) DO UPDATE SET
            TMDB_ID = excluded.TMDB_ID,
            LibraryTitle = excluded.LibraryTitle,
            PosterSetUser = excluded.PosterSetUser,
            PosterSet_JSON = excluded.PosterSet_JSON,
            LastDownloaded = excluded.LastDownloaded,
            SelectedTypes = excluded.SelectedTypes,
            AutoDownload = excluded.AutoDownload;
    `, posterSet.PosterSet.ID, item.MediaItem.TMDB_ID, item.MediaItem.LibraryTitle,
			posterSet.PosterSet.User.Name, string(posterSetJSON),
			now, strings.Join(posterSet.SelectedTypes, ","), posterSet.AutoDownload)
		if err != nil {
			tx.Rollback()
			actionAddToPosterSetsTable.SetError("Failed to insert poster set", err.Error(), map[string]any{
				"item": item,
			})
			return *actionAddToPosterSetsTable.Error
		}
		actionAddToPosterSetsTable.Complete()

		// Insert into SavedItems table
		actionAddToSavedItemsTable := logAction.AddSubAction("Insert Saved Item for Poster Set", logging.LevelTrace)
		_, err = tx.Exec(`
		INSERT OR IGNORE INTO SavedItems (TMDB_ID, LibraryTitle, PosterSetID) VALUES (?, ?, ?)
	`, item.MediaItem.TMDB_ID, item.MediaItem.LibraryTitle, posterSet.PosterSet.ID)
		if err != nil {
			tx.Rollback()
			actionAddToSavedItemsTable.SetError("Failed to insert saved item", err.Error(), map[string]any{
				"item": item,
			})
			return *actionAddToSavedItemsTable.Error
		}
		actionAddToSavedItemsTable.Complete()
	}

	// Commit the transaction
	actionCommitTx := logAction.AddSubAction("Commit DB Transaction", logging.LevelTrace)
	if err = tx.Commit(); err != nil {
		tx.Rollback()
		actionCommitTx.SetError("Failed to commit DB transaction", err.Error(), map[string]any{
			"item": item,
		})
		return *actionCommitTx.Error
	}
	actionCommitTx.Complete()

	transactionType := "inserted"
	if isUpdate {
		transactionType = "updated"
	}
	logAction.AppendResult("transaction_type", transactionType)
	logAction.AppendResult("title", item.MediaItem.Title)
	logAction.AppendResult("tmdb_id", item.MediaItem.TMDB_ID)
	logAction.AppendResult("library", item.MediaItem.LibraryTitle)

	return logging.LogErrorInfo{}
}
