package api

import (
	"aura/internal/logging"
	"database/sql"
	"fmt"
	"os"
	"path"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

var LATEST_DB_VERSION = 1

func DB_Init() bool {
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

	// Check if DB file exists
	_, statErr := os.Stat(dbPath)
	dbNew := os.IsNotExist(statErr)

	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		logging.LOG.Error(fmt.Sprintf("Failed to open database: %v", err))
		return false
	}

	// If the DB file did not exist, we consider it a new database
	if dbNew {
		logging.LOG.Info("Database file not found; creating new database...")
		// Create new tables: VERSION, MediaItems, PosterSets, SavedItems
		createBaseTablesErr := DB_CreateBaseTables()
		if createBaseTablesErr.Message != "" {
			logging.LOG.ErrorWithLog(createBaseTablesErr)
			return false
		}
		// Set version to latest
		err = DB_UpdateVersionTable(db, LATEST_DB_VERSION)
		if err != nil {
			logging.LOG.Error(fmt.Sprintf("Failed to set database version: %v", err))
			return false
		}
		logging.LOG.Info(fmt.Sprintf("New database initialized with version %d.", LATEST_DB_VERSION))
		return true
	}

	// Check if VERSION table exists
	var tableName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='VERSION';").Scan(&tableName)
	if err == sql.ErrNoRows {
		// If the VERSION table does not exist, run migration from 0 to 1
		Err := DB_Migrate_0_to_1(dbPath)
		if Err.Message != "" {
			logging.LOG.ErrorWithLog(Err)
			DB_UpdateVersionTable(db, 0)
			return false
		}
		err = DB_UpdateVersionTable(db, 1)
		if err != nil {
			logging.LOG.Error(fmt.Sprintf("Failed to update database version: %v", err))
			return false
		}
		logging.LOG.Info(fmt.Sprintf("Database migration to version %d completed successfully.", 1))
	} else if err != nil {
		logging.LOG.Error(fmt.Sprintf("Failed to check for VERSION table: %v", err))
		return false
	} else {
		// If the VERSION table exists, check the version and run necessary migrations
		var version int
		err = db.QueryRow("SELECT version FROM VERSION;").Scan(&version)
		if err != nil {
			logging.LOG.Error(fmt.Sprintf("Failed to get database version: %v", err))
			DB_UpdateVersionTable(db, 0)
			return false
		}

		// Run migrations as needed
		for v := version; v < LATEST_DB_VERSION; v++ {
			switch v {
			case 0:
				Err := DB_Migrate_0_to_1(dbPath)
				if Err.Message != "" {
					logging.LOG.ErrorWithLog(Err)
					DB_UpdateVersionTable(db, v)
					return false
				}

			default:
				logging.LOG.Error(fmt.Sprintf("No migration path for database version %d", v))
				return false
			}

			err = DB_UpdateVersionTable(db, v+1)
			if err != nil {
				logging.LOG.Error(fmt.Sprintf("Failed to update database version: %v", err))
				return false
			}
			logging.LOG.Info(fmt.Sprintf("Database migration to version %d completed successfully.", v+1))

		}
	}

	logging.LOG.Info("Successfully initialized the database at: " + dbPath)
	return true
}

func DB_CreateMediaItemsTable() logging.StandardError {
	logging.LOG.Info("Creating MediaItems table...")
	Err := logging.NewStandardError()

	query := `
	CREATE TABLE IF NOT EXISTS MediaItems (
    TMDB_ID TEXT NOT NULL,
    LibraryTitle TEXT NOT NULL,
    RatingKey TEXT,
    Type TEXT,
    Title TEXT,
    Year INTEGER,
    Thumb TEXT,
    ContentRating TEXT,
    Summary TEXT,
    UpdatedAt INTEGER,
    AddedAt INTEGER,
    ReleasedAt INTEGER,
    Guids_JSON JSON,		-- Store Guids as JSON string
    Movie_JSON JSON,        -- Store Movie struct as JSON string
    Series_JSON JSON,       -- Store Series struct as JSON string
    Full_JSON JSON,         -- Store full MediaItem as JSON string
    PRIMARY KEY (TMDB_ID, LibraryTitle)
);`
	_, err := db.Exec(query)
	if err != nil {
		Err.Message = "Failed to create MediaItems table"
		Err.HelpText = "Database migration failed"
		Err.Details = map[string]any{
			"error": err.Error(),
			"query": query,
		}
		return Err
	}

	return Err
}

func DB_CreatePosterSetsTable() logging.StandardError {
	logging.LOG.Info("Creating PosterSets table...")
	Err := logging.NewStandardError()

	query := `
	CREATE TABLE IF NOT EXISTS PosterSets (
	PosterSetID TEXT NOT NULL,
	TMDB_ID TEXT NOT NULL,
	LibraryTitle TEXT NOT NULL,
	PosterSetUser TEXT,
	PosterSet_JSON JSON,
	LastDownloaded DATETIME,
	SelectedTypes TEXT,
	AutoDownload BOOLEAN,
	PRIMARY KEY (PosterSetID, TMDB_ID, LibraryTitle),
	FOREIGN KEY (TMDB_ID, LibraryTitle) REFERENCES MediaItem(TMDB_ID, LibraryTitle)
	);`
	_, err := db.Exec(query)
	if err != nil {
		Err.Message = "Failed to create PosterSets table"
		Err.HelpText = "Database migration failed"
		Err.Details = map[string]any{
			"error": err.Error(),
			"query": query,
		}
		return Err
	}

	return Err
}

func DB_CreateSavedItemsTable() logging.StandardError {
	logging.LOG.Info("Creating SavedItems table...")
	Err := logging.NewStandardError()

	query := `
	CREATE TABLE IF NOT EXISTS SavedItems (
	TMDB_ID TEXT NOT NULL,
	LibraryTitle TEXT NOT NULL,
    PosterSetID TEXT NOT NULL,
    PRIMARY KEY (TMDB_ID, LibraryTitle, PosterSetID),
    FOREIGN KEY (TMDB_ID, LibraryTitle) REFERENCES MediaItem(TMDB_ID, LibraryTitle),
    FOREIGN KEY (PosterSetID) REFERENCES PosterSet(PosterSetID)
);
`
	_, err := db.Exec(query)
	if err != nil {
		Err.Message = "Failed to create new SavedItems table"
		Err.HelpText = "Database migration failed"
		Err.Details = map[string]any{
			"error": err.Error(),
			"query": query,
		}
		return Err
	}

	return Err
}
