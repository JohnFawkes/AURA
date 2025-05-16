package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"poster-setter/internal/logging"
	"poster-setter/internal/modals"
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

	// Create the Media_Item table for media items
	createMediaTableQuery := `
CREATE TABLE IF NOT EXISTS Media_Item (
    id TEXT PRIMARY KEY UNIQUE NOT NULL,
    media_item JSON NOT NULL
);`
	_, err = db.Exec(createMediaTableQuery)
	if err != nil {
		logging.LOG.Error(fmt.Sprintf("Failed to create Media_Item table: %v", err))
		return false
	}

	// Create the Poster_Sets table for poster sets related to a media item
	createPosterSetsQuery := `
CREATE TABLE IF NOT EXISTS Poster_Sets (
    id TEXT PRIMARY KEY UNIQUE NOT NULL,
    media_item_id TEXT NOT NULL,
    poster_set JSON NOT NULL,
	selected_types TEXT NOT NULL,
	auto_download BOOLEAN NOT NULL,
    last_update DATETIME NOT NULL,
    FOREIGN KEY (media_item_id) REFERENCES Media_Item(id) ON DELETE CASCADE
);`
	_, err = db.Exec(createPosterSetsQuery)
	if err != nil {
		logging.LOG.Error(fmt.Sprintf("Failed to create Poster_Sets table: %v", err))
		return false
	}

	migrateDatabase()

	return true
}

func checkIfMigrationNeeded() bool {
	// Check if a file called "poster_setter.db" exists in the config path
	// If it does, migration is needed
	// If it doesn't, migration is not needed
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "/config"
	}
	oldDBPath := path.Join(configPath, "poster_setter.db")
	if _, err := os.Stat(oldDBPath); os.IsNotExist(err) {
		// File does not exist, migration is not needed
		return false
	}
	// File exists, migration is needed
	return true
}

func migrateDatabase() {
	// Check if the migration is needed
	if !checkIfMigrationNeeded() {
		return
	}

	// Open the old database
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "/config"
	}
	oldDBPath := path.Join(configPath, "poster_setter.db")

	// Open the old database
	oldDB, err := sql.Open("sqlite3", oldDBPath)
	if err != nil {
		logging.LOG.Error(fmt.Sprintf("Failed to open old database: %v", err))
		return
	}
	defer oldDB.Close()

	// Run PRAGMA to see if the id column is a int or a string
	// Run PRAGMA to see if the id column is a int or a string
	pragmaQuery := "PRAGMA table_info(auto_downloader)"
	pragmaRows, err := oldDB.Query(pragmaQuery)
	if err != nil {
		logging.LOG.Error(fmt.Sprintf("Failed to run PRAGMA: %v", err))
		return
	}
	defer pragmaRows.Close()

	var idType string
	for pragmaRows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dfltValue sql.NullString
		err = pragmaRows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk)
		if err != nil {
			logging.LOG.Error(fmt.Sprintf("Failed to scan PRAGMA result: %v", err))
			return
		}
		if name == "id" {
			idType = ctype
		}
	}

	// Query all data from the old auto_downloader table
	rows, err := oldDB.Query(`
        SELECT id, media_item, poster_set, selected_types, auto_download, last_update
        FROM auto_downloader
    `)
	if err != nil {
		logging.LOG.Error(fmt.Sprintf("Failed to query old database: %v", err))
		return
	}
	defer rows.Close()

	// Assume migration is successful unless an error occurs.
	migrationSuccess := true

	for rows.Next() {
		var mediaItemID, mediaItemJSON, posterSetJSON, selectedTypes, lastUpdate string
		var autoDownload bool

		if idType == "INTEGER" {
			var oldIDInt int
			err = rows.Scan(&oldIDInt, &mediaItemJSON, &posterSetJSON, &selectedTypes, &autoDownload, &lastUpdate)
			if err != nil {
				logging.LOG.Error(fmt.Sprintf("Failed to scan row: %v", err))
				migrationSuccess = false
				continue
			}
			// Convert int to string
			mediaItemID = fmt.Sprintf("%d", oldIDInt)
		} else {
			err = rows.Scan(&mediaItemID, &mediaItemJSON, &posterSetJSON, &selectedTypes, &autoDownload, &lastUpdate)
			if err != nil {
				logging.LOG.Error(fmt.Sprintf("Failed to scan row: %v", err))
				migrationSuccess = false
				continue
			}
		}

		// Insert into Media_Item table if not already present.
		_, err = db.Exec(`
            INSERT OR IGNORE INTO Media_Item (id, media_item)
            VALUES (?, ?)`, mediaItemID, mediaItemJSON)
		if err != nil {
			logging.LOG.Error(fmt.Sprintf("Failed to insert into Media_Item: %v", err))
			migrationSuccess = false
			continue
		}

		// Unmarshal the poster set JSON into a struct
		// This is a placeholder, you should define the struct according to your JSON structure
		var posterSet modals.PosterSet
		err = json.Unmarshal([]byte(posterSetJSON), &posterSet)
		if err != nil {
			logging.LOG.Error(fmt.Sprintf("Failed to unmarshal poster set JSON: %v", err))
			migrationSuccess = false
			continue
		}

		// Get the ID of the poster set
		posterSetID := posterSet.ID
		if posterSetID == "" {
			logging.LOG.Warn("Poster set ID is empty, skipping insert")
			continue
		}

		// Insert the poster set into the Poster_Sets table
		_, err = db.Exec(`
            INSERT INTO Poster_Sets (id, media_item_id, poster_set, selected_types, auto_download, last_update)
            VALUES (?, ?, ?, ?, ?, ?)`, posterSetID, mediaItemID, posterSetJSON, selectedTypes, autoDownload, lastUpdate)
		if err != nil {
			logging.LOG.Error(fmt.Sprintf("Failed to insert into Poster_Sets: %v", err))
			migrationSuccess = false
			continue
		}
		logging.LOG.Info(fmt.Sprintf("Migrated poster set with ID %s for media item %s into new DB", posterSetID, mediaItemID))
	}

	logging.LOG.Info("Migration completed successfully")

	if migrationSuccess {
		// Delete the old database file
		err = os.Remove(oldDBPath)
		if err != nil {
			logging.LOG.Error(fmt.Sprintf("Failed to delete old database file: %v", err))
		} else {
			logging.LOG.Info("Old database file deleted successfully")
		}
	} else {
		logging.LOG.Error("Migration completed with errors, old database file not deleted")
	}
}
