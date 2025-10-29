package api

import (
	"aura/internal/logging"
	"context"
	"fmt"
	"strings"
	"time"
)

func AutoDownload_CheckForUpdatesToPosters() {
	// Get all items from the database
	dbSavedItems, Err := DB_GetAllItems()
	if Err.Message != "" {
		return
	}

	// Loop through each item in the database
	for _, dbSavedItem := range dbSavedItems {
		ctx, ld := logging.CreateLoggingContext(context.Background(), "AutoDownload - Check For Updates")
		defer ld.Log()
		logAction := ld.AddAction(fmt.Sprintf("Checking for Updates to '%s' in %s", dbSavedItem.MediaItem.Title, dbSavedItem.LibraryTitle), logging.LevelInfo)
		ctx = logging.WithCurrentAction(ctx, logAction)
		AutoDownload_CheckItem(ctx, dbSavedItem)
	}
}

type AutoDownloadResult struct {
	MediaItemTitle       string
	Sets                 []AutoDownloadSetResult
	OverAllResult        string
	OverAllResultMessage string
}

type AutoDownloadSetResult struct {
	PosterSetID string
	Result      string
	Reason      string
}

type PosterFileWithReason struct {
	File          PosterFile
	ReasonTitle   string
	ReasonDetails string
}

func AutoDownload_CheckItem(ctx context.Context, dbSavedItem DBMediaItemWithPosterSets) AutoDownloadResult {
	ctx, logAction := logging.AddSubActionToContext(ctx, fmt.Sprintf("Checking '%s' (%s)", dbSavedItem.MediaItem.Title, dbSavedItem.MediaItem.LibraryTitle), logging.LevelInfo)
	defer logAction.Complete()

	var result AutoDownloadResult
	result.MediaItemTitle = dbSavedItem.MediaItem.Title

	// If the item is a movie, then just check for rating key changes
	if dbSavedItem.MediaItem.Type == "movie" {
		return AutoDownload_CheckMovieForKeyChanges(ctx, dbSavedItem)
	}

	// If item is not a show, Skip
	if dbSavedItem.MediaItem.Type != "show" {
		result.OverAllResult = "Skipped"
		result.OverAllResultMessage = fmt.Sprintf("Skipping '%s' - Not a show", dbSavedItem.MediaItem.Title)
		logAction.AppendResult("item_skipped", result.OverAllResultMessage)
		return result
	}

	// If there is no TMDB ID, Skip
	if dbSavedItem.MediaItem.TMDB_ID == "" {
		result.OverAllResult = "Skipped"
		result.OverAllResultMessage = fmt.Sprintf("Skipping '%s' - No TMDB ID", dbSavedItem.MediaItem.Title)
		logAction.AppendResult("item_skipped", result.OverAllResultMessage)
		return result
	}

	// Get the latest Media Item from the Server
	latestMediaItem, ratingKeyChanged, Err := AutoDownload_GetLatestMediaItemAndCheckForRatingKeyChanges(ctx, dbSavedItem)
	if Err.Message != "" {
		result.OverAllResult = "Error"
		result.OverAllResultMessage = Err.Message
		return result
	}

	// Update the media item title in case it changed
	result.MediaItemTitle = latestMediaItem.Title

	// Loop through each poster set for the item
	for _, dbPosterSet := range dbSavedItem.PosterSets {
		var setResult AutoDownloadSetResult
		setResult.PosterSetID = dbPosterSet.PosterSetID

		// If AutoDownload false and rating key has not changed
		if !dbPosterSet.AutoDownload && !ratingKeyChanged {
			setResult.Result = "Skipped"
			setResult.Reason = "Auto download is not selected"
			result.Sets = append(result.Sets, setResult)
			logAction.AppendResult("set_"+dbPosterSet.PosterSetID, "Auto download is not selected")
			continue
		}

		// If Selected Types are empty, Skip
		if len(dbPosterSet.SelectedTypes) == 0 {
			setResult.Result = "Skipped"
			setResult.Reason = "No selected types"
			result.Sets = append(result.Sets, setResult)
			logAction.AppendResult("set_"+dbPosterSet.PosterSetID, "No selected types")
			continue
		}

		// Get the latest set information from MediUX
		latestSet, Err := Mediux_FetchShowSetByID(ctx, dbSavedItem.MediaItem.LibraryTitle, dbSavedItem.MediaItem.TMDB_ID, dbPosterSet.PosterSetID)
		if Err.Message != "" {
			setResult.Result = "Error"
			setResult.Reason = fmt.Sprintf("Error fetching updated set - %s", Err.Message)
			result.Sets = append(result.Sets, setResult)
			logAction.AppendResult("set_"+dbPosterSet.PosterSetID, fmt.Sprintf("Error fetching updated set - %s", Err.Message))
			continue
		}

		// Check to see if Last Downloaded is greater than Poster Set Last Updated
		posterSetDateNewer := Time_IsLastDownloadedBeforeLatestPosterSetDate(ctx, dbPosterSet.LastDownloaded, latestSet.DateUpdated)
		var formattedLastDownloaded string
		lastDownloadedTime, err := time.Parse("2006-01-02T15:04:05Z07:00", dbPosterSet.LastDownloaded)
		if err == nil {
			formattedLastDownloaded = lastDownloadedTime.Format("2006-01-02 15:04:05")
		} else {
			formattedLastDownloaded = dbPosterSet.LastDownloaded
		}

		// Check if more seasons or episodes were added
		addedSeasonOrEpisodes := MS_AddedMoreSeasonsOrEpisodes(dbSavedItem.MediaItem, latestMediaItem)
		// Check if any episode paths have changed
		episodePathChanges := MS_CheckEpisodePathChanges(dbSavedItem.MediaItem, latestMediaItem)

		if !posterSetDateNewer && !addedSeasonOrEpisodes && !ratingKeyChanged && !episodePathChanges {
			setResult.Result = "Skipped"
			setResult.Reason = "No updates to poster set, seasons/episodes or rating key"
			result.Sets = append(result.Sets, setResult)
			logAction.AppendResult("set_"+dbPosterSet.PosterSetID, fmt.Sprintf("Skipping set - Last downloaded on %s, no updates to poster set, seasons/episodes or rating key", formattedLastDownloaded))
			continue
		}

		updateReasons := []string{}
		if posterSetDateNewer {
			updateReasons = append(updateReasons, "Poster set updated")
		}
		if addedSeasonOrEpisodes {
			updateReasons = append(updateReasons, "New seasons or episodes added")
		}
		if ratingKeyChanged {
			updateReasons = append(updateReasons, "Rating key changed")
		}
		if episodePathChanges {
			updateReasons = append(updateReasons, "Episode path changes detected")
		}
		updateReason := strings.Join(updateReasons, " and ")
		logAction.AppendResult("set_"+dbPosterSet.PosterSetID, fmt.Sprintf("Updating set - %s", updateReason))

		// Check if the selected types contains each image type
		posterSelected := false
		backdropSelected := false
		seasonSelected := false
		specialSeasonSelected := false
		titlecardSelected := false
		for _, selectedType := range dbPosterSet.SelectedTypes {
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

		filesToDownload := []PosterFileWithReason{}
		if posterSelected {
			posterFileReason := AutoDownload_ShouldDownloadFile(ctx, dbPosterSet, *latestSet.Poster, dbSavedItem.MediaItem, latestMediaItem)
			if posterFileReason.ReasonTitle != "" && posterFileReason.ReasonDetails != "" {
				if ratingKeyChanged {
					posterFileReason.ReasonTitle = "Redownloading"
					posterFileReason.ReasonDetails = "Rating key changed and " + posterFileReason.ReasonDetails
				}
				if posterFileReason.ReasonDetails != "" {
					filesToDownload = append(filesToDownload, posterFileReason)
					logAction.AppendResult(fmt.Sprintf("set_%s_poster", dbPosterSet.PosterSetID),
						fmt.Sprintf("Queuing poster file for download - %s", posterFileReason.ReasonDetails))
				}
			}
		}
		if backdropSelected {
			backdropFileReason := AutoDownload_ShouldDownloadFile(ctx, dbPosterSet, *latestSet.Backdrop, dbSavedItem.MediaItem, latestMediaItem)
			if backdropFileReason.ReasonTitle != "" && backdropFileReason.ReasonDetails != "" {
				if ratingKeyChanged {
					backdropFileReason.ReasonTitle = "Redownloading"
					backdropFileReason.ReasonDetails = "Rating key changed and " + backdropFileReason.ReasonDetails
				}
				if backdropFileReason.ReasonDetails != "" {
					filesToDownload = append(filesToDownload, backdropFileReason)
					logAction.AppendResult(fmt.Sprintf("set_%s_backdrop", dbPosterSet.PosterSetID),
						fmt.Sprintf("Queuing backdrop file for download - %s", backdropFileReason.ReasonDetails))
				}
			}
		}
		for _, season := range latestSet.SeasonPosters {
			if season.Season.Number == 0 {
				if specialSeasonSelected {
					specialSeasonFileReason := AutoDownload_ShouldDownloadFile(ctx, dbPosterSet, season, dbSavedItem.MediaItem, latestMediaItem)
					if specialSeasonFileReason.ReasonTitle != "" && specialSeasonFileReason.ReasonDetails != "" {
						if ratingKeyChanged {
							specialSeasonFileReason.ReasonTitle = "Redownloading"
							specialSeasonFileReason.ReasonDetails = "Rating key changed and " + specialSeasonFileReason.ReasonDetails
						}
						if specialSeasonFileReason.ReasonDetails != "" {
							filesToDownload = append(filesToDownload, specialSeasonFileReason)
							logAction.AppendResult(fmt.Sprintf("set_%s_season_s0", dbPosterSet.PosterSetID),
								fmt.Sprintf("Queuing special season poster file for download - %s", specialSeasonFileReason.ReasonDetails))
						}
					}
				}
			} else {
				if seasonSelected {
					seasonFileReason := AutoDownload_ShouldDownloadFile(ctx, dbPosterSet, season, dbSavedItem.MediaItem, latestMediaItem)
					if seasonFileReason.ReasonTitle != "" && seasonFileReason.ReasonDetails != "" {
						if ratingKeyChanged {
							seasonFileReason.ReasonTitle = "Redownloading"
							seasonFileReason.ReasonDetails = "Rating key changed and " + seasonFileReason.ReasonDetails
						}
						if seasonFileReason.ReasonDetails != "" {
							filesToDownload = append(filesToDownload, seasonFileReason)
							logAction.AppendResult(fmt.Sprintf("set_%s_season_s%d", dbPosterSet.PosterSetID, season.Season.Number),
								fmt.Sprintf("Queuing season poster file for download - %s", seasonFileReason.ReasonDetails))
						}
					}
				}
			}
		}
		if titlecardSelected {
			for _, titlecard := range latestSet.TitleCards {
				titlecardFileReason := AutoDownload_ShouldDownloadFile(ctx, dbPosterSet, titlecard, dbSavedItem.MediaItem, latestMediaItem)
				if titlecardFileReason.ReasonTitle != "" && titlecardFileReason.ReasonDetails != "" {
					if ratingKeyChanged {
						titlecardFileReason.ReasonTitle = "Redownloading"
						titlecardFileReason.ReasonDetails = "Rating key changed and " + titlecardFileReason.ReasonDetails
					}
					if titlecardFileReason.ReasonDetails != "" {
						filesToDownload = append(filesToDownload, titlecardFileReason)
						logAction.AppendResult(fmt.Sprintf("set_%s_titlecard_s%d_e%d", dbPosterSet.PosterSetID, titlecard.Episode.SeasonNumber, titlecard.Episode.EpisodeNumber),
							fmt.Sprintf("Queuing titlecard file for download - %s", titlecardFileReason.ReasonDetails))
					}
				}
			}
		}

		if len(filesToDownload) == 0 {
			setResult.Result = "Skipped"
			setResult.Reason = fmt.Sprintf("No files to download for '%s' in set '%s'", dbSavedItem.MediaItem.Title, dbPosterSet.PosterSetID)
			result.Sets = append(result.Sets, setResult)
			logAction.AppendResult("files_to_download", setResult.Reason)
			continue
		} else {
			logAction.AppendResult("files_to_download", fmt.Sprintf("Downloading %d files for set '%s'", len(filesToDownload), dbPosterSet.PosterSetID))
		}

		// Go through each file and download it
		for _, file := range filesToDownload {
			// Add in the reason for download to the log and file modified and last downloaded
			logAction.AppendResult("file_"+file.File.ID, fmt.Sprintf("Downloading file type %s - %s", file.File.Type, file.ReasonDetails))
			logAction.AppendResult("file_"+file.File.ID+"_modified", file.File.Modified.Format("2006-01-02 15:04:05"))
			logAction.AppendResult("file_"+file.File.ID+"_last_downloaded", dbPosterSet.LastDownloaded)

			Err = CallDownloadAndUpdatePosters(ctx, latestMediaItem, file.File)
			if Err.Message != "" {
				setResult.Result = "Error"
				setResult.Reason = fmt.Sprintf("Error downloading '%s' for '%s' in set '%s' - %s", file.File.Type, dbSavedItem.MediaItem.Title, dbPosterSet.PosterSetID, Err.Message)
				result.Sets = append(result.Sets, setResult)
				logAction.AppendResult(fmt.Sprintf("file_download_error_%s", file.File.ID), setResult.Reason)
				continue
			}
			logAction.AppendResult("file_"+file.File.ID, fmt.Sprintf("Downloaded file type %s", file.File.Type))

			// Send a notification to all configured providers
			// Do this using go func so that it runs asynchronously
			go func() {
				SendFileDownloadNotification(dbSavedItem.MediaItem.Title, dbPosterSet.PosterSetID, file)
			}()
		}

		// Update the database with the latest information
		newDBItem := dbSavedItem
		newDBItem.MediaItem = latestMediaItem
		newDBItem.MediaItemJSON = ""
		newDBItem.PosterSets = []DBPosterSetDetail{}
		newDBItem.PosterSets = append([]DBPosterSetDetail{}, dbSavedItem.PosterSets...)
		Err = DB_InsertAllInfoIntoTables(ctx, newDBItem)
		if Err.Message != "" {
			setResult.Result = "Error"
			setResult.Reason = fmt.Sprintf("Error updating database for '%s' in set '%s' - %s", dbSavedItem.MediaItem.Title, dbPosterSet.PosterSetID, Err.Message)
			result.Sets = append(result.Sets, setResult)
			logAction.AppendResult("db_update_error", setResult.Reason)
			continue
		}
		setResult.Result = "Success"
		setResult.Reason = fmt.Sprintf("Successfully downloaded files for '%s' in set '%s'", dbSavedItem.MediaItem.Title, dbPosterSet.PosterSetID)
		result.Sets = append(result.Sets, setResult)
		logAction.AppendResult("set_completed", setResult.Reason)
	}

	// Set overall result based on the results of the sets
	if len(result.Sets) == 0 {
		result.OverAllResult = "Skipped"
		result.OverAllResultMessage = fmt.Sprintf("No sets to check for '%s'", dbSavedItem.MediaItem.Title)
	} else {
		successCount := 0
		errorCount := 0
		skippedCount := 0
		totalCount := len(result.Sets)

		for _, setResult := range result.Sets {
			switch setResult.Result {
			case "Success":
				successCount++
			case "Error":
				errorCount++
			case "Skipped":
				skippedCount++
			}
		}

		switch {
		case errorCount == totalCount:
			result.OverAllResult = "Error"
			result.OverAllResultMessage = "All set downloads failed"
		case errorCount > 0 && successCount > 0:
			result.OverAllResult = "Warn"
			result.OverAllResultMessage = fmt.Sprintf("Success %d | Error %d | Skipped %d",
				successCount, errorCount, skippedCount)
		case successCount == totalCount:
			result.OverAllResult = "Success"
			result.OverAllResultMessage = "All set downloads successful"
		case skippedCount == totalCount:
			result.OverAllResult = "Skipped"
			result.OverAllResultMessage = "No set updates found"
		}
	}
	return result
}

func AutoDownload_CheckMovieForKeyChanges(ctx context.Context, dbSavedItem DBMediaItemWithPosterSets) AutoDownloadResult {
	ctx, logAction := logging.AddSubActionToContext(ctx, fmt.Sprintf("Checking Movie '%s' (%s) for Key changes", dbSavedItem.MediaItem.Title, dbSavedItem.MediaItem.LibraryTitle), logging.LevelInfo)
	defer logAction.Complete()

	var result AutoDownloadResult
	result.MediaItemTitle = dbSavedItem.MediaItem.Title

	// Get the latest Media Item information and check if the Rating Key has changed
	latestMediaItem, ratingKeyChanged, Err := AutoDownload_GetLatestMediaItemAndCheckForRatingKeyChanges(ctx, dbSavedItem)
	if Err.Message != "" {
		result.OverAllResult = "Error"
		result.OverAllResultMessage = Err.Message
		return result
	}

	ratingKeyString := ""
	if ratingKeyChanged {
		ratingKeyString = fmt.Sprintf("Rating key changed from %s to %s", dbSavedItem.MediaItem.RatingKey, latestMediaItem.RatingKey)
		logAction.AppendResult("message", ratingKeyString)
	}

	// Check to see if the Path has changed
	pathChanged := false
	pathChangedString := ""
	if dbSavedItem.MediaItem.Movie != nil && latestMediaItem.Movie != nil {
		if dbSavedItem.MediaItem.Movie.File.Path != latestMediaItem.Movie.File.Path {
			pathChanged = true
			pathChangedString = fmt.Sprintf("Path changed from '%s' to '%s'. ", dbSavedItem.MediaItem.Movie.File.Path, latestMediaItem.Movie.File.Path)
			logAction.AppendResult("message", pathChangedString)
		}
	}

	if !ratingKeyChanged && !pathChanged {
		result.OverAllResult = "Skipped"
		result.OverAllResultMessage = fmt.Sprintf("Skipping '%s' - No changes to rating key or path", dbSavedItem.MediaItem.Title)
		logAction.AppendResult("message", result.OverAllResultMessage)
		return result
	}

	// Update the media item title in case it changed
	result.MediaItemTitle = latestMediaItem.Title

	// Since the rating key has changed, redownload the images and update the database
	for _, dbPosterSet := range dbSavedItem.PosterSets {
		filesToDownload := []PosterFile{}

		posterSelected := false
		backdropSelected := false
		for _, selectedType := range dbPosterSet.SelectedTypes {
			switch selectedType {
			case "poster":
				posterSelected = true
			case "backdrop":
				backdropSelected = true
			}
		}

		if posterSelected && dbPosterSet.PosterSet.Poster != nil {
			filesToDownload = append(filesToDownload, *dbPosterSet.PosterSet.Poster)
		}
		if backdropSelected && dbPosterSet.PosterSet.Backdrop != nil {
			filesToDownload = append(filesToDownload, *dbPosterSet.PosterSet.Backdrop)
		}

		if len(filesToDownload) == 0 {
			var setResult AutoDownloadSetResult
			setResult.PosterSetID = dbPosterSet.PosterSetID
			setResult.Result = "Skipped"
			setResult.Reason = "No files to download for selected types"
			result.Sets = append(result.Sets, setResult)
			logAction.AppendResult("set_"+dbPosterSet.PosterSetID, "No files to download for selected types")
			continue
		}

		// Download and update the media item with the new files
		for _, file := range filesToDownload {
			logAction.AppendResult("file_"+file.Type, fmt.Sprintf("Downloading file type %s", file.Type))
			Err := CallDownloadAndUpdatePosters(ctx, latestMediaItem, file)
			if Err.Message != "" {
				continue
			}
			logAction.AppendResult("file_"+file.Type, fmt.Sprintf("Downloaded file type %s", file.Type))
			// Send a notification to all configured providers
			// Do this using go func so that it runs asynchronously
			go func() {
				var details string
				if ratingKeyString != "" {
					details += ratingKeyString
				}
				if pathChangedString != "" {
					if details != "" {
						details += " and "
					}
					details += pathChangedString
				}
				psFileWithReason := PosterFileWithReason{
					File:          file,
					ReasonTitle:   "Redownloading",
					ReasonDetails: details,
				}
				SendFileDownloadNotification(dbSavedItem.MediaItem.Title, dbPosterSet.PosterSetID, psFileWithReason)
			}()
		}

	}

	// Update the database with the latest information
	newDBItem := dbSavedItem
	newDBItem.MediaItem = latestMediaItem
	newDBItem.MediaItemJSON = "" // Clear out the JSON so it gets regenerated
	Err = DB_InsertAllInfoIntoTables(ctx, newDBItem)
	if Err.Message != "" {
		result.OverAllResult = "Error"
		result.OverAllResultMessage = Err.Message
		return result
	}

	result.OverAllResult = "Redownloaded"
	result.OverAllResultMessage = fmt.Sprintf("Redownloaded %d images for '%s' - Rating key changed from %s to %s", len(dbSavedItem.PosterSets), dbSavedItem.MediaItem.Title, dbSavedItem.MediaItem.RatingKey, latestMediaItem.RatingKey)
	return result
}

func AutoDownload_ShouldDownloadFile(ctx context.Context, dbPosterSet DBPosterSetDetail, file PosterFile, dbSavedItem, latestMediaItem MediaItem) PosterFileWithReason {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Checking if file should be downloaded", logging.LevelDebug)
	defer logAction.Complete()

	var psFile PosterFileWithReason
	psFile.File = file
	psFile.ReasonTitle = ""
	psFile.ReasonDetails = ""

	// Check if file was modified after last download
	fileUpdated := Time_IsLastDownloadedBeforeLatestPosterSetDate(ctx, dbPosterSet.LastDownloaded, file.Modified)

	// If the file was updated, download it
	if fileUpdated {
		psFile.ReasonTitle = "Downloading - File Updated"
		psFile.ReasonDetails = fmt.Sprintf("File updated on %s (last download was %s)", file.Modified.Format("2006-01-02 15:04:05"), dbPosterSet.LastDownloaded)
		return psFile
	}

	// For season posters and titlecards, also check if new seasons/episodes were added
	// If so, download the file
	// If not, also check if the season/episode path has changed
	// If so, download the file
	// If none of these conditions are met, do not download the file
	switch file.Type {
	case "seasonPoster", "specialSeasonPoster":
		CheckSeasonAdded(ctx, file.Season.Number, dbSavedItem, latestMediaItem, &psFile)
		return psFile
	case "titlecard":
		CheckEpisodeAdded(ctx, file.Episode.SeasonNumber, file.Episode.EpisodeNumber, dbSavedItem, latestMediaItem, &psFile)
	default:
		return PosterFileWithReason{}
	}
	return psFile
}

func AutoDownload_GetLatestMediaItemAndCheckForRatingKeyChanges(ctx context.Context, dbSavedItem DBMediaItemWithPosterSets) (MediaItem, bool, logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, fmt.Sprintf("Checking '%s' (%s) for Rating Key changes", dbSavedItem.MediaItem.Title, dbSavedItem.MediaItem.LibraryTitle), logging.LevelDebug)
	defer logAction.Complete()

	// Get the latest Media Item information from the Media Server
	latestMediaItem, Err := CallFetchItemContent(ctx, dbSavedItem.MediaItem.RatingKey, dbSavedItem.MediaItem.LibraryTitle)
	if Err.Message != "" {
		return MediaItem{}, false, Err
	}

	// If the TMDB ID and Library Title do not match, skip
	if dbSavedItem.MediaItem.TMDB_ID != latestMediaItem.TMDB_ID || dbSavedItem.MediaItem.LibraryTitle != latestMediaItem.LibraryTitle {
		logAction.SetError("TMDB ID or Library Title does not match", "If you think this is a mistake, try again.",
			map[string]any{
				"title":             dbSavedItem.MediaItem.Title,
				"year":              dbSavedItem.MediaItem.Year,
				"tmdb_id_old":       dbSavedItem.MediaItem.TMDB_ID,
				"tmdb_id_new":       latestMediaItem.TMDB_ID,
				"library_title_old": dbSavedItem.MediaItem.LibraryTitle,
				"library_title_new": latestMediaItem.LibraryTitle,
			})
	}

	// Check to see if the Rating Key has changed
	ratingKeyChanged := dbSavedItem.MediaItem.RatingKey != latestMediaItem.RatingKey

	return latestMediaItem, ratingKeyChanged, logging.LogErrorInfo{}
}

func SendFileDownloadNotification(itemTitle, posterSetID string, psFile PosterFileWithReason) {

	if len(Global_Config.Notifications.Providers) == 0 {
		return
	}

	messageBody := fmt.Sprintf(
		"%s (Set: %s) - %s\n\nReason:\n%s",
		itemTitle,
		posterSetID,
		MediaServer_GetFileDownloadName(psFile.File),
		psFile.ReasonDetails,
	)

	imageURL := fmt.Sprintf("%s/%s?v=%s&key=jpg",
		"https://images.mediux.io/assets",
		psFile.File.ID,
		psFile.File.Modified.Format("20060102150405"),
	)

	ctx, ld := logging.CreateLoggingContext(context.Background(), "Notification - Send File Download Message")
	logAction := ld.AddAction("Sending File Download Notification", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)
	defer ld.Log()
	defer logAction.Complete()

	// Send a notification to all configured providers
	for _, provider := range Global_Config.Notifications.Providers {
		if provider.Enabled {
			switch provider.Provider {
			case "Discord":
				Notification_SendDiscordMessage(
					ctx,
					provider.Discord,
					messageBody,
					imageURL,
					psFile.ReasonTitle,
				)
			case "Pushover":
				Notification_SendPushoverMessage(
					ctx,
					provider.Pushover,
					messageBody,
					imageURL,
					psFile.ReasonTitle,
				)
			case "Gotify":
				Notification_SendGotifyMessage(
					ctx,
					provider.Gotify,
					messageBody,
					imageURL,
					psFile.ReasonTitle,
				)
			}
		}
	}
}
