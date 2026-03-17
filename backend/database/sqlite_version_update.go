package database

import (
	"aura/logging"
	"context"
	"fmt"
)

func (s *SQliteDB) UpdateVersionTable(ctx context.Context, newVersion int) (Err logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, fmt.Sprintf("Updating Database Version to %d", newVersion), logging.LevelInfo)
	defer logAction.Complete()

	query := `UPDATE VERSION SET version = ?;`
	res, err := s.conn.ExecContext(ctx, query, newVersion)
	if err != nil {
		logAction.SetError(
			"Failed to update VERSION table",
			"Ensure the database is accessible and not corrupted.",
			map[string]any{
				"error": err.Error(),
				"query": query,
			})
		return *logAction.Error
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		logAction.SetError(
			"Failed to retrieve rows affected after updating VERSION table",
			"Ensure the database is accessible and not corrupted.",
			map[string]any{
				"error": err.Error(),
			})
		return *logAction.Error
	}

	if rowsAffected == 0 {
		// No rows updated, insert new version
		insertQuery := `INSERT INTO VERSION (version) VALUES (?);`
		_, err := s.conn.ExecContext(ctx, insertQuery, newVersion)
		if err != nil {
			logAction.SetError(
				"Failed to insert new version into VERSION table",
				"Ensure the database is accessible and not corrupted.",
				map[string]any{
					"error": err.Error(),
					"query": insertQuery,
				})
			return *logAction.Error
		}
	}

	return logging.LogErrorInfo{}
}
