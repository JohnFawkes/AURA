package api

import (
	"aura/internal/logging"
	"context"
)

func DB_Delete_PosterSet(ctx context.Context, posterSetID string, tmdbID string, libraryTitle string) logging.LogErrorInfo {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Deleting Poster Set from DB", logging.LevelInfo)
	defer logAction.Complete()

	// Start a DB transaction
	actionStartTx := logAction.AddSubAction("Start DB Transaction", logging.LevelTrace)
	tx, err := db.Begin()
	if err != nil {
		actionStartTx.SetError("Failed to begin DB transaction", "Could not start a transaction to delete poster set",
			map[string]any{
				"error":        err.Error(),
				"posterSetID":  posterSetID,
				"tmdbID":       tmdbID,
				"libraryTitle": libraryTitle,
			})
		return *actionStartTx.Error
	}
	actionStartTx.Complete()

	// Delete from SavedItems table
	actionDeleteSavedItems := logAction.AddSubAction("Delete from SavedItems", logging.LevelTrace)
	_, err = tx.Exec(`DELETE FROM SavedItems WHERE PosterSetID = ? AND TMDB_ID = ? AND LibraryTitle = ?;`, posterSetID, tmdbID, libraryTitle)
	if err != nil {
		tx.Rollback()
		actionDeleteSavedItems.SetError("Failed to delete from SavedItems", "Could not delete poster set from SavedItems table",
			map[string]any{
				"error":        err.Error(),
				"posterSetID":  posterSetID,
				"tmdbID":       tmdbID,
				"libraryTitle": libraryTitle,
			})
		return *actionDeleteSavedItems.Error
	}
	actionDeleteSavedItems.Complete()

	// Delete from PosterSets table
	actionDeletePosterSets := logAction.AddSubAction("Delete from PosterSets", logging.LevelTrace)
	_, err = tx.Exec(`DELETE FROM PosterSets WHERE PosterSetID = ? AND TMDB_ID = ? AND LibraryTitle = ?;`, posterSetID, tmdbID, libraryTitle)
	if err != nil {
		tx.Rollback()
		actionDeletePosterSets.SetError("Failed to delete from PosterSets", "Could not delete poster set from PosterSets table",
			map[string]any{
				"error":        err.Error(),
				"posterSetID":  posterSetID,
				"tmdbID":       tmdbID,
				"libraryTitle": libraryTitle,
			})
		return *actionDeletePosterSets.Error
	}
	actionDeletePosterSets.Complete()

	// Commit the transaction
	actionCommitTx := logAction.AddSubAction("Commit DB Transaction", logging.LevelTrace)
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		actionCommitTx.SetError("Failed to commit DB transaction", "Could not commit the transaction to delete poster set",
			map[string]any{
				"error":        err.Error(),
				"posterSetID":  posterSetID,
				"tmdbID":       tmdbID,
				"libraryTitle": libraryTitle,
			})
		return *actionCommitTx.Error
	}
	actionCommitTx.Complete()

	logAction.AppendResult("action", "deleted poster set from database")
	logAction.AppendResult("poster_set_id", posterSetID)
	logAction.AppendResult("tmdb_id", tmdbID)
	logAction.AppendResult("library_title", libraryTitle)

	return logging.LogErrorInfo{}
}
