package database

import (
	"aura/logging"
	"context"
)

func (s *SQliteDB) Vacuum(ctx context.Context) (Err logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Running VACUUM on Database", logging.LevelInfo)
	defer logAction.Complete()

	// Check if freelist_count is greater than 10000
	var freeListCount int64
	err := s.conn.QueryRowContext(ctx, "PRAGMA freelist_count;").Scan(&freeListCount)
	if err != nil {
		logAction.SetError("Failed to get freelist_count", err.Error(), map[string]any{
			"error": err.Error(),
		})
		return *logAction.Error
	}

	if freeListCount < 10000 {
		logAction.AppendResult("freelist_count", freeListCount)
		logAction.AppendResult("vacuum_performed", false)
		return logging.LogErrorInfo{}
	}

	// Perform VACUUM
	_, err = s.conn.ExecContext(ctx, "VACUUM;")
	if err != nil {
		logAction.SetError("Failed to VACUUM database", err.Error(), map[string]any{
			"error": err.Error(),
		})
		return *logAction.Error
	}

	logAction.AppendResult("freelist_count", freeListCount)
	logAction.AppendResult("vacuum_performed", true)
	return logging.LogErrorInfo{}
}
