package api

import (
	"aura/internal/logging"
	"fmt"
)

func DB_Delete_PosterSet(posterSetID, tmdbID, libraryTitle string) logging.StandardError {
	Err := logging.NewStandardError()

	// Start a DB transaction
	tx, err := db.Begin()
	if err != nil {
		Err.Message = "Failed to begin database transaction"
		Err.HelpText = "Check the details for more information"
		Err.Details = map[string]any{
			"error":        err.Error(),
			"posterSet":    posterSetID,
			"tmdbID":       tmdbID,
			"libraryTitle": libraryTitle,
		}
		return Err
	}

	// Delete from SavedItems table
	_, err = tx.Exec(`DELETE FROM SavedItems WHERE PosterSetID = ? AND TMDB_ID = ? AND LibraryTitle = ?;`, posterSetID, tmdbID, libraryTitle)
	if err != nil {
		tx.Rollback()
		Err.Message = "Failed to delete from SavedItems"
		Err.HelpText = "Transaction rolled back"
		Err.Details = map[string]any{"error": err.Error(), "PosterSetID": posterSetID, "TMDB_ID": tmdbID, "LibraryTitle": libraryTitle}
		return Err
	}

	// Delete from PosterSets table
	_, err = tx.Exec(`DELETE FROM PosterSets WHERE PosterSetID = ? AND TMDB_ID = ? AND LibraryTitle = ?;`, posterSetID, tmdbID, libraryTitle)
	if err != nil {
		tx.Rollback()
		Err.Message = "Failed to delete from PosterSets"
		Err.HelpText = "Transaction rolled back"
		Err.Details = map[string]any{"error": err.Error(), "PosterSetID": posterSetID, "TMDB_ID": tmdbID, "LibraryTitle": libraryTitle}
		return Err
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		Err.Message = "Failed to commit transaction"
		Err.HelpText = "Check the details for more information"
		Err.Details = map[string]any{"error": err.Error()}
		return Err
	}

	logging.LOG.Info(fmt.Sprintf("Deleted PosterSet '%s' for item (TMDB ID: %s | Library Title: %s) from database", posterSetID, tmdbID, libraryTitle))
	return Err
}
