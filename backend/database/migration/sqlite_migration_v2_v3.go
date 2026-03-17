package migration

import (
	"aura/database"
	"aura/logging"
	"context"
)

func migrate_2_to_3(ctx context.Context) (Err logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Migrating Database from v2 to v3", logging.LevelInfo)
	defer logAction.Complete()
	logging.LOGGER.Info().Timestamp().Int("From Version", 2).Int("To Version", 3).Msg("Starting database migration")

	Err = logging.LogErrorInfo{}

	// Create a backup of the current database
	backupErr := database.Backup(ctx, 2, 3)
	if backupErr.Message != "" {
		return backupErr
	}

	// Get DB connection
	conn, _, getDBConnErr := database.GetDBConnection(ctx)
	if getDBConnErr.Message != "" {
		return getDBConnErr
	}

	// Add a new column "on_server" to the MediaItems table with a default value of 0 (false)
	alterTableQuery := `ALTER TABLE MediaItems ADD COLUMN on_server INTEGER NOT NULL DEFAULT 0;`
	_, err := conn.ExecContext(ctx, alterTableQuery)
	if err != nil {
		logAction.SetError("Failed to alter MediaItems table to add on_server column", "", map[string]any{"error": err.Error()})
		return *logAction.Error
	}

	logging.LOGGER.Info().Timestamp().Msg("Database migration v2.0 to v3.0 completed successfully")
	return Err
}
