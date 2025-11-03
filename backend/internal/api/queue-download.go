package api

import (
	"aura/internal/logging"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
)

func ProcessDownloadQueue() {
	ctx, ld := logging.CreateLoggingContext(context.Background(), "Download Queue Processing")
	logAction := ld.AddAction("Processing Download Queue", logging.LevelInfo)
	defer logAction.Complete()
	ctx = logging.WithCurrentAction(ctx, logAction)

	// Get the download-queue folder path
	queueFolderPath := GetDownloadQueueFolderPath(ctx)
	if queueFolderPath == "" {
		return
	}

	// Read all JSON files in the download-queue folder
	files, err := os.ReadDir(queueFolderPath)
	if err != nil {
		logAction.SetError("Failed to read download queue folder",
			"Ensure that the download-queue folder exists and is accessible",
			map[string]any{"error": err.Error()})
		return
	}

	if len(files) == 0 {
		logAction.AppendResult("queue_empty", "No files in download queue")
		return
	}

	// Process each file
	for _, file := range files {
		if file.IsDir() || path.Ext(file.Name()) != ".json" {
			continue // Skip non-JSON files
		}

		// If file starts with "error_" or "warning_", skip it
		if len(file.Name()) > 6 && (file.Name()[:6] == "error_" || file.Name()[:8] == "warning_") {
			continue
		}

		ctx, ld := logging.CreateLoggingContext(context.Background(), "Download Queue - Processing")
		subAction := ld.AddAction(fmt.Sprintf("Processing file: %s", file.Name()), logging.LevelInfo)
		ctx = logging.WithCurrentAction(ctx, subAction)

		filePath := path.Join(queueFolderPath, file.Name())
		result := "success"

		data, err := os.ReadFile(filePath)
		if err != nil {
			subAction.AppendWarning(fmt.Sprintf("file_%s", file.Name()), "Failed to read file")
			result = "error"
			subAction.Complete()
			ld.Log()
			continue
		}

		var queueItem DBMediaItemWithPosterSets
		err = json.Unmarshal(data, &queueItem)
		if err != nil {
			subAction.AppendWarning(fmt.Sprintf("file_%s", file.Name()), "Failed to parse JSON")
			result = "error"
			subAction.Complete()
			ld.Log()
			continue
		}

		// Fetch the latest media item data
		latestMediaItem, Err := CallFetchItemContent(ctx, queueItem.MediaItem.RatingKey, queueItem.MediaItem.LibraryTitle)
		if Err.Message != "" {
			subAction.AppendWarning(fmt.Sprintf("file_%s", file.Name()), "Failed to fetch latest media item data")
			result = "error"
			subAction.Complete()
			ld.Log()
			continue
		}
		subAction.AppendResult("media_item_title", latestMediaItem.Title)
		subAction.AppendResult("media_item_library", latestMediaItem.LibraryTitle)

		// Process each poster set in the queue item
		for _, posterSet := range queueItem.PosterSets {
			if len(posterSet.SelectedTypes) == 0 {
				subAction.AppendWarning(fmt.Sprintf("posterset_%s", posterSet.PosterSetID), "No selected types for poster set")
				result = "warning"
				subAction.Complete()
				ld.Log()
				continue
			}

			// Check if the selected types contains each image type
			posterSelected := false
			backdropSelected := false
			seasonSelected := false
			specialSeasonSelected := false
			titlecardSelected := false
			for _, selectedType := range posterSet.SelectedTypes {
				switch selectedType {
				case "poster":
					posterSelected = true
				case "backdrop":
					backdropSelected = true
				case "season":
					seasonSelected = true
				case "special_season":
					specialSeasonSelected = true
				case "titlecard":
					titlecardSelected = true
				}
			}

			// Download and update each selected type
			if posterSelected && posterSet.PosterSet.Poster != nil {
				downloadFileName := MediaServer_GetFileDownloadName(*posterSet.PosterSet.Poster)
				Err = CallDownloadAndUpdatePosters(ctx, latestMediaItem, *posterSet.PosterSet.Poster)
				if Err.Message != "" {
					subAction.AppendWarning(downloadFileName, "Failed to download and update poster")
					result = "warning"
				} else {
					DeleteTempImageForNextLoad(ctx, *posterSet.PosterSet.Poster, latestMediaItem.RatingKey)
				}
			}
			if backdropSelected && posterSet.PosterSet.Backdrop != nil {
				downloadFileName := MediaServer_GetFileDownloadName(*posterSet.PosterSet.Backdrop)
				Err = CallDownloadAndUpdatePosters(ctx, latestMediaItem, *posterSet.PosterSet.Backdrop)
				if Err.Message != "" {
					subAction.AppendWarning(downloadFileName, "Failed to download and update backdrop")
					result = "warning"
				} else {
					DeleteTempImageForNextLoad(ctx, *posterSet.PosterSet.Backdrop, latestMediaItem.RatingKey)
				}
			}
			if seasonSelected || specialSeasonSelected {
				for _, seasonPoster := range posterSet.PosterSet.SeasonPosters {
					downloadFileName := MediaServer_GetFileDownloadName(seasonPoster)
					if seasonSelected && seasonPoster.Season.Number > 0 {
						Err = CallDownloadAndUpdatePosters(ctx, latestMediaItem, seasonPoster)
						if Err.Message != "" {
							subAction.AppendWarning(downloadFileName, "Failed to download and update season poster")
							result = "warning"
						} else {
							DeleteTempImageForNextLoad(ctx, seasonPoster, latestMediaItem.RatingKey)
						}
					} else if specialSeasonSelected && seasonPoster.Season.Number == 0 {
						Err = CallDownloadAndUpdatePosters(ctx, latestMediaItem, seasonPoster)
						if Err.Message != "" {
							subAction.AppendWarning(downloadFileName, "Failed to download and update special season poster")
							result = "warning"
						} else {
							DeleteTempImageForNextLoad(ctx, seasonPoster, latestMediaItem.RatingKey)
						}
					}
				}
			}
			if titlecardSelected {
				for _, titleCard := range posterSet.PosterSet.TitleCards {
					downloadFileName := MediaServer_GetFileDownloadName(titleCard)
					Err = CallDownloadAndUpdatePosters(ctx, latestMediaItem, titleCard)
					if Err.Message != "" {
						subAction.AppendWarning(downloadFileName, "Failed to download and update title card")
						result = "warning"
					} else {
						DeleteTempImageForNextLoad(ctx, titleCard, latestMediaItem.RatingKey)
					}
				}
			}
		}

		// Now that the files have been processed, add the item to the database
		Err = DB_InsertAllInfoIntoTables(ctx, queueItem)
		if Err.Message != "" {
			subAction.AppendWarning(fmt.Sprintf("file_%s", file.Name()), "Failed to insert item into database")
			result = "error"
			subAction.Complete()
			ld.Log()
			continue
		}

		// Update the in-memory cache
		inDB, posterSummary, Err := DB_CheckIfMediaItemExists(ctx, queueItem.TMDB_ID, queueItem.LibraryTitle)
		if Err.Message == "" && inDB && len(posterSummary) > 0 {
			addToCacheAction := subAction.AddSubAction("Add Item To Cache", logging.LevelDebug)
			queueItem.MediaItem.DBSavedSets = posterSummary
			queueItem.MediaItem.ExistInDatabase = true
			// Update the in-memory cache
			Global_Cache_LibraryStore.UpdateMediaItem(queueItem.LibraryTitle, &queueItem.MediaItem)
			addToCacheAction.Complete()
		}

		// Post-Processing: Remove or Rename processed files
		switch result {
		case "success":
			err := os.Remove(filePath)
			if err != nil {
				subAction.AppendWarning(fmt.Sprintf("file_%s", file.Name()), "Failed to delete processed file")
			}
		case "warning":
			newPath := path.Join(queueFolderPath, fmt.Sprintf("warning_%s", file.Name()))
			err := os.Rename(filePath, newPath)
			if err != nil {
				subAction.AppendWarning(fmt.Sprintf("file_%s", file.Name()), "Failed to rename warning file")
			}
		case "error":
			newPath := path.Join(queueFolderPath, fmt.Sprintf("error_%s", file.Name()))
			err := os.Rename(filePath, newPath)
			if err != nil {
				subAction.AppendWarning(fmt.Sprintf("file_%s", file.Name()), "Failed to rename error file")
			}
		}

		// Handle any labels and tags asynchronously
		go func() {
			Plex_HandleLabels(latestMediaItem)
			SR_CallHandleTags(context.Background(), latestMediaItem)
		}()

		subAction.Complete()
		ld.Log()
	}

}

func GetDownloadQueueFolderPath(ctx context.Context) string {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Get Download Queue Folder Path", logging.LevelTrace)
	defer logAction.Complete()

	// Use an environment variable to determine the config path
	// By default, it will use /config
	// This is useful for testing and local development
	// In Docker, the config path is set to /config
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "/config"
	}
	queueFolderPath := path.Join(configPath, "download-queue")

	// Ensure the download-queue folder exists
	Err := Util_File_CheckFolderExists(ctx, queueFolderPath)
	if Err.Message != "" {
		return ""
	}

	return queueFolderPath
}
