package database

import (
	"aura/logging"
	"aura/models"
	"context"
)

func (s *SQliteDB) GetAllMediaItems(ctx context.Context) (items []models.MediaItem, logErr logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Retrieving all MediaItems from SQLite database", logging.LevelDebug)
	defer logAction.Complete()

	items = []models.MediaItem{}
	logErr = logging.LogErrorInfo{}

	// Start a transaction
	tx, err := s.conn.BeginTx(ctx, nil)
	if err != nil {
		logAction.SetError("Failed to start transaction", "", map[string]any{"error": err.Error()})
		return items, *logAction.Error
	}
	defer func() { _ = tx.Rollback() }()

	// Query all MediaItems
	rows, err := tx.QueryContext(ctx, `
		SELECT tmdb_id, library_title, rating_key
		FROM MediaItems;
	`)
	if err != nil {
		logAction.SetError("Failed to query MediaItems", "", map[string]any{"error": err.Error()})
		return items, *logAction.Error
	}
	defer rows.Close()

	// Iterate through the rows
	for rows.Next() {
		var item models.MediaItem
		if err := rows.Scan(&item.TMDB_ID, &item.LibraryTitle, &item.RatingKey); err != nil {
			logAction.SetError("Failed to scan MediaItem row", "", map[string]any{"error": err.Error()})
			return items, *logAction.Error
		}
		items = append(items, item)
	}

	return items, logErr
}
