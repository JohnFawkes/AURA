package api

import (
	"aura/internal/logging"
	"context"
)

func DB_DeleteItem(ctx context.Context, tmdbID string, libraryTitle string) logging.LogErrorInfo {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Deleting Item from DB", logging.LevelInfo)
	defer logAction.Complete()

	// Start a DB transaction
	actionStartTx := logAction.AddSubAction("Start DB Transaction", logging.LevelTrace)
	tx, err := db.Begin()
	if err != nil {
		actionStartTx.SetError("Failed to begin DB transaction", "Could not start a transaction to delete item",
			map[string]any{
				"error":        err.Error(),
				"tmdbID":       tmdbID,
				"libraryTitle": libraryTitle,
			})
		return *actionStartTx.Error
	}
	actionStartTx.Complete()

	// Delete from SavedItems table
	actionDeleteSavedItems := logAction.AddSubAction("Delete from SavedItems", logging.LevelTrace)
	_, err = tx.Exec(`DELETE FROM SavedItems WHERE TMDB_ID = ? AND LibraryTitle = ?;`, tmdbID, libraryTitle)
	if err != nil {
		tx.Rollback()
		actionDeleteSavedItems.SetError("Failed to delete from SavedItems", "Could not delete item from SavedItems table",
			map[string]any{
				"error":        err.Error(),
				"tmdbID":       tmdbID,
				"libraryTitle": libraryTitle,
			})
		return *actionDeleteSavedItems.Error
	}
	actionDeleteSavedItems.Complete()

	// Delete from PosterSets table
	actionDeletePosterSets := logAction.AddSubAction("Delete from PosterSets", logging.LevelTrace)
	_, err = tx.Exec(`DELETE FROM PosterSets WHERE TMDB_ID = ? AND LibraryTitle = ?;`, tmdbID, libraryTitle)
	if err != nil {
		tx.Rollback()
		actionDeletePosterSets.SetError("Failed to delete from PosterSets", "Could not delete item from PosterSets table",
			map[string]any{
				"error":        err.Error(),
				"tmdbID":       tmdbID,
				"libraryTitle": libraryTitle,
			})
		return *actionDeletePosterSets.Error
	}
	actionDeletePosterSets.Complete()

	// Delete from MediaItems table
	actionDeleteMediaItems := logAction.AddSubAction("Delete from MediaItems", logging.LevelTrace)
	_, err = tx.Exec(`DELETE FROM MediaItems WHERE TMDB_ID = ? AND LibraryTitle = ?;`, tmdbID, libraryTitle)
	if err != nil {
		tx.Rollback()
		actionDeleteMediaItems.SetError("Failed to delete from MediaItems", "Could not delete item from MediaItems table",
			map[string]any{
				"error":        err.Error(),
				"tmdbID":       tmdbID,
				"libraryTitle": libraryTitle,
			})
		return *actionDeleteMediaItems.Error
	}
	actionDeleteMediaItems.Complete()

	// Commit the transaction
	actionCommitTx := logAction.AddSubAction("Commit DB Transaction", logging.LevelTrace)
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		actionCommitTx.SetError("Failed to commit DB transaction", "Could not commit the transaction to delete item",
			map[string]any{
				"error":        err.Error(),
				"tmdbID":       tmdbID,
				"libraryTitle": libraryTitle,
			})
		return *actionCommitTx.Error
	}
	actionCommitTx.Complete()

	logAction.AppendResult("tmdbID", tmdbID)
	logAction.AppendResult("libraryTitle", libraryTitle)
	logAction.AppendResult("action", "delete")

	return logging.LogErrorInfo{}
}
