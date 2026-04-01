package migration

import (
	"aura/database"
	"aura/logging"
	"context"
)

func migrate_4_to_5(ctx context.Context) (Err logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Migrating Database from v4 to v5", logging.LevelInfo)
	defer logAction.Complete()
	logging.LOGGER.Info().Timestamp().Int("From Version", 4).Int("To Version", 5).Msg("Starting database migration")

	Err = logging.LogErrorInfo{}

	// Create a backup of the current database
	backupErr := database.Backup(ctx, 4, 5)
	if backupErr.Message != "" {
		return backupErr
	}

	// Get DB connection
	conn, _, getDBConnErr := database.GetDBConnection(ctx)
	if getDBConnErr.Message != "" {
		return getDBConnErr
	}

	// Check if the "sets" column already exists to avoid duplicate column error
	setsColumnExists, checkColumnErr := checkColumnExists(ctx, "IgnoredItems", "sets")
	if checkColumnErr.Message != "" {
		return checkColumnErr
	}

	if !setsColumnExists {
		// Since SQLite doesn't support altering CHECK constraints directly, we need to recreate the table with the new schema. This involves:
		// 1. Renaming the existing table
		// 2. Creating a new table with the updated schema
		// 3. Copying data from the old table to the new table
		// 4. Dropping the old table

		tx, err := conn.BeginTx(ctx, nil)
		if err != nil {
			logAction.SetError("Failed to begin transaction for altering IgnoredItems table", "", map[string]any{"error": err.Error()})
			return *logAction.Error
		}

		// Step 1: Rename the existing table
		renameTableQuery := `ALTER TABLE IgnoredItems RENAME TO IgnoredItems_old;`
		_, err = tx.ExecContext(ctx, renameTableQuery)
		if err != nil {
			tx.Rollback()
			logAction.SetError("Failed to rename IgnoredItems table", "", map[string]any{"error": err.Error()})
			return *logAction.Error
		}

		// Step 2: Create a new table with the updated schema
		createTableQuery := `
			CREATE TABLE IgnoredItems (
				tmdb_id INTEGER NOT NULL,
				library_title TEXT NOT NULL,
				mode TEXT NOT NULL CHECK (mode IN ('always','until-set-available','until-new-set-available')),
				current_sets TEXT NOT NULL DEFAULT '[]',
				PRIMARY KEY (tmdb_id, library_title)
			) WITHOUT ROWID;
		`
		_, err = tx.ExecContext(ctx, createTableQuery)
		if err != nil {
			tx.Rollback()
			logAction.SetError("Failed to create new IgnoredItems table with updated schema", "", map[string]any{"error": err.Error()})
			return *logAction.Error
		}

		// Step 3: Copy data from the old table to the new table
		// Convert 'temp' mode to 'until-set-available' for existing records
		copyDataQuery := `
			INSERT INTO IgnoredItems (tmdb_id, library_title, mode)
			SELECT tmdb_id, library_title,
				CASE 
					WHEN mode = 'temp' THEN 'until-set-available'
					ELSE mode
				END as mode
			FROM IgnoredItems_old;
		`
		_, err = tx.ExecContext(ctx, copyDataQuery)
		if err != nil {
			tx.Rollback()
			logAction.SetError("Failed to copy data from old IgnoredItems table to new table", "", map[string]any{"error": err.Error()})
			return *logAction.Error
		}

		// Step 4: Drop the old table
		dropTableQuery := `DROP TABLE IgnoredItems_old;`
		_, err = tx.ExecContext(ctx, dropTableQuery)
		if err != nil {
			tx.Rollback()
			logAction.SetError("Failed to drop old IgnoredItems table", "", map[string]any{"error": err.Error()})
			return *logAction.Error
		}

		err = tx.Commit()
		if err != nil {
			logAction.SetError("Failed to commit transaction for altering IgnoredItems table", "", map[string]any{"error": err.Error()})
			return *logAction.Error
		}
	}

	logging.LOGGER.Info().Timestamp().Msg("Database migration v4.0 to v5.0 completed successfully")
	return Err
}
