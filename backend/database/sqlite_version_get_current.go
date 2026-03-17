package database

import (
	"aura/logging"
	"context"
	"database/sql"
)

func (s *SQliteDB) GetCurrentVersion(ctx context.Context) (version int, Err logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Getting Current Database Version", logging.LevelDebug)
	defer logAction.Complete()

	version = 0

	// Check if the VERSION table exists
	var tableName string
	checkTableErr := s.conn.QueryRowContext(ctx, "SELECT name FROM sqlite_master WHERE type='table' AND name='VERSION';").Scan(&tableName)
	if checkTableErr == sql.ErrNoRows {
		logAction.AppendWarning("VERSION table does not exist. Assuming version 0.", nil)
		return version, logging.LogErrorInfo{}
	} else if checkTableErr != nil {
		logAction.SetError("Failed to check for VERSION table", "", map[string]any{
			"error": checkTableErr.Error(),
		})
		return version, *logAction.Error
	} else {
		// VERSION table exists, get the current version number
		getVersionErr := s.conn.QueryRowContext(ctx, "SELECT version FROM VERSION;").Scan(&version)
		if getVersionErr != nil {
			logAction.SetError("Failed to get current database version", "", map[string]any{
				"error": getVersionErr.Error(),
			})
			return version, *logAction.Error
		}
	}

	return version, logging.LogErrorInfo{}
}
