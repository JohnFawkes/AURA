package database

import (
	"aura/internal/logging"
	"database/sql"
	"fmt"
	"os"
	"path"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func InitDB() bool {
	logging.LOG.Debug("Initializing database...")
	var err error

	// Use an environment variable to determine the config path
	// By default, it will use /config
	// This is useful for testing and local development
	// In Docker, the config path is set to /config
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "/config"
	}
	dbPath := path.Join(configPath, "AURA.db")

	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		logging.LOG.Error(fmt.Sprintf("Failed to open database: %v", err))
		return false
	}

	// Set the timezone to your local timezone
	location, err := time.LoadLocation("America/New_York") // Replace with your timezone
	if err != nil {
		logging.LOG.Error(fmt.Sprintf("Failed to load timezone: %v", err))
		return false
	}
	time.Local = location // Set the default timezone for the application

	// Create the SavedItems table
	createSavedItemsTableQuery := `
CREATE TABLE IF NOT EXISTS SavedItems (
    media_item_id TEXT NOT NULL,
    media_item JSON NOT NULL,
	poster_set_id TEXT NOT NULL,
	poster_set JSON NOT NULL,
	selected_types TEXT NOT NULL,
	auto_download BOOLEAN NOT NULL,
	last_update DATETIME NOT NULL,
	PRIMARY KEY (media_item_id, poster_set_id)
);`
	_, err = db.Exec(createSavedItemsTableQuery)
	if err != nil {
		logging.LOG.Error(fmt.Sprintf("Failed to create SavedItems table: %v", err))
		return false
	}

	logging.LOG.Info("Successfully initialized the database at: " + dbPath)
	return true
}
