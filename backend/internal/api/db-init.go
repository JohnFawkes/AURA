package api

import (
	"aura/internal/logging"
	"context"
	"database/sql"
	"fmt"
	"os"
	"path"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

var LATEST_DB_VERSION = 1

func DB_Init() bool {
	ctx, ld := logging.CreateLoggingContext(context.Background(), "Database - Initialization")
	defer ld.Log()
	logAction := ld.AddAction("Initializing Database", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)
	defer logAction.Complete()

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

	openAction := logAction.AddSubAction("Opening Database", logging.LevelDebug)
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		openAction.SetError("Failed to open database", "", map[string]any{
			"dbPath": dbPath,
			"error":  err.Error(),
		})
		return false
	}
	openAction.Complete()

	// If the DB file did not exist, we consider it a new database
	if dbNew {
		logging.LOGGER.Info().Timestamp().Msg("Database file not found, creating new database")
		// Create new tables: VERSION, MediaItems, PosterSets, SavedItems
		createBaseTablesErr := DB_CreateBaseTables(ctx)
		if createBaseTablesErr.Message != "" {
			return false
		}
		// Set version to latest
		updateVersionTableErr := DB_UpdateVersionTable(ctx, db, LATEST_DB_VERSION)
		if updateVersionTableErr.Message != "" {
			return false
		}
		return true
	}

	// Check if VERSION table exists
	var tableName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='VERSION';").Scan(&tableName)
	if err == sql.ErrNoRows {
		// If the VERSION table does not exist, run migration from 0 to 1
		Err := DB_Migrate_0_to_1(ctx, dbPath)
		if Err.Message != "" {
			return false
		}
		Err = DB_UpdateVersionTable(ctx, db, 1)
		if Err.Message != "" {
			return false
		}
	} else if err != nil {
		logging.LOGGER.Error().Err(err).Msg("Failed to check for VERSION table")
		return false
	} else {
		// If the VERSION table exists, check the version and run necessary migrations
		var version int
		err = db.QueryRow("SELECT version FROM VERSION;").Scan(&version)
		if err != nil {
			logging.LOGGER.Error().Err(err).Msg("Failed to get database version")
			DB_UpdateVersionTable(ctx, db, 0)
			return false
		}

		// Run migrations as needed
		for v := version; v < LATEST_DB_VERSION; v++ {
			switch v {
			case 0:
				Err := DB_Migrate_0_to_1(ctx, dbPath)
				if Err.Message != "" {
					DB_UpdateVersionTable(ctx, db, v)
					return false
				}

			default:
				logging.LOGGER.Error().Msg(fmt.Sprintf("No migration path for database version %d", v))
				return false
			}

			Err := DB_UpdateVersionTable(ctx, db, v+1)
			if Err.Message != "" {
				return false
			}

		}
	}

	// Check if we can vacuum the database
	vacuumErr := DB_Vacuum(ctx)
	if vacuumErr.Message != "" {
		return false
	}

	return true
}

func DB_Vacuum(ctx context.Context) logging.LogErrorInfo {
	ctx, logAction := logging.AddSubActionToContext(ctx, "DB: VACUUM Database", logging.LevelInfo)
	defer logAction.Complete()

	// Check if freelist_count is greater than 10000
	var freeListCount int64
	err := db.QueryRow("PRAGMA freelist_count;").Scan(&freeListCount)
	if err != nil {
		logAction.SetError("Failed to get freelist_count", err.Error(), map[string]any{
			"error": err.Error(),
		})
		return *logAction.Error
	}

	if freeListCount < 10000 {
		logAction.AppendResult("freelist_count", freeListCount)
		logAction.AppendResult("vacuum_performed", false)
		return logging.LogErrorInfo{}
	}

	// Perform VACUUM
	_, err = db.Exec("VACUUM;")
	if err != nil {
		logAction.SetError("Failed to VACUUM database", err.Error(), map[string]any{
			"error": err.Error(),
		})
		return *logAction.Error
	}

	logAction.AppendResult("freelist_count", freeListCount)
	logAction.AppendResult("vacuum_performed", true)
	return logging.LogErrorInfo{}
}

func DB_CreateBaseTables(ctx context.Context) logging.LogErrorInfo {
	// Create VERSION table
	createVersionTableErr := DB_CreateVersionTable(ctx)
	if createVersionTableErr.Message != "" {
		return createVersionTableErr
	}

	// Create MediaItems table
	createMediaItemsTableErr := DB_CreateMediaItemsTable(ctx)
	if createMediaItemsTableErr.Message != "" {
		return createMediaItemsTableErr
	}

	// Create PosterSets table
	createPosterSetsTableErr := DB_CreatePosterSetsTable(ctx)
	if createPosterSetsTableErr.Message != "" {
		return createPosterSetsTableErr
	}

	// Create new SavedItems table
	createNewSavedItemsTableErr := DB_CreateSavedItemsTable(ctx)
	if createNewSavedItemsTableErr.Message != "" {
		return createNewSavedItemsTableErr
	}
	return logging.LogErrorInfo{}
}

func DB_CreateVersionTable(ctx context.Context) logging.LogErrorInfo {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Creating VERSION Table", logging.LevelDebug)
	defer logAction.Complete()

	query := `
	CREATE TABLE IF NOT EXISTS VERSION (
		version INTEGER NOT NULL
	);
	`
	_, err := db.Exec(query)
	if err != nil {
		logAction.SetError("Failed to create VERSION table", err.Error(), map[string]any{
			"error": err.Error(),
			"query": query,
		})
		return *logAction.Error
	}

	return logging.LogErrorInfo{}
}

func DB_CreateMediaItemsTable(ctx context.Context) logging.LogErrorInfo {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Creating MediaItems Table", logging.LevelDebug)
	defer logAction.Complete()

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
		logAction.SetError("Failed to create MediaItems table", err.Error(), map[string]any{
			"error": err.Error(),
			"query": query,
		})
		return *logAction.Error
	}

	return logging.LogErrorInfo{}
}

func DB_CreatePosterSetsTable(ctx context.Context) logging.LogErrorInfo {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Creating PosterSets Table", logging.LevelDebug)
	defer logAction.Complete()

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
		logAction.SetError("Failed to create PosterSets table", err.Error(), map[string]any{
			"error": err.Error(),
			"query": query,
		})
		return *logAction.Error
	}

	return logging.LogErrorInfo{}
}

func DB_CreateSavedItemsTable(ctx context.Context) logging.LogErrorInfo {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Creating SavedItems Table", logging.LevelDebug)
	defer logAction.Complete()

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
		logAction.SetError("Failed to create SavedItems table", err.Error(), map[string]any{
			"error": err.Error(),
			"query": query,
		})
		return *logAction.Error
	}

	return logging.LogErrorInfo{}
}
