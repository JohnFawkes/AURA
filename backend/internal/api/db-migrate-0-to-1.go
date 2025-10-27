package api

import (
	"aura/internal/logging"
	"encoding/json"
	"fmt"
	"strings"
)

func DB_Migrate_0_to_1(dbPath string) logging.StandardError {
	logging.LOG.Info("Migrating database from version 0 to version 1...")
	Err := logging.NewStandardError()

	// Create a backup before migration
	backupErr := DB_MakeBackup(dbPath, 0)
	if backupErr.Message != "" {
		return backupErr
	}

	// Rename SavedItems table to SavedItemsBackup
	renameSavedItemsTableErr := DB_0_1_RenameSavedItemsTable()
	if renameSavedItemsTableErr.Message != "" {
		return renameSavedItemsTableErr
	}

	// Create new tables: VERSION, MediaItems, PosterSets, SavedItems
	createBaseTablesErr := DB_CreateBaseTables()
	if createBaseTablesErr.Message != "" {
		return createBaseTablesErr
	}

	// Convert old SavedItems data to new format
	convertOldSavedItemsErr := DB_0_1_ConvertOldSavedItems()
	if convertOldSavedItemsErr.Message != "" {
		return convertOldSavedItemsErr
	}

	// Drop the old SavedItemsBackup table
	dropOldTableQuery := `DROP TABLE IF EXISTS SavedItemsBackup;`
	_, err := db.Exec(dropOldTableQuery)
	if err != nil {
		Err.Message = "Failed to drop old SavedItemsBackup table"
		Err.HelpText = "Database migration failed"
		Err.Details = map[string]any{
			"error": err.Error(),
			"query": dropOldTableQuery,
		}
		return Err
	}

	return Err
}

func DB_CreateBaseTables() logging.StandardError {
	// Create VERSION table
	createVersionTableErr := DB_CreateVersionTable()
	if createVersionTableErr.Message != "" {
		return createVersionTableErr
	}

	// Create MediaItems table
	createMediaItemsTableErr := DB_CreateMediaItemsTable()
	if createMediaItemsTableErr.Message != "" {
		return createMediaItemsTableErr
	}

	// Create PosterSets table
	createPosterSetsTableErr := DB_CreatePosterSetsTable()
	if createPosterSetsTableErr.Message != "" {
		return createPosterSetsTableErr
	}

	// Create new SavedItems table
	createSavedItemsTableErr := DB_CreateSavedItemsTable()
	if createSavedItemsTableErr.Message != "" {
		return createSavedItemsTableErr
	}
	return logging.NewStandardError()
}

func DB_CreateVersionTable() logging.StandardError {
	logging.LOG.Info("Creating VERSION table...")
	Err := logging.NewStandardError()

	query := `
	CREATE TABLE IF NOT EXISTS VERSION (
		version INTEGER NOT NULL
	);
	`
	_, err := db.Exec(query)
	if err != nil {
		Err.Message = "Failed to create VERSION table"
		Err.HelpText = "Database migration failed"
		Err.Details = map[string]any{
			"error": err.Error(),
			"query": query,
		}
		return Err
	}

	return Err
}

func DB_0_1_RenameSavedItemsTable() logging.StandardError {
	logging.LOG.Info("Migration (0-1): Renaming SavedItems table...")
	Err := logging.NewStandardError()

	query := `
	ALTER TABLE SavedItems RENAME TO SavedItemsBackup;
	`
	_, err := db.Exec(query)
	if err != nil {
		Err.Message = "Failed to rename SavedItems table"
		Err.HelpText = "Database migration failed"
		Err.Details = map[string]any{
			"error": err.Error(),
			"query": query,
		}
		return Err
	}

	return Err
}

func DB_0_1_ConvertOldSavedItems() logging.StandardError {
	logging.LOG.Info("Migration (0-1): Converting old saved items into new format...")
	Err := logging.NewStandardError()

	// Fetch all entries from the old SavedItemsBackup table
	fetchAllQuery := `SELECT media_item_id, media_item, poster_set_id, poster_set, selected_types, auto_download, last_update FROM SavedItemsBackup;`
	rows, err := db.Query(fetchAllQuery)
	if err != nil {
		Err.Message = "Failed to fetch data from old SavedItemsBackup table"
		Err.HelpText = "Database migration failed"
		Err.Details = map[string]any{
			"error": err.Error(),
			"query": fetchAllQuery,
		}
		return Err
	}

	// Read all rows into a slice
	type OldSavedItem struct {
		MediaItemID   string
		MediaItemJSON string
		PosterSetID   string
		PosterSetJSON string
		SelectedTypes string
		AutoDownload  bool
		LastUpdate    string
	}
	var items []OldSavedItem
	results := map[string]int{}

	results["SavedItemFails"] = 0
	results["SavedItemSuccess"] = 0

	WarnErr := logging.NewStandardError()

	for rows.Next() {
		var item OldSavedItem
		err := rows.Scan(&item.MediaItemID, &item.MediaItemJSON, &item.PosterSetID, &item.PosterSetJSON, &item.SelectedTypes, &item.AutoDownload, &item.LastUpdate)
		if err != nil {
			WarnErr.Message = "Failed to scan row from old SavedItemsBackup table"
			WarnErr.Details = map[string]any{
				"error": err.Error(),
			}
			DB_Migration_CreateWarningFile(1, WarnErr)
			results["SavedItemFails"]++
			continue
		}
		items = append(items, item)
	}
	rows.Close()

	for _, item := range items {
		logging.LOG.Debug(fmt.Sprintf("Migration (0-1): Processing SavedItem with MediaItemID: %s and PosterSetID: %s", item.MediaItemID, item.PosterSetID))
		// Get MediaItem struct from JSON
		mediaItem, mediaItemErr := DB_0_1_MediaItemJSONToMediaItem(item.MediaItemJSON)
		if mediaItemErr.Message != "" {
			results["SavedItemFails"]++
			continue
		}

		// Get PosterSet struct from JSON
		posterSet, posterSetErr := DB_0_1_PosterSetJSONToPosterSet(item.PosterSetJSON)
		if posterSetErr.Message != "" {
			results["SavedItemFails"]++
			continue
		}

		// If the Media Item Type is empty, try to get it from the PosterSet
		if mediaItem.Type == "" {
			if posterSet.Poster != nil {
				if posterSet.Poster.Movie != nil {
					mediaItem.Type = "movie"
				} else if posterSet.Poster.Show != nil {
					mediaItem.Type = "show"
				}
			}
			if mediaItem.Type == "" && posterSet.Backdrop != nil {
				if posterSet.Backdrop.Movie != nil {
					mediaItem.Type = "movie"
				} else if posterSet.Backdrop.Show != nil {
					mediaItem.Type = "show"
				}
			}
			if mediaItem.Type == "" && len(posterSet.SeasonPosters) > 0 {
				for _, seasonPoster := range posterSet.SeasonPosters {
					if seasonPoster.Show != nil {
						mediaItem.Type = "show"
						break
					}
				}
			}
			if mediaItem.Type == "" && len(posterSet.TitleCards) > 0 {
				for _, titleCard := range posterSet.TitleCards {
					if titleCard.Show != nil {
						mediaItem.Type = "show"
						break
					}
				}
			}
		}

		// If we still don't have a valid type, skip this item
		if mediaItem.Type != "movie" && mediaItem.Type != "show" {
			WarnErr.Message = "MediaItem has invalid Type"
			WarnErr.HelpText = fmt.Sprintf("Type must be 'movie' or 'show', got '%s'", mediaItem.Type)
			WarnErr.Details = map[string]any{
				"MediaItem": mediaItem,
				"PosterSet": posterSet,
			}
			DB_Migration_CreateWarningFile(1, WarnErr)
			results["SavedItemFails"]++
			continue
		}

		partialFound := 0
		if mediaItem.Type == "movie" && (mediaItem.Movie == nil || mediaItem.Movie.File.Path == "") {
			if posterSet.Poster != nil && posterSet.Poster.Movie != nil {
				movie := posterSet.Poster.Movie
				if mediaItem.LibraryTitle == "" && movie.MediaItem.LibraryTitle != "" {
					mediaItem.LibraryTitle = movie.MediaItem.LibraryTitle
					partialFound = 1
				}
				if mediaItem.TMDB_ID == "" && movie.ID != "" {
					mediaItem.TMDB_ID = movie.ID
					partialFound = 1
				}
			}
			if posterSet.Backdrop != nil && posterSet.Backdrop.Movie != nil {
				movie := posterSet.Backdrop.Movie
				if mediaItem.LibraryTitle == "" && movie.MediaItem.LibraryTitle != "" {
					mediaItem.LibraryTitle = movie.MediaItem.LibraryTitle
					partialFound = 1
				}
				if mediaItem.TMDB_ID == "" && movie.ID != "" {
					mediaItem.TMDB_ID = movie.ID
					partialFound = 1
				}
			}
		} else if mediaItem.Type == "show" && (mediaItem.Series == nil) {
			// Try to get missing info from PosterSet's Poster or Backdrop
			if posterSet.Poster != nil && posterSet.Poster.Show != nil {
				show := posterSet.Poster.Show
				if mediaItem.LibraryTitle == "" && show.MediaItem.LibraryTitle != "" {
					mediaItem.LibraryTitle = show.MediaItem.LibraryTitle
					partialFound = 1
				}
				if mediaItem.TMDB_ID == "" && show.ID != "" {
					mediaItem.TMDB_ID = show.ID
					partialFound = 1
				}
			}
			if posterSet.Backdrop != nil && posterSet.Backdrop.Show != nil {
				show := posterSet.Backdrop.Show
				if mediaItem.LibraryTitle == "" && show.MediaItem.LibraryTitle != "" {
					mediaItem.LibraryTitle = show.MediaItem.LibraryTitle
					partialFound = 1
				}
				if mediaItem.TMDB_ID == "" && show.ID != "" {
					mediaItem.TMDB_ID = show.ID
					partialFound = 1
				}
			}
			if mediaItem.TMDB_ID == "" || mediaItem.LibraryTitle == "" {
				for _, seasonPoster := range posterSet.SeasonPosters {
					if seasonPoster.Show != nil {
						if mediaItem.LibraryTitle == "" && seasonPoster.Show.MediaItem.LibraryTitle != "" {
							mediaItem.LibraryTitle = seasonPoster.Show.MediaItem.LibraryTitle
							partialFound = 1
						}
						if mediaItem.TMDB_ID == "" && seasonPoster.Show.ID != "" {
							mediaItem.TMDB_ID = seasonPoster.Show.ID
							partialFound = 1
						}
					}
				}
			}
			if mediaItem.TMDB_ID == "" || mediaItem.LibraryTitle == "" {
				for _, titleCard := range posterSet.TitleCards {
					if titleCard.Show != nil {
						if mediaItem.LibraryTitle == "" && titleCard.Show.MediaItem.LibraryTitle != "" {
							mediaItem.LibraryTitle = titleCard.Show.MediaItem.LibraryTitle
							partialFound = 1
						}
						if mediaItem.TMDB_ID == "" && titleCard.Show.ID != "" {
							mediaItem.TMDB_ID = titleCard.Show.ID
							partialFound = 1
						}
					}
				}
			}
		}

		// Check we have the minimum required information
		if mediaItem.TMDB_ID == "" || mediaItem.LibraryTitle == "" || partialFound == 1 {
			// Try and get the media item from the cache
			mediaItemFromCache, exist := Global_Cache_LibraryStore.GetMediaItemFromSectionByTitleAndYear(mediaItem.LibraryTitle, mediaItem.Title, mediaItem.Year)
			if exist {
				mediaItem = *mediaItemFromCache
			} else {
				WarnErr.Message = "MediaItem missing TMDB_ID or LibraryTitle"
				WarnErr.HelpText = fmt.Sprintf("Title: '%s', TMDB_ID: '%s', LibraryTitle: '%s', Year: '%d'", mediaItem.Title, mediaItem.TMDB_ID, mediaItem.LibraryTitle, mediaItem.Year)
				WarnErr.Details = map[string]any{
					"MediaItem": mediaItem,
					"PosterSet": posterSet,
				}
				DB_Migration_CreateWarningFile(1, WarnErr)
				results["SavedItemFails"]++
				continue
			}
		}

		// Create a DBMediaItemWithPosterSets struct
		dbItem := DBMediaItemWithPosterSets{
			TMDB_ID:      mediaItem.TMDB_ID,
			LibraryTitle: mediaItem.LibraryTitle,
			MediaItem:    mediaItem,
			PosterSets: []DBPosterSetDetail{
				{
					PosterSetID:    posterSet.ID,
					PosterSet:      posterSet,
					LastDownloaded: item.LastUpdate,
					SelectedTypes:  strings.Split(item.SelectedTypes, ","), // Convert string to []string
					AutoDownload:   item.AutoDownload,
				},
			},
		}

		// Insert all info into the new tables
		insertErr := DB_InsertAllInfoIntoTables(dbItem)
		if insertErr.Message != "" {
			results["SavedItemFails"]++
			logging.LOG.Error(fmt.Sprintf("Migration (0-1): Failed to insert SavedItem with MediaItemID: %s and PosterSetID: %s. Error: %s\n%s", item.MediaItemID, item.PosterSetID, insertErr.Message, insertErr.Details))
			continue
		}

		results["SavedItemSuccess"]++

	}

	logging.LOG.Info(fmt.Sprintf("Migration (0-1) Results: Success=%d, Fails=%d, Total=%d", results["SavedItemSuccess"], results["SavedItemFails"], results["SavedItemSuccess"]+results["SavedItemFails"]))

	return Err
}

func DB_0_1_MediaItemJSONToMediaItem(mediaItemJSON string) (MediaItem, logging.StandardError) {
	WarnErr := logging.NewStandardError()
	// Unmarshal media item JSON
	var mediaItem MediaItem
	err := json.Unmarshal([]byte(mediaItemJSON), &mediaItem)
	if err != nil {
		WarnErr.Message = "Failed to unmarshal MediaItem JSON"
		WarnErr.Details = map[string]any{
			"error": err.Error(),
			"json":  mediaItemJSON,
		}
		DB_Migration_CreateWarningFile(1, WarnErr)
		return mediaItem, WarnErr
	}

	// Get the TMDB ID from the Guids
	tmdbID := ""
	for _, guid := range mediaItem.Guids {
		if guid.Provider == "tmdb" {
			tmdbID = guid.ID
			break
		}
	}
	mediaItem.TMDB_ID = tmdbID

	if mediaItem.Type != "movie" && mediaItem.Type != "show" {
		mediaItem.Type = ""
	}

	return mediaItem, WarnErr
}

func DB_0_1_PosterSetJSONToPosterSet(posterSetJSON string) (PosterSet, logging.StandardError) {
	WarnErr := logging.NewStandardError()
	// Unmarshal poster set JSON
	var posterSet PosterSet
	err := json.Unmarshal([]byte(posterSetJSON), &posterSet)
	if err != nil {
		WarnErr.Message = "Failed to unmarshal PosterSet JSON"
		WarnErr.Details = map[string]any{
			"error": err.Error(),
			"json":  posterSetJSON,
		}
		DB_Migration_CreateWarningFile(1, WarnErr)
		return posterSet, WarnErr
	}

	// Check PosterSet ID
	if posterSet.ID == "" {
		WarnErr.Message = "PosterSet missing ID"
		WarnErr.Details = map[string]any{
			"PosterSet": posterSet,
		}
		DB_Migration_CreateWarningFile(1, WarnErr)
		return posterSet, WarnErr
	}

	return posterSet, WarnErr
}
