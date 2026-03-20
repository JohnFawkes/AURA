package migration

import (
	"aura/database"
	"aura/logging"
	"context"
)

func migrate_3_to_4(ctx context.Context) (Err logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Migrating Database from v3 to v4", logging.LevelInfo)
	defer logAction.Complete()
	logging.LOGGER.Info().Timestamp().Int("From Version", 3).Int("To Version", 4).Msg("Starting database migration")

	Err = logging.LogErrorInfo{}

	// Create a backup of the current database
	backupErr := database.Backup(ctx, 3, 4)
	if backupErr.Message != "" {
		return backupErr
	}

	// Get DB connection
	conn, _, getDBConnErr := database.GetDBConnection(ctx)
	if getDBConnErr.Message != "" {
		return getDBConnErr
	}

	// Add a new column "auto_add_new_collection_items" to the SavedItems table with a default value of 0 (false)
	alterTableQuery := `ALTER TABLE SavedItems ADD COLUMN auto_add_new_collection_items INTEGER NOT NULL DEFAULT 0 CHECK (auto_add_new_collection_items IN (0,1));`
	_, err := conn.ExecContext(ctx, alterTableQuery)
	if err != nil {
		logAction.SetError("Failed to alter SavedItems table to add auto_add_new_collection_items column", "", map[string]any{"error": err.Error()})
		return *logAction.Error
	}

	logging.LOGGER.Info().Timestamp().Msg("Database migration v3.0 to v4.0 completed successfully")
	return Err
}
