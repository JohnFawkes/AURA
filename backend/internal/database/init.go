package database

import (
	"database/sql"
	"fmt"
	"os"
	"path"
	"poster-setter/internal/logging"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func InitDB() bool {
	var err error

	// Use an environment variable to determine the config path
	// By default, it will use /config
	// This is useful for testing and local development
	// In Docker, the config path is set to /config
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "/config"
	}
	dbPath := path.Join(configPath, "poster_setter.db")

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

	// Create the table if it doesn't exist
	createTableQuery := `
CREATE TABLE IF NOT EXISTS auto_downloader (
id INTEGER PRIMARY KEY UNIQUE NOT NULL,
plex JSON NOT NULL,
poster_set JSON NOT NULL,
selected_types TEXT NOT NULL,
auto_download BOOLEAN NOT NULL,
last_update DATETIME NOT NULL
);`

	_, err = db.Exec(createTableQuery)
	if err != nil {
		logging.LOG.Error(fmt.Sprintf("Failed to create table: %v", err))
		return false
	}

	return true
}
