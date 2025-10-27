package api

import (
	"aura/internal/logging"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

func DB_InsertAllInfoIntoTables(item DBMediaItemWithPosterSets) logging.StandardError {
	Err := logging.NewStandardError()

	// Start a DB transaction
	tx, err := db.Begin()
	if err != nil {
		Err.Message = "Failed to begin database transaction"
		Err.HelpText = "Check the details for more information"
		Err.Details = map[string]any{
			"error": err.Error(),
			"item":  item,
		}
		return Err
	}

	// Remove any "ignore" poster sets for this TMDB_ID and LibraryTitle from the database
	_, err = tx.Exec(`
        DELETE FROM PosterSets
        WHERE PosterSetID = ? AND TMDB_ID = ? AND LibraryTitle = ?
    `, "ignore", item.MediaItem.TMDB_ID, item.MediaItem.LibraryTitle)
	if err != nil {
		tx.Rollback()
		Err.Message = "Failed to remove ignore PosterSet from database"
		Err.HelpText = "Transaction rolled back"
		Err.Details = map[string]any{"error": err.Error(), "item": item}
		return Err
	}

	// Remove any ignore entries from the SavedItems table
	_, err = tx.Exec(`
		DELETE FROM SavedItems
		WHERE PosterSetID = ? AND TMDB_ID = ? AND LibraryTitle = ?
	`, "ignore", item.MediaItem.TMDB_ID, item.MediaItem.LibraryTitle)
	if err != nil {
		tx.Rollback()
		Err.Message = "Failed to remove ignore SavedItem from database"
		Err.HelpText = "Transaction rolled back"
		Err.Details = map[string]any{"error": err.Error(), "item": item}
		return Err
	}

	// Check if the media item already exists
	isUpdate := false
	var exists int
	err = tx.QueryRow(`
    SELECT COUNT(*) FROM MediaItems WHERE TMDB_ID = ? AND LibraryTitle = ?
`, item.MediaItem.TMDB_ID, item.MediaItem.LibraryTitle).Scan(&exists)
	if err != nil {
		tx.Rollback()
		Err.Message = "Failed to check for existing MediaItem"
		Err.HelpText = "Transaction rolled back"
		Err.Details = map[string]any{"error": err.Error(), "item": item.MediaItem}
		return Err
	}
	if exists > 0 {
		isUpdate = true
	}

	// Insert into MediaItems table
	mediaItemJSON, _ := json.Marshal(item.MediaItem)
	guidsJSON, _ := json.Marshal(item.MediaItem.Guids)
	movieJSON, _ := json.Marshal(item.MediaItem.Movie)
	seriesJSON, _ := json.Marshal(item.MediaItem.Series)
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
		Err.Message = "Failed to insert MediaItem"
		Err.HelpText = "Transaction rolled back"
		Err.Details = map[string]any{"error": err.Error(), "item": item.MediaItem}
		return Err
	}

	// Find the other PosterSet Items for this TMDB_ID and LibraryTitle
	// We need to do this to remove any Selected Types from this set
	// If there is none left, then delete that entry entirely
	rows, err := tx.Query(`
		SELECT PosterSetID, LibraryTitle, SelectedTypes FROM PosterSets
		WHERE TMDB_ID = ? AND LibraryTitle = ?
	`, item.MediaItem.TMDB_ID, item.MediaItem.LibraryTitle)
	if err != nil {
		tx.Rollback()
		Err.Message = "Failed to query existing PosterSets"
		Err.HelpText = "Transaction rolled back"
		Err.Details = map[string]any{"error": err.Error(), "item": item}
		return Err
	}
	defer rows.Close()

	existingPosterSets := make(map[string][]string) // Map of PosterSetID to SelectedTypes
	for rows.Next() {
		var posterSetID string
		var selectedTypesStr string
		if err := rows.Scan(&posterSetID, &item.MediaItem.LibraryTitle, &selectedTypesStr); err != nil {
			tx.Rollback()
			Err.Message = "Failed to scan PosterSet row"
			Err.HelpText = "Transaction rolled back"
			Err.Details = map[string]any{"error": err.Error(), "item": item}
			return Err
		}
		selectedTypes := strings.Split(selectedTypesStr, ",")
		existingPosterSets[posterSetID] = selectedTypes
	}
	if err := rows.Err(); err != nil {
		tx.Rollback()
		Err.Message = "Error iterating over PosterSet rows"
		Err.HelpText = "Transaction rolled back"
		Err.Details = map[string]any{"error": err.Error(), "item": item}
		return Err
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
					for _, newType := range newPosterSet.SelectedTypes {
						if t == newType {
							keep = false
							break
						}
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
				_, err = tx.Exec(`
					DELETE FROM PosterSets
					WHERE PosterSetID = ? AND TMDB_ID = ? AND LibraryTitle = ?
				`, existingID, item.MediaItem.TMDB_ID, item.MediaItem.LibraryTitle)
				if err != nil {
					tx.Rollback()
					Err.Message = "Failed to delete PosterSet with no SelectedTypes"
					Err.HelpText = "Transaction rolled back"
					Err.Details = map[string]any{"error": err.Error(), "item": item}
					return Err
				}
				logging.LOG.Info(fmt.Sprintf("Deleted PosterSet '%s' for '%s' (TMDB_ID: %s | Library Title: %s) as it has no SelectedTypes left", existingID, item.MediaItem.Title, item.MediaItem.TMDB_ID, item.MediaItem.LibraryTitle))

				// Delete from SavedItems table as well
				_, err = tx.Exec(`
					DELETE FROM SavedItems
					WHERE PosterSetID = ? AND TMDB_ID = ? AND LibraryTitle = ?
				`, existingID, item.MediaItem.TMDB_ID, item.MediaItem.LibraryTitle)
				if err != nil {
					tx.Rollback()
					Err.Message = "Failed to delete SavedItem for removed PosterSet"
					Err.HelpText = "Transaction rolled back"
					Err.Details = map[string]any{"error": err.Error(), "item": item}
					return Err
				}
				logging.LOG.Info(fmt.Sprintf("Deleted SavedItem for PosterSet '%s' for '%s' (TMDB_ID: %s | Library Title: %s)", existingID, item.MediaItem.Title, item.MediaItem.TMDB_ID, item.MediaItem.LibraryTitle))
			} else {
				// Update the PosterSet with the new SelectedTypes
				_, err = tx.Exec(`
					UPDATE PosterSets
					SET SelectedTypes = ?
					WHERE PosterSetID = ? AND TMDB_ID = ? AND LibraryTitle = ?
				`, strings.Join(newTypes, ","), existingID, item.MediaItem.TMDB_ID, item.MediaItem.LibraryTitle)
				if err != nil {
					tx.Rollback()
					Err.Message = "Failed to update PosterSet SelectedTypes"
					Err.HelpText = "Transaction rolled back"
					Err.Details = map[string]any{"error": err.Error(), "item": item}
					return Err
				}
				logging.LOG.Info(fmt.Sprintf("Updated PosterSet '%s' for '%s' (TMDB_ID: %s | Library Title: %s) to remove obsolete SelectedTypes", existingID, item.MediaItem.Title, item.MediaItem.TMDB_ID, item.MediaItem.LibraryTitle))
			}
		}
	}

	// Insert each of the PosterSet items
	for _, posterSet := range item.PosterSets {
		posterSetJSON, _ := json.Marshal(posterSet.PosterSet)
		now := time.Now().UTC().Format(time.RFC3339)
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
			Err.Message = "Failed to insert PosterSet"
			Err.HelpText = "Transaction rolled back"
			Err.Details = map[string]any{"error": err.Error(), "item": posterSet}
			return Err
		}

		// Insert into SavedItems table
		_, err = tx.Exec(`
		INSERT OR IGNORE INTO SavedItems (TMDB_ID, LibraryTitle, PosterSetID) VALUES (?, ?, ?)
	`, item.MediaItem.TMDB_ID, item.MediaItem.LibraryTitle, posterSet.PosterSet.ID)
		if err != nil {
			tx.Rollback()
			Err.Message = "Failed to insert SavedItem"
			Err.HelpText = "Transaction rolled back"
			Err.Details = map[string]any{"error": err.Error(), "item": item}
			return Err
		}
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		Err.Message = "Failed to commit transaction"
		Err.HelpText = "Transaction rolled back"
		Err.Details = map[string]any{"error": err.Error(), "item": item}
		return Err
	}

	transactionType := "inserted"
	if isUpdate {
		transactionType = "updated"
	}
	logging.LOG.Info(fmt.Sprintf("Successfully %s '%s' (TMDB_ID: %s | Library Title %s) in the DB", transactionType, item.MediaItem.Title, item.MediaItem.TMDB_ID, item.MediaItem.LibraryTitle))

	return Err
}
