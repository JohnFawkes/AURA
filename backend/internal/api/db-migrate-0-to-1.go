package api

import (
	"aura/internal/logging"
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

func DB_Migrate_0_to_1(ctx context.Context, dbPath string) logging.LogErrorInfo {
	logging.LOGGER.Info().Timestamp().Int("From Version", 0).Int("To Version", 1).Msg("Starting database migration")

	// Create a backup before migration
	backupErr := DB_MakeBackup(ctx, dbPath, 0)
	if backupErr.Message != "" {
		return backupErr
	}

	// Rename SavedItems table to SavedItemsBackup
	renameSavedItemsTableErr := DB_0_1_RenameSavedItemsTable(ctx)
	if renameSavedItemsTableErr.Message != "" {
		return renameSavedItemsTableErr
	}

	// Create new tables: VERSION, MediaItems, PosterSets, SavedItems
	createBaseTablesErr := DB_CreateBaseTables(ctx)
	if createBaseTablesErr.Message != "" {
		return createBaseTablesErr
	}

	// Convert old SavedItems data to new format
	convertOldSavedItemsErr := DB_0_1_ConvertOldSavedItems(ctx)
	if convertOldSavedItemsErr.Message != "" {
		return convertOldSavedItemsErr
	}

	// Drop the old SavedItemsBackup table
	dropOldTableQuery := `DROP TABLE IF EXISTS SavedItemsBackup;`
	_, err := db.Exec(dropOldTableQuery)
	if err != nil {
		return logging.LogErrorInfo{
			Message: "Failed to drop old SavedItemsBackup table after migration",
			Detail: map[string]any{
				"error": err.Error(),
			},
		}
	}

	return logging.LogErrorInfo{}
}

func DB_0_1_RenameSavedItemsTable(ctx context.Context) logging.LogErrorInfo {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Renaming SavedItems Table", logging.LevelDebug)
	defer logAction.Complete()

	query := `
	ALTER TABLE SavedItems RENAME TO SavedItemsBackup;
	`
	_, err := db.Exec(query)
	if err != nil {
		logAction.SetError("Failed to rename SavedItems table", "", map[string]any{
			"error": err.Error(),
		})
		return *logAction.Error
	}

	return logging.LogErrorInfo{}
}

func DB_0_1_ConvertOldSavedItems(ctx context.Context) logging.LogErrorInfo {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Converting Old SavedItems Data", logging.LevelInfo)
	defer logAction.Complete()

	// Fetch all entries from the old SavedItemsBackup table
	fetchAllQuery := `SELECT media_item_id, media_item, poster_set_id, poster_set, selected_types, auto_download, last_update FROM SavedItemsBackup;`
	rows, err := db.Query(fetchAllQuery)
	if err != nil {
		logAction.SetError("Failed to fetch data from SavedItemsBackup", "", map[string]any{
			"error": err.Error(),
		})
		return *logAction.Error
	}
	defer rows.Close()

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

	for rows.Next() {
		var item OldSavedItem
		err := rows.Scan(&item.MediaItemID, &item.MediaItemJSON, &item.PosterSetID, &item.PosterSetJSON, &item.SelectedTypes, &item.AutoDownload, &item.LastUpdate)
		if err != nil {
			results["SavedItemFails"]++
			DB_Migration_CreateWarningFile(1, logging.LogErrorInfo{
				Message: "Failed to scan row from old SavedItemsBackup table",
				Detail:  map[string]any{"error": err.Error()},
			})
			continue
		}
		items = append(items, item)
	}
	rows.Close()

	for _, item := range items {
		logging.LOGGER.Debug().Timestamp().Str("Media Item ID", item.MediaItemID).Str("Poster Set ID", item.PosterSetID).Msg("Migration (0-1): Processing Saved Item from DB")

		// Get MediaItem struct from JSON
		mediaItem, mediaItemPassed := DB_0_1_MediaItemJSONToMediaItem(item.MediaItemJSON)
		if !mediaItemPassed {
			logging.LOGGER.Warn().Timestamp().Str("Media Item ID", item.MediaItemID).Msg("Migration (0-1): Skipping Saved Item due to invalid MediaItem JSON")
			results["SavedItemFails"]++
			continue
		}

		// Get PosterSet struct from JSON
		posterSet, posterSetPassed := DB_0_1_PosterSetJSONToPosterSet(item.PosterSetJSON)
		if !posterSetPassed {
			logging.LOGGER.Warn().Timestamp().Str("Poster Set ID", item.PosterSetID).Msg("Migration (0-1): Skipping Saved Item due to invalid PosterSet JSON")
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
			results["SavedItemFails"]++
			DB_Migration_CreateWarningFile(1, logging.LogErrorInfo{
				Message: "MediaItem type is invalid during migration",
				Help:    fmt.Sprintf("Type must be 'movie' or 'show', got '%s'", mediaItem.Type),
				Detail: map[string]any{
					"MediaItem": mediaItem,
					"PosterSet": posterSet,
				},
			})
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
			logging.LOGGER.Debug().Timestamp().Str("Media Item ID", item.MediaItemID).Str("TMDB ID", mediaItem.TMDB_ID).Msg("Migration (0-1): Trying to get missing Media Item info from cache")
			// Try and get the media item from the cache
			mediaItemFromCache, exist := Global_Cache_LibraryStore.GetMediaItemFromSectionByTitleAndYear(mediaItem.LibraryTitle, mediaItem.Title, mediaItem.Year)
			if exist {
				mediaItem = *mediaItemFromCache
			} else {
				results["SavedItemFails"]++
				DB_Migration_CreateWarningFile(1, logging.LogErrorInfo{
					Message: "Insufficient MediaItem data during migration and not found in cache",
					Help:    fmt.Sprintf("Title: '%s', TMDB_ID: '%s', LibraryTitle: '%s', Year: '%d'", mediaItem.Title, mediaItem.TMDB_ID, mediaItem.LibraryTitle, mediaItem.Year),
					Detail: map[string]any{
						"MediaItem": mediaItem,
						"PosterSet": posterSet,
					},
				})
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
		insertErr := DB_InsertAllInfoIntoTables(ctx, dbItem)
		if insertErr.Message != "" {
			results["SavedItemFails"]++
			DB_Migration_CreateWarningFile(1, logging.LogErrorInfo{
				Message: "Failed to insert migrated SavedItem into new tables",
				Detail:  insertErr.Detail,
			})
			continue
		}

		results["SavedItemSuccess"]++
	}

	// Convert results to map[string]any for assignment
	resultAny := make(map[string]any, len(results))
	for k, v := range results {
		resultAny[k] = v
	}
	logAction.Result = resultAny
	return logging.LogErrorInfo{}
}

func DB_0_1_MediaItemJSONToMediaItem(mediaItemJSON string) (MediaItem, bool) {
	// Unmarshal media item JSON
	var mediaItem MediaItem
	err := json.Unmarshal([]byte(mediaItemJSON), &mediaItem)
	if err != nil {
		DB_Migration_CreateWarningFile(1, logging.LogErrorInfo{
			Message: "Failed to unmarshal MediaItem JSON during migration",
			Detail: map[string]any{
				"error": err.Error(),
				"json":  mediaItemJSON,
			},
		})
		return MediaItem{}, false
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
	return mediaItem, true
}

func DB_0_1_PosterSetJSONToPosterSet(posterSetJSON string) (PosterSet, bool) {
	// Unmarshal poster set JSON
	var posterSet PosterSet
	err := json.Unmarshal([]byte(posterSetJSON), &posterSet)
	if err != nil {
		DB_Migration_CreateWarningFile(1, logging.LogErrorInfo{
			Message: "Failed to unmarshal PosterSet JSON during migration",
			Detail: map[string]any{
				"error": err.Error(),
				"json":  posterSetJSON,
			},
		})
		return posterSet, false
	}

	// Check PosterSet ID
	if posterSet.ID == "" {
		DB_Migration_CreateWarningFile(1, logging.LogErrorInfo{
			Message: "PosterSet ID is empty during migration",
			Detail:  map[string]any{"posterSetJSON": posterSetJSON},
		})
		return posterSet, false
	}

	return posterSet, true
}
