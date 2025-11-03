package api

import (
	"aura/internal/logging"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"
)

type FileIssues struct {
	Errors   []string
	Warnings []string
}

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

	fileIssuesMap := make(map[string]FileIssues)

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

		// Create an array of errors and warning for each file
		fileErrors := []string{}
		fileWarnings := []string{}

		filePath := path.Join(queueFolderPath, file.Name())

		// Read and parse the JSON file
		data, err := os.ReadFile(filePath)
		if err != nil {
			subAction.AppendWarning(fmt.Sprintf("file_%s", file.Name()), "Failed to read file")
			fileErrors = append(fileErrors, fmt.Sprintf("Failed to read file: %s", err.Error()))
			fileIssuesMap[file.Name()] = FileIssues{Errors: fileErrors, Warnings: fileWarnings}
			subAction.Complete()
			go SendDownloadQueueNotification(fileIssuesMap[file.Name()], "", "")
			ld.Log()
			continue
		}

		var queueItem DBMediaItemWithPosterSets
		err = json.Unmarshal(data, &queueItem)
		if err != nil {
			subAction.AppendWarning(fmt.Sprintf("file_%s", file.Name()), "Failed to parse JSON")
			fileErrors = append(fileErrors, fmt.Sprintf("Failed to parse JSON: %s", err.Error()))
			fileIssuesMap[file.Name()] = FileIssues{Errors: fileErrors, Warnings: fileWarnings}
			subAction.Complete()
			go SendDownloadQueueNotification(fileIssuesMap[file.Name()], "", "")
			ld.Log()
			continue
		}

		// Ensure there is at least one poster set
		if len(queueItem.PosterSets) == 0 {
			subAction.AppendWarning(fmt.Sprintf("file_%s", file.Name()), "No poster sets found in queue item")
			fileErrors = append(fileErrors, "No poster sets found in queue item")
			fileIssuesMap[file.Name()] = FileIssues{Errors: fileErrors, Warnings: fileWarnings}
			subAction.Complete()
			go SendDownloadQueueNotification(fileIssuesMap[file.Name()], queueItem.MediaItem.Title, "")
			ld.Log()
			continue
		}

		// Get the first poster set
		posterSet := queueItem.PosterSets[0]

		// Fetch the latest media item data
		latestMediaItem, Err := CallFetchItemContent(ctx, queueItem.MediaItem.RatingKey, queueItem.MediaItem.LibraryTitle)
		if Err.Message != "" {
			subAction.AppendWarning(fmt.Sprintf("file_%s", file.Name()), "Failed to fetch latest media item data")
			fileErrors = append(fileErrors, fmt.Sprintf("Failed to fetch latest data for '%s (%s): %s", latestMediaItem.Title, latestMediaItem.LibraryTitle, Err.Message))
			fileIssuesMap[file.Name()] = FileIssues{Errors: fileErrors, Warnings: fileWarnings}
			subAction.Complete()
			go SendDownloadQueueNotification(fileIssuesMap[file.Name()], latestMediaItem.Title, posterSet.PosterSetID)
			ld.Log()
			continue
		}
		subAction.AppendResult("media_item_title", latestMediaItem.Title)
		subAction.AppendResult("media_item_library", latestMediaItem.LibraryTitle)

		if len(posterSet.SelectedTypes) == 0 {
			subAction.AppendWarning(fmt.Sprintf("posterset_%s", posterSet.PosterSetID), "No selected types for poster set")
			fileWarnings = append(fileWarnings, "No selected types for poster set")
			fileIssuesMap[file.Name()] = FileIssues{Errors: fileErrors, Warnings: fileWarnings}
			subAction.Complete()
			go SendDownloadQueueNotification(fileIssuesMap[file.Name()], latestMediaItem.Title, posterSet.PosterSetID)
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
			case "seasonPoster":
				seasonSelected = true
			case "specialSeasonPoster":
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
				fileWarnings = append(fileWarnings, fmt.Sprintf("Poster Download: %s", Err.Message))
			} else {
				DeleteTempImageForNextLoad(ctx, *posterSet.PosterSet.Poster, latestMediaItem.RatingKey)
			}
		}
		if backdropSelected && posterSet.PosterSet.Backdrop != nil {
			downloadFileName := MediaServer_GetFileDownloadName(*posterSet.PosterSet.Backdrop)
			Err = CallDownloadAndUpdatePosters(ctx, latestMediaItem, *posterSet.PosterSet.Backdrop)
			if Err.Message != "" {
				subAction.AppendWarning(downloadFileName, "Failed to download and update backdrop")
				fileWarnings = append(fileWarnings, fmt.Sprintf("Backdrop Download: %s", Err.Message))
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
						subAction.AppendWarning(downloadFileName, fmt.Sprintf("Failed to download and update season %d poster", seasonPoster.Season.Number))
						fileWarnings = append(fileWarnings, fmt.Sprintf("Season %d Poster Download: %s", seasonPoster.Season.Number, Err.Message))
					} else {
						DeleteTempImageForNextLoad(ctx, seasonPoster, latestMediaItem.RatingKey)
					}
				} else if specialSeasonSelected && seasonPoster.Season.Number == 0 {
					Err = CallDownloadAndUpdatePosters(ctx, latestMediaItem, seasonPoster)
					if Err.Message != "" {
						subAction.AppendWarning(downloadFileName, "Failed to download and update special season poster")
						fileWarnings = append(fileWarnings, fmt.Sprintf("Special Season Poster Download: %s", Err.Message))
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
					subAction.AppendWarning(downloadFileName, fmt.Sprintf("Failed to download and update S%dE%d title card", titleCard.Episode.SeasonNumber, titleCard.Episode.EpisodeNumber))
					fileWarnings = append(fileWarnings, fmt.Sprintf("S%dE%d Title Card Download: %s", titleCard.Episode.SeasonNumber, titleCard.Episode.EpisodeNumber, Err.Message))
				} else {
					DeleteTempImageForNextLoad(ctx, titleCard, latestMediaItem.RatingKey)
				}
			}
		}

		// Now that the files have been processed, add the item to the database
		Err = DB_InsertAllInfoIntoTables(ctx, queueItem)
		if Err.Message != "" {
			subAction.AppendWarning(fmt.Sprintf("file_%s", file.Name()), "Failed to insert item into database")
			fileErrors = append(fileErrors, fmt.Sprintf("Database Insert Failed: %s", Err.Message))
			fileIssuesMap[file.Name()] = FileIssues{Errors: fileErrors, Warnings: fileWarnings}
			subAction.Complete()
			go SendDownloadQueueNotification(fileIssuesMap[file.Name()], latestMediaItem.Title, posterSet.PosterSetID)
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
		var finalErr error
		if len(fileErrors) > 0 {
			newPath := path.Join(queueFolderPath, fmt.Sprintf("error_%s", file.Name()))
			finalErr = os.Rename(filePath, newPath)
		} else if len(fileWarnings) > 0 {
			newPath := path.Join(queueFolderPath, fmt.Sprintf("warning_%s", file.Name()))
			finalErr = os.Rename(filePath, newPath)
		} else {
			finalErr = os.Remove(filePath)
		}
		if finalErr != nil {
			subAction.AppendWarning(fmt.Sprintf("file_%s", file.Name()), "Failed to process file (rename/delete)")
		}

		// Handle any labels and tags asynchronously
		go func() {
			fileIssuesMap[file.Name()] = FileIssues{Errors: fileErrors, Warnings: fileWarnings}
			SendDownloadQueueNotification(fileIssuesMap[file.Name()], latestMediaItem.Title, posterSet.PosterSetID)
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

func SendDownloadQueueNotification(fileIssues FileIssues, itemTitle, posterSetID string) {
	if len(Global_Config.Notifications.Providers) == 0 {
		return
	}

	result := ""
	if len(fileIssues.Errors) > 0 {
		result = "Error"
	} else if len(fileIssues.Warnings) > 0 {
		result = "Warning"
	} else {
		result = "Success"
	}

	notificationTitle := ""
	messageBody := ""

	if itemTitle == "" {
		itemTitle = "Unknown Title"
	}
	if posterSetID == "" {
		posterSetID = "Unknown Set ID"
	}

	switch result {
	case "Success":
		notificationTitle = "Download Queue - Success"
		messageBody = fmt.Sprintf("%s (Set: %s)", itemTitle, posterSetID)
	case "Warning":
		notificationTitle = "Download Queue - Warning"
		messageBody = fmt.Sprintf("%s (Set: %s)\n\nWarnings:\n%s", itemTitle, posterSetID, strings.Join(fileIssues.Warnings, "\n"))
	case "Error":
		notificationTitle = "Download Queue - Error"
		messageBody = fmt.Sprintf("%s (Set: %s)\n\nErrors:\n%s", itemTitle, posterSetID, strings.Join(fileIssues.Errors, "\n"))
		if len(fileIssues.Warnings) > 0 {
			messageBody = fmt.Sprintf("%s\n\nWarnings:\n%s", messageBody, strings.Join(fileIssues.Warnings, "\n"))
		}
	}

	ctx, ld := logging.CreateLoggingContext(context.Background(), "Notification - Send Download Queue Update")
	logAction := ld.AddAction("Sending Download Queue Notification", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)
	defer ld.Log()
	defer logAction.Complete()

	// Send notification using all configured providers
	for _, provider := range Global_Config.Notifications.Providers {
		if provider.Enabled {
			switch provider.Provider {
			case "Discord":
				Notification_SendDiscordMessage(
					ctx,
					provider.Discord,
					messageBody,
					"",
					notificationTitle,
				)
			case "Pushover":
				Notification_SendPushoverMessage(
					ctx,
					provider.Pushover,
					messageBody,
					"",
					notificationTitle,
				)
			case "Gotify":
				Notification_SendGotifyMessage(
					ctx,
					provider.Gotify,
					messageBody,
					"",
					notificationTitle,
				)
			case "Webhook":
				Notification_SendWebhookMessage(
					ctx,
					provider.Webhook,
					messageBody,
					"",
					notificationTitle,
				)
			}
		}
	}
}
