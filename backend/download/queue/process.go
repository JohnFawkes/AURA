package downloadqueue

import (
	"aura/database"
	"aura/logging"
	"aura/mediaserver"
	"aura/mediux"
	"aura/models"
	sonarr_radarr "aura/sonarr-radarr"
	"aura/utils"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
)

func ProcessQueueItems() {
	ctx, ld := logging.CreateLoggingContext(context.Background(), "Download Queue Processing")
	logAction := ld.AddAction("Processing Download Queue", logging.LevelInfo)
	defer logAction.Complete()
	ctx = logging.WithCurrentAction(ctx, logAction)

	// Read all JSON files in the download-queue directory
	files, err := os.ReadDir(FolderPath)
	if err != nil {
		logging.LOGGER.Warn().Timestamp().Err(err).Msg("Failed to read download queue directory")
		logAction.SetError("Failed to read download queue directory", "Ensure the directory exists and is accessible",
			map[string]any{
				"error":      err.Error(),
				"folderPath": FolderPath,
			})
		return
	}

	if len(files) == 0 {
		logAction.AppendResult("result", "queue is empty")
		return
	}

	fileIssuesMap := make(map[string]FileIssues)

	// Process each file in the directory
	for _, file := range files {
		if file.IsDir() || path.Ext(file.Name()) != ".json" {
			continue
		}

		// If a file starts with "error_" or "warning_", skip it
		if len(file.Name()) > 6 && (file.Name()[:6] == "error_" || file.Name()[:8] == "warning_") {
			continue
		}

		ctx, ld := logging.CreateLoggingContext(context.Background(), "Download Queue - Processing")
		subAction := ld.AddAction(fmt.Sprintf("Processing file: %s", file.Name()), logging.LevelInfo)
		ctx = logging.WithCurrentAction(ctx, subAction)

		// Reset the Latest Info for this file
		LatestInfo.Status = LAST_STATUS_PROCESSING
		LatestInfo.Message = fmt.Sprintf("Processing file: %s", file.Name())
		LatestInfo.Errors = []string{}
		LatestInfo.Warnings = []string{}

		// Create an array of errors and warnings for this file
		fileErrors := []string{}
		fileWarnings := []string{}

		filePath := path.Join(FolderPath, file.Name())

		// Read and parse the JSON file
		data, err := os.ReadFile(filePath)
		if err != nil {
			subAction.AppendWarning(fmt.Sprintf("file_%s", file.Name()), "Failed to read file")
			fileErrors = append(fileErrors, fmt.Sprintf("Failed to read file: %s", err.Error()))
			fileIssuesMap[file.Name()] = FileIssues{Errors: fileErrors, Warnings: fileWarnings}
			go SendNotification(fileIssuesMap[file.Name()], "Unknown Title", models.DBPosterSetDetail{}, "", "")
			ld.Log()
			continue
		}

		var queueItem models.DBSavedItem
		err = json.Unmarshal(data, &queueItem)
		if err != nil {
			subAction.AppendWarning(fmt.Sprintf("file_%s", file.Name()), "Failed to parse JSON")
			fileErrors = append(fileErrors, fmt.Sprintf("Failed to parse JSON: %s", err.Error()))
			fileIssuesMap[file.Name()] = FileIssues{Errors: fileErrors, Warnings: fileWarnings}
			go SendNotification(fileIssuesMap[file.Name()], "Unknown Title", models.DBPosterSetDetail{}, "", "")
			ld.Log()
			continue
		}

		// Ensure the Media Item has the basic required info
		if queueItem.MediaItem.RatingKey == "" || queueItem.MediaItem.Title == "" || queueItem.MediaItem.LibraryTitle == "" || queueItem.MediaItem.TMDB_ID == "" {
			subAction.AppendWarning(fmt.Sprintf("file_%s", file.Name()), "Media item is missing required information (ID, Title, LibraryTitle, or TMDB_ID)")
			fileErrors = append(fileErrors, "Media item is missing required information (ID, Title, LibraryTitle, or TMDB_ID)")
			fileIssuesMap[file.Name()] = FileIssues{Errors: fileErrors, Warnings: fileWarnings}
			go SendNotification(fileIssuesMap[file.Name()], "Unknown Title", models.DBPosterSetDetail{}, "", "")
			ld.Log()
			continue
		}

		// Ensure there is at least one Poster Set in the item
		if len(queueItem.PosterSets) == 0 {
			subAction.AppendWarning(fmt.Sprintf("file_%s", file.Name()), "No poster sets found in the item")
			fileWarnings = append(fileWarnings, "No poster sets found in the item")
			fileIssuesMap[file.Name()] = FileIssues{Errors: fileErrors, Warnings: fileWarnings}
			go SendNotification(fileIssuesMap[file.Name()], queueItem.MediaItem.Title, models.DBPosterSetDetail{}, "", "")
			ld.Log()
			continue
		}

		// Fetch the Mediux Info for this Item so that we can get the TMDB poster/backdrop paths for the notification
		mediuxItemInfo, Err := mediux.GetBaseItemInfoByTMDB_ID(queueItem.MediaItem.TMDB_ID, queueItem.MediaItem.Type)
		if Err.Message != "" {
			subAction.AppendWarning(fmt.Sprintf("file_%s", file.Name()), "Failed to fetch MediUX info for the item")
			fileWarnings = append(fileWarnings, fmt.Sprintf("Failed to fetch MediUX info for the item: %s", Err.Message))
		}

		// Fetch the latest Media Item info
		found, Err := mediaserver.GetMediaItemDetails(ctx, &queueItem.MediaItem)
		if Err.Message != "" || !found {
			subAction.AppendWarning(fmt.Sprintf("file_%s", file.Name()), "Failed to fetch media item details")
			fileErrors = append(fileErrors, fmt.Sprintf("Failed to fetch latest data for '%s (%s): %s", queueItem.MediaItem.Title, queueItem.MediaItem.LibraryTitle, Err.Message))
			fileIssuesMap[file.Name()] = FileIssues{Errors: fileErrors, Warnings: fileWarnings}
			go SendNotification(fileIssuesMap[file.Name()], queueItem.MediaItem.Title, models.DBPosterSetDetail{}, mediuxItemInfo.TMDB_PosterPath, mediuxItemInfo.TMDB_BackdropPath)
			ld.Log()
			continue
		}

		for _, posterSet := range queueItem.PosterSets {
			// Ensure that the poster set has the required info
			if posterSet.ID == "" || posterSet.Type == "" || posterSet.Title == "" {
				subAction.AppendWarning(fmt.Sprintf("file_%s", file.Name()), fmt.Sprintf("Poster set '%s' is missing required information (ID, Type, or Title)", posterSet.Title))
				fileErrors = append(fileErrors, fmt.Sprintf("Poster set '%s' is missing required information (ID, Type, or Title)", posterSet.Title))
				fileIssuesMap[file.Name()] = FileIssues{Errors: fileErrors, Warnings: fileWarnings}
				go SendNotification(fileIssuesMap[file.Name()], queueItem.MediaItem.Title, posterSet, mediuxItemInfo.TMDB_PosterPath, mediuxItemInfo.TMDB_BackdropPath)
				continue
			}

			if !posterSet.SelectedTypes.Poster &&
				!posterSet.SelectedTypes.Backdrop &&
				!posterSet.SelectedTypes.SeasonPoster &&
				!posterSet.SelectedTypes.Titlecard {
				subAction.AppendWarning(fmt.Sprintf("file_%s", file.Name()), fmt.Sprintf("Poster set '%s' has no selected image types", posterSet.Title))
				fileWarnings = append(fileWarnings, fmt.Sprintf("Poster set '%s' has no selected image types", posterSet.Title))
				go SendNotification(fileIssuesMap[file.Name()], queueItem.MediaItem.Title, posterSet, mediuxItemInfo.TMDB_PosterPath, mediuxItemInfo.TMDB_BackdropPath)
				continue
			}

			LatestInfo.Message = fmt.Sprintf("%s (Set: %s)", queueItem.MediaItem.Title, posterSet.ID)

			for idx, image := range posterSet.Images {
				switch image.Type {
				case "poster":
					if !posterSet.SelectedTypes.Poster {
						continue
					}
				case "backdrop":
					if !posterSet.SelectedTypes.Backdrop {
						continue
					}
				case "season_poster":
					if image.SeasonNumber == nil {
						continue
					} else if *image.SeasonNumber == 0 {
						if !posterSet.SelectedTypes.SpecialSeasonPoster {
							continue
						}
					} else {
						if !posterSet.SelectedTypes.SeasonPoster {
							continue
						}
					}
				case "titlecard":
					if !posterSet.SelectedTypes.Titlecard {
						continue
					}
				default:
					subAction.AppendWarning(fmt.Sprintf("file_%s_image_%d", file.Name(), idx), fmt.Sprintf("Image has unrecognized type '%s'", image.Type))
					fileWarnings = append(fileWarnings, fmt.Sprintf("Image '%s' has unrecognized type '%s'", image.Src, image.Type))
					continue
				}

				// Get the Download File Name
				downloadFileName := utils.GetFileDownloadName(queueItem.MediaItem.Title, image)

				// Download and apply the Image to the Media Item
				Err := mediaserver.DownloadApplyImageToMediaItem(ctx, &queueItem.MediaItem, image)
				if Err.Message != "" {
					subAction.AppendWarning(downloadFileName, "Failed to download and apply image")
					fileErrors = append(fileErrors, fmt.Sprintf("Failed to download and apply image: %s", Err.Message))
				}
			}
		}

		// Now that all of the images have been processed for this item, upsert into the database
		Err = database.UpsertSavedItem(ctx, queueItem)
		if Err.Message != "" {
			subAction.AppendWarning(fmt.Sprintf("file_%s", file.Name()), "Failed to upsert item into database")
			fileErrors = append(fileErrors, fmt.Sprintf("Failed to upsert item into database: %s", Err.Message))
			fileIssuesMap[file.Name()] = FileIssues{Errors: fileErrors, Warnings: fileWarnings}
			go SendNotification(fileIssuesMap[file.Name()], queueItem.MediaItem.Title, models.DBPosterSetDetail{}, mediuxItemInfo.TMDB_PosterPath, mediuxItemInfo.TMDB_BackdropPath)
			ld.Log()
			continue
		}

		var finalErr error
		if len(fileErrors) > 0 {
			newPath := path.Join(FolderPath, fmt.Sprintf("error_%s", file.Name()))
			finalErr = os.Rename(filePath, newPath)
		} else if len(fileWarnings) > 0 {
			newPath := path.Join(FolderPath, fmt.Sprintf("warning_%s", file.Name()))
			finalErr = os.Rename(filePath, newPath)
		} else {
			finalErr = os.Remove(filePath)
		}
		if finalErr != nil {
			subAction.AppendWarning(fmt.Sprintf("file_%s", file.Name()), "Failed to move or delete processed file")
			fileWarnings = append(fileWarnings, fmt.Sprintf("Failed to move or delete processed file: %s", finalErr.Error()))
			fileIssuesMap[file.Name()] = FileIssues{Errors: fileErrors, Warnings: fileWarnings}
			go SendNotification(fileIssuesMap[file.Name()], queueItem.MediaItem.Title, models.DBPosterSetDetail{}, mediuxItemInfo.TMDB_PosterPath, mediuxItemInfo.TMDB_BackdropPath)
			ld.Log()
			continue
		}

		// Handle any labels and tags asynchronously
		go func() {
			ctx, ld := logging.CreateLoggingContext(context.Background(), "Download Queue - Labels and Tags Handling")
			logAction := ld.AddAction("Handle Labels and Tags for Added Item", logging.LevelInfo)
			ctx = logging.WithCurrentAction(ctx, logAction)
			defer ld.Log()

			fileIssuesMap[file.Name()] = FileIssues{Errors: fileErrors, Warnings: fileWarnings}
			SendNotification(fileIssuesMap[file.Name()], queueItem.MediaItem.Title, models.DBPosterSetDetail{}, mediuxItemInfo.TMDB_PosterPath, mediuxItemInfo.TMDB_BackdropPath)

			selectedTypes := models.SelectedTypes{}
			for _, posterSet := range queueItem.PosterSets {
				selectedTypes.Poster = selectedTypes.Poster || posterSet.SelectedTypes.Poster
				selectedTypes.Backdrop = selectedTypes.Backdrop || posterSet.SelectedTypes.Backdrop
				selectedTypes.SeasonPoster = selectedTypes.SeasonPoster || posterSet.SelectedTypes.SeasonPoster
				selectedTypes.SpecialSeasonPoster = selectedTypes.SpecialSeasonPoster || posterSet.SelectedTypes.SpecialSeasonPoster
				selectedTypes.Titlecard = selectedTypes.Titlecard || posterSet.SelectedTypes.Titlecard
			}

			mediaserver.AddLabelToMediaItem(ctx, queueItem.MediaItem, selectedTypes)
			sonarr_radarr.HandleTags(ctx, queueItem.MediaItem, selectedTypes)
		}()

		ld.Log()
	}
}
