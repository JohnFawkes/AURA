package database

import (
	"aura/logging"
	"context"
)

func (s *SQliteDB) CreateVersionTable(ctx context.Context) (Err logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Creating VERSION Table", logging.LevelDebug)
	defer logAction.Complete()

	query := `
	CREATE TABLE IF NOT EXISTS VERSION (
		version INTEGER NOT NULL
	);
	`
	_, err := s.conn.ExecContext(ctx, query)
	if err != nil {
		logAction.SetError("Failed to create VERSION table", err.Error(), map[string]any{
			"error": err.Error(),
			"query": query,
		})
		return *logAction.Error
	}

	return logging.LogErrorInfo{}
}
