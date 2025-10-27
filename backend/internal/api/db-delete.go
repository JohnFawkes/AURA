package api

import (
	"aura/internal/logging"
)

func DB_DeleteItem(TMDB_ID, LibraryTitle string) logging.StandardError {
	Err := logging.NewStandardError()

	// Start a DB transaction
	tx, err := db.Begin()
	if err != nil {
		Err.Message = "Failed to begin database transaction"
		Err.HelpText = "Check the details for more information"
		Err.Details = map[string]any{
			"error":   err.Error(),
			"item":    TMDB_ID,
			"library": LibraryTitle,
		}
		return Err
	}

	// Delete from SavedItems table
	_, err = tx.Exec(`DELETE FROM SavedItems WHERE TMDB_ID = ? AND LibraryTitle = ?;`, TMDB_ID, LibraryTitle)
	if err != nil {
		tx.Rollback()
		Err.Message = "Failed to delete from SavedItems"
		Err.HelpText = "Transaction rolled back"
		Err.Details = map[string]any{"error": err.Error(), "TMDB_ID": TMDB_ID, "LibraryTitle": LibraryTitle}
		return Err
	}

	// Delete from PosterSets table
	_, err = tx.Exec(`DELETE FROM PosterSets WHERE TMDB_ID = ? AND LibraryTitle = ?;`, TMDB_ID, LibraryTitle)
	if err != nil {
		tx.Rollback()
		Err.Message = "Failed to delete from PosterSets"
		Err.HelpText = "Transaction rolled back"
		Err.Details = map[string]any{"error": err.Error(), "TMDB_ID": TMDB_ID, "LibraryTitle": LibraryTitle}
		return Err
	}

	// Delete from MediaItems table
	_, err = tx.Exec(`DELETE FROM MediaItems WHERE TMDB_ID = ? AND LibraryTitle = ?;`, TMDB_ID, LibraryTitle)
	if err != nil {
		tx.Rollback()
		Err.Message = "Failed to delete from MediaItems"
		Err.HelpText = "Transaction rolled back"
		Err.Details = map[string]any{"error": err.Error(), "TMDB_ID": TMDB_ID, "LibraryTitle": LibraryTitle}
		return Err
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		Err.Message = "Failed to commit transaction"
		Err.HelpText = "Check the details for more information"
		Err.Details = map[string]any{"error": err.Error()}
		return Err
	}

	logging.LOG.Info("Deleted item (TMDB ID: " + TMDB_ID + " | Library Title: " + LibraryTitle + ") from database")
	return Err
}
