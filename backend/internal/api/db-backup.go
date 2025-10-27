package api

import (
	"aura/internal/logging"
	"fmt"
	"io"
	"os"
	"time"
)

func DB_MakeBackup(dbPath string, dbVersion int) logging.StandardError {
	// Create a StandardError to return
	Err := logging.NewStandardError()

	// Check if the database file exists
	// If it doesn't exist, there's nothing to back up
	// This can happen on first run
	// So we just return nil (no error)
	// and let the migration create the database fresh
	_, err := os.Stat(dbPath)
	if os.IsNotExist(err) {
		return Err
	} else if err != nil {
		Err.Message = "Failed to access database file for backup"
		Err.HelpText = "Ensure the database file path is correct and accessible."
		Err.Details = map[string]any{
			"error": err.Error(),
			"path":  dbPath,
		}
		return Err
	}

	// Create a backup file path with version and timestamp
	timestamp := time.Now().Format("20060102_150405")
	backupPath := fmt.Sprintf("%s_backup_v%d_%s.db", dbPath[:len(dbPath)-3], dbVersion, timestamp)

	// Copy the database file to the backup location
	sourceFile, err := os.Open(dbPath)
	if err != nil {
		Err.Message = "Failed to open database file for backup"
		Err.HelpText = "Ensure the database file path is correct and accessible."
		Err.Details = map[string]any{
			"error": err.Error(),
			"path":  dbPath,
		}
		return Err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(backupPath)
	if err != nil {
		Err.Message = "Failed to create backup file"
		Err.HelpText = "Ensure the backup file path is correct and writable."
		Err.Details = map[string]any{
			"error": err.Error(),
			"path":  backupPath,
		}
		return Err
	}
	defer destFile.Close()

	// Copy the contents from the source file to the destination file
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		Err.Message = "Failed to copy database file contents"
		Err.HelpText = "Ensure the source and destination file paths are correct and accessible."
		Err.Details = map[string]any{
			"error": err.Error(),
			"path":  backupPath,
		}
		return Err
	}
	logging.LOG.Info("Database backup created at: " + backupPath)
	return Err
}
