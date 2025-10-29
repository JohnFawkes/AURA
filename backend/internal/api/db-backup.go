package api

import (
	"aura/internal/logging"
	"context"
	"fmt"
	"io"
	"os"
	"time"
)

func DB_MakeBackup(ctx context.Context, dbPath string, dbVersion int) logging.LogErrorInfo {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Creating database backup", logging.LevelDebug)
	defer logAction.Complete()

	// Check if the database file exists
	// If it doesn't exist, there's nothing to back up
	// This can happen on first run
	// So we just return nil (no error)
	// and let the migration create the database fresh
	_, err := os.Stat(dbPath)
	if os.IsNotExist(err) {
		return logging.LogErrorInfo{}
	} else if err != nil {
		logAction.SetError("Failed to stat database file for backup",
			"Ensure the database path is accessible.",
			map[string]any{
				"error": err.Error(),
				"path":  dbPath,
			})
		return *logAction.Error
	}

	// Create a backup file path with version and timestamp
	timestamp := time.Now().Format("20060102_150405")
	backupPath := fmt.Sprintf("%s_backup_v%d_%s.db", dbPath[:len(dbPath)-3], dbVersion, timestamp)

	// Copy the database file to the backup location
	actionBackupFile := logAction.AddSubAction("Backing up database file", logging.LevelTrace)
	sourceFile, err := os.Open(dbPath)
	if err != nil {
		actionBackupFile.SetError("Failed to open source database file for backup",
			"Ensure the database path is accessible.",
			map[string]any{
				"error": err.Error(),
				"path":  dbPath,
			})
		return *actionBackupFile.Error
	}
	defer sourceFile.Close()

	// Create the backup file
	backupFile, err := os.Create(backupPath)
	if err != nil {
		actionBackupFile.SetError("Failed to create backup database file",
			"Ensure the destination path is accessible and writable.",
			map[string]any{
				"error": err.Error(),
				"path":  backupPath,
			})
		return *actionBackupFile.Error
	}
	defer backupFile.Close()

	// Copy the contents from the source file to the destination file
	_, err = io.Copy(backupFile, sourceFile)
	if err != nil {
		actionBackupFile.SetError("Failed to copy database file for backup",
			"Ensure there is sufficient disk space and permissions.",
			map[string]any{
				"error":       err.Error(),
				"source":      dbPath,
				"destination": backupPath,
			})
		return *actionBackupFile.Error
	}
	actionBackupFile.Complete()

	return logging.LogErrorInfo{}
}
