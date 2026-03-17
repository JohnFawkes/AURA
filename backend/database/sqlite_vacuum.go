package database

import (
	"aura/logging"
	"context"
	"fmt"
)

func getDBSizeMiB(ctx context.Context, s *SQliteDB) (sizeMiB float64, err error) {
	var pageCount int64
	var pageSize int64

	err = s.conn.QueryRowContext(ctx, "PRAGMA page_count;").Scan(&pageCount)
	if err != nil {
		return 0, err
	}

	err = s.conn.QueryRowContext(ctx, "PRAGMA page_size;").Scan(&pageSize)
	if err != nil {
		return 0, err
	}

	return float64(pageCount*pageSize) / (1024.0 * 1024.0), nil
}

func (s *SQliteDB) Vacuum(ctx context.Context) (Err logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Running VACUUM on Database", logging.LevelInfo)
	defer logAction.Complete()

	dbSizeCurrent, err := getDBSizeMiB(ctx, s)
	if err != nil {
		logAction.SetError("Failed to get current database size", err.Error(), map[string]any{
			"error": err.Error(),
		})
		return *logAction.Error
	}
	logAction.AppendResult("db_size_current", dbSizeCurrent)

	// Check if freelist_count is greater than 5000
	var freeListCount int64
	err = s.conn.QueryRowContext(ctx, "PRAGMA freelist_count;").Scan(&freeListCount)
	if err != nil {
		logAction.SetError("Failed to get freelist_count", err.Error(), map[string]any{
			"error": err.Error(),
		})
		return *logAction.Error
	}

	if freeListCount < 5000 {
		logging.Dev().Timestamp().Int64("freelist_count", freeListCount).Msg("Checked freelist_count before VACUUM")
		logAction.AppendResult("db_size_new", dbSizeCurrent)
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

	dbSizeNew, err := getDBSizeMiB(ctx, s)
	if err != nil {
		logAction.SetError("Failed to get new database size", err.Error(), map[string]any{
			"error": err.Error(),
		})
		return *logAction.Error
	}

	logAction.AppendResult("db_size_new", dbSizeNew)
	logAction.AppendResult("freelist_count", freeListCount)
	logAction.AppendResult("vacuum_performed", true)
	logging.LOGGER.Info().Timestamp().Str("db_size_current", fmt.Sprintf("%.2f MiB", dbSizeCurrent)).Str("db_size_new", fmt.Sprintf("%.2f MiB", dbSizeNew)).Int64("freelist_count", freeListCount).Msg("VACUUM completed")
	return logging.LogErrorInfo{}
}
