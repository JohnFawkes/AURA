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

		actionResultMap := make(map[string]any)

		// If AutoDownload false and rating key has not changed
		if !dbPosterSet.AutoDownload && !ratingKeyChanged {
			setResult.Result = "Skipped"
			setResult.Reason = "Auto download is not selected"
			result.Sets = append(result.Sets, setResult)
			actionResultMap["status"] = "skipped"
			actionResultMap["status_reason"] = "Auto download is not selected"
			logAction.AppendResult(fmt.Sprintf("set_%s", dbPosterSet.PosterSetID), actionResultMap)
			continue
		}

		// If Selected Types are empty, Skip
		if len(dbPosterSet.SelectedTypes) == 0 {
			setResult.Result = "Skipped"
			setResult.Reason = "No selected types"
			result.Sets = append(result.Sets, setResult)
			actionResultMap["status"] = "skipped"
			actionResultMap["status_reason"] = "No selected types"
			logAction.AppendResult(fmt.Sprintf("set_%s", dbPosterSet.PosterSetID), actionResultMap)
			continue
		}

		// Get the latest set information from MediUX
		latestSet, Err := Mediux_FetchShowSetByID(ctx, dbSavedItem.MediaItem.LibraryTitle, dbSavedItem.MediaItem.TMDB_ID, dbPosterSet.PosterSetID)
		if Err.Message != "" {
			setResult.Result = "Error"
			setResult.Reason = fmt.Sprintf("Error fetching updated set - %s", Err.Message)
			result.Sets = append(result.Sets, setResult)
			actionResultMap["status"] = "error"
			actionResultMap["status_reason"] = setResult.Reason
			logAction.AppendResult(fmt.Sprintf("set_%s", dbPosterSet.PosterSetID), actionResultMap)
			continue
		}

		// Check to see if Last Downloaded is greater than Poster Set Last Updated
		posterSetDateNewer := Time_IsLastDownloadedBeforeLatestPosterSetDate(dbPosterSet.LastDownloaded, latestSet.DateUpdated)
		formattedLastDownloaded := formatDateString(dbPosterSet.LastDownloaded)

		// Check if more seasons or episodes were added
		addedSeasonOrEpisodes := MS_AddedMoreSeasonsOrEpisodes(dbSavedItem.MediaItem, latestMediaItem)
		// Check if any episode paths have changed
		episodePathChanges := MS_CheckEpisodePathChanges(dbSavedItem.MediaItem, latestMediaItem)

		actionResultMap["reasons"] = map[string]any{
			"poster_set_updated":     posterSetDateNewer,
			"rating_key_changed":     ratingKeyChanged,
			"added_seasons_episodes": addedSeasonOrEpisodes,
			"episode_path_changes":   episodePathChanges,
		}
		reasonsResultMap := actionResultMap["reasons"].(map[string]any)

		if !posterSetDateNewer && !addedSeasonOrEpisodes && !ratingKeyChanged && !episodePathChanges {
			setResult.Result = "Skipped"
			setResult.Reason = "No updates to poster set, seasons/episodes or rating key"
			result.Sets = append(result.Sets, setResult)
			actionResultMap["status"] = "skipped"
			actionResultMap["status_reason"] = fmt.Sprintf("Skipping set - Last downloaded on %s (Last Update: %s), no updates to poster set, seasons/episodes or rating key",
				formattedLastDownloaded, latestSet.DateUpdated.Format("2006-01-02 15:04:05"))
			logAction.AppendResult(fmt.Sprintf("set_%s", dbPosterSet.PosterSetID), actionResultMap)
			continue
		}

		updateReasons := []string{}
		if posterSetDateNewer {
			updateReasons = append(updateReasons, "Poster set updated")
			reasonsResultMap["poster_set_date_updated"] = latestSet.DateUpdated
			reasonsResultMap["poster_set_last_downloaded"] = dbPosterSet.LastDownloaded
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
		reasonsResultMap["full_reason"] = updateReason

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
		actionResultMap["selected_types"] = dbPosterSet.SelectedTypes

		determineFilesSubAction := logAction.AddSubAction("Determining Files to Download", logging.LevelDebug)
		determineFilesSubActionResultMap := make(map[string]any)
		determineFilesSubActionResultMap["files"] = map[string]any{}
		filesMap := determineFilesSubActionResultMap["files"].(map[string]any)

		filesToDownload := []PosterFileWithReason{}
		filesMap["poster"] = map[string]string{}
		posterMap := filesMap["poster"].(map[string]string)
		if posterSelected {
			posterFileReason := AutoDownload_ShouldDownloadFile(dbPosterSet, *latestSet.Poster, dbSavedItem.MediaItem, latestMediaItem)
			if posterFileReason.ReasonTitle != "" && posterFileReason.ReasonDetails != "" {
				if ratingKeyChanged {
					posterFileReason.ReasonTitle = "Redownloading"
					posterFileReason.ReasonDetails = "Rating key changed and " + posterFileReason.ReasonDetails
				}
				filesToDownload = append(filesToDownload, posterFileReason)
				posterMap["action"] = "download"
				posterMap["action_reason"] = posterFileReason.ReasonDetails
			}
		}

		filesMap["backdrop"] = map[string]string{}
		backdropMap := filesMap["backdrop"].(map[string]string)
		if backdropSelected {
			backdropFileReason := AutoDownload_ShouldDownloadFile(dbPosterSet, *latestSet.Backdrop, dbSavedItem.MediaItem, latestMediaItem)
			if backdropFileReason.ReasonTitle != "" && backdropFileReason.ReasonDetails != "" {
				if ratingKeyChanged {
					backdropFileReason.ReasonTitle = "Redownloading"
					backdropFileReason.ReasonDetails = "Rating key changed and " + backdropFileReason.ReasonDetails
				}
				if backdropFileReason.ReasonDetails != "" {
					filesToDownload = append(filesToDownload, backdropFileReason)
					backdropMap["action"] = "download"
					backdropMap["action_reason"] = backdropFileReason.ReasonDetails
				}
			}
		}
		if seasonSelected {
			for _, season := range latestSet.SeasonPosters {
				if season.Season.Number == 0 {
					continue
				}
				filesMap[fmt.Sprintf("season_s%d", season.Season.Number)] = map[string]string{}
				seasonMap := filesMap[fmt.Sprintf("season_s%d", season.Season.Number)].(map[string]string)
				seasonFileReason := AutoDownload_ShouldDownloadFile(dbPosterSet, season, dbSavedItem.MediaItem, latestMediaItem)
				if seasonFileReason.ReasonTitle != "" && seasonFileReason.ReasonDetails != "" {
					if ratingKeyChanged {
						seasonFileReason.ReasonTitle = "Redownloading"
						seasonFileReason.ReasonDetails = "Rating key changed and " + seasonFileReason.ReasonDetails
					}
					if seasonFileReason.ReasonDetails != "" {
						filesToDownload = append(filesToDownload, seasonFileReason)
						seasonMap["action"] = "download"
						seasonMap["action_reason"] = seasonFileReason.ReasonDetails
					}
				}
			}
		}

		if specialSeasonSelected {
			for _, season := range latestSet.SeasonPosters {
				if season.Season.Number != 0 {
					continue
				}
				filesMap[fmt.Sprintf("season_s%d", season.Season.Number)] = map[string]string{}
				seasonMap := filesMap[fmt.Sprintf("season_s%d", season.Season.Number)].(map[string]string)
				seasonFileReason := AutoDownload_ShouldDownloadFile(dbPosterSet, season, dbSavedItem.MediaItem, latestMediaItem)
				if seasonFileReason.ReasonTitle != "" && seasonFileReason.ReasonDetails != "" {
					if ratingKeyChanged {
						seasonFileReason.ReasonTitle = "Redownloading"
						seasonFileReason.ReasonDetails = "Rating key changed and " + seasonFileReason.ReasonDetails
					}
					if seasonFileReason.ReasonDetails != "" {
						filesToDownload = append(filesToDownload, seasonFileReason)
						seasonMap["action"] = "download"
						seasonMap["action_reason"] = seasonFileReason.ReasonDetails
					}
				}
			}
		}

		if titlecardSelected {
			for _, titlecard := range latestSet.TitleCards {
				filesMap[fmt.Sprintf("titlecard_s%de%d", titlecard.Episode.SeasonNumber, titlecard.Episode.EpisodeNumber)] = map[string]string{}
				titlecardMap := filesMap[fmt.Sprintf("titlecard_s%de%d", titlecard.Episode.SeasonNumber, titlecard.Episode.EpisodeNumber)].(map[string]string)
				titlecardFileReason := AutoDownload_ShouldDownloadFile(dbPosterSet, titlecard, dbSavedItem.MediaItem, latestMediaItem)
				if titlecardFileReason.ReasonTitle != "" && titlecardFileReason.ReasonDetails != "" {
					if ratingKeyChanged {
						titlecardFileReason.ReasonTitle = "Redownloading"
						titlecardFileReason.ReasonDetails = "Rating key changed and " + titlecardFileReason.ReasonDetails
					}
					if titlecardFileReason.ReasonDetails != "" {
						filesToDownload = append(filesToDownload, titlecardFileReason)
						titlecardMap["action"] = "download"
						titlecardMap["action_reason"] = titlecardFileReason.ReasonDetails
					}
				}
			}
		}

		if len(filesToDownload) == 0 {
			setResult.Result = "Skipped"
			setResult.Reason = fmt.Sprintf("No files to download for '%s' in set '%s'", dbSavedItem.MediaItem.Title, dbPosterSet.PosterSetID)
			result.Sets = append(result.Sets, setResult)
			actionResultMap["files_to_download"] = 0
			actionResultMap["status"] = "skipped"
			actionResultMap["status_reason"] = setResult.Reason
			logAction.AppendResult(fmt.Sprintf("set_%s", dbPosterSet.PosterSetID), actionResultMap)
			AutoDownload_UpdateDatabase(ctx, dbSavedItem, latestMediaItem, dbPosterSet, latestSet)
			continue
		} else {
			actionResultMap["files_to_download"] = len(filesToDownload)
		}
		determineFilesSubAction.Complete()

		// Go through each file and download it
		for _, file := range filesToDownload {

			var fileMapID string
			switch file.File.Type {
			case "poster":
				fileMapID = "poster"
			case "backdrop":
				fileMapID = "backdrop"
			case "seasonPoster", "specialSeasonPoster":
				fileMapID = fmt.Sprintf("season_s%d", file.File.Season.Number)
			case "titlecard":
				fileMapID = fmt.Sprintf("titlecard_s%de%d", file.File.Episode.SeasonNumber, file.File.Episode.EpisodeNumber)
			default:
				fileMapID = "unknown"
			}
			fileMap := filesMap[fileMapID].(map[string]string)
			fileMap["file_modified"] = file.File.Modified.Format("2006-01-02 15:04:05")
			fileMap["file_last_downloaded"] = dbPosterSet.LastDownloaded
			fileMap["id"] = file.File.ID
			fileMap["download_reason"] = file.ReasonDetails

			Err = CallDownloadAndUpdatePosters(ctx, latestMediaItem, file.File)
			if Err.Message != "" {
				setResult.Result = "Error"
				setResult.Reason = fmt.Sprintf("Error downloading '%s' for '%s' in set '%s' - %s", file.File.Type, dbSavedItem.MediaItem.Title, dbPosterSet.PosterSetID, Err.Message)
				result.Sets = append(result.Sets, setResult)
				fileMap["download_status"] = "error"
				fileMap["download_error"] = setResult.Reason
				logAction.AppendResult(fmt.Sprintf("file_download_error_%s", file.File.ID), setResult.Reason)
				continue
			}
			fileMap["download_status"] = "success"

			// Send a notification to all configured providers
			// Do this using go func so that it runs asynchronously
			go func() {
				SendFileDownloadNotification(dbSavedItem.MediaItem.Title, dbPosterSet.PosterSetID, file)
			}()
		}

		// Update the database with the latest information
		Err = AutoDownload_UpdateDatabase(ctx, dbSavedItem, latestMediaItem, dbPosterSet, latestSet)
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
		logAction.AppendResult(fmt.Sprintf("set_%s", dbPosterSet.PosterSetID), actionResultMap)
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

func AutoDownload_UpdateMovieDatabase(ctx context.Context, dbSavedItem DBMediaItemWithPosterSets, latestMediaItem MediaItem) logging.LogErrorInfo {
	// Update the database with the latest information
	newDBItem := dbSavedItem
	newDBItem.MediaItem = latestMediaItem
	newDBItem.MediaItemJSON = "" // Clear out the JSON so it gets regenerated
	Err := DB_InsertAllInfoIntoTables(ctx, newDBItem)
	if Err.Message != "" {
		return Err
	}
	return logging.LogErrorInfo{}
}

func AutoDownload_UpdateDatabase(ctx context.Context, dbSavedItem DBMediaItemWithPosterSets, latestMediaItem MediaItem, dbPosterSet DBPosterSetDetail, latestSet PosterSet) logging.LogErrorInfo {
	// Update the database with the latest information
	newDBItem := dbSavedItem
	newDBItem.MediaItem = latestMediaItem
	newDBItem.MediaItemJSON = ""
	newDBItem.PosterSets = []DBPosterSetDetail{}
	newDBItem.PosterSets = make([]DBPosterSetDetail, len(dbSavedItem.PosterSets))
	for i, ps := range dbSavedItem.PosterSets {
		if ps.PosterSetID == dbPosterSet.PosterSetID {
			// Update only the set being worked on
			newDBItem.PosterSets[i] = DBPosterSetDetail{
				PosterSetID:    latestSet.ID,
				PosterSet:      latestSet,
				AutoDownload:   ps.AutoDownload,
				SelectedTypes:  ps.SelectedTypes,
				LastDownloaded: time.Now().Format("2006-01-02 15:04:05"),
			}
		} else {
			// Keep other sets unchanged
			newDBItem.PosterSets[i] = ps
		}
	}
	Err := DB_InsertAllInfoIntoTables(ctx, newDBItem)
	if Err.Message != "" {
		return Err
	}
	return logging.LogErrorInfo{}
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
	Err = AutoDownload_UpdateMovieDatabase(ctx, dbSavedItem, latestMediaItem)
	if Err.Message != "" {
		result.OverAllResult = "Error"
		result.OverAllResultMessage = Err.Message
		return result
	}

	result.OverAllResult = "Redownloaded"
	result.OverAllResultMessage = fmt.Sprintf("Redownloaded %d images for '%s' - Rating key changed from %s to %s", len(dbSavedItem.PosterSets), dbSavedItem.MediaItem.Title, dbSavedItem.MediaItem.RatingKey, latestMediaItem.RatingKey)
	return result
}

func AutoDownload_ShouldDownloadFile(dbPosterSet DBPosterSetDetail, file PosterFile, dbSavedItem, latestMediaItem MediaItem) PosterFileWithReason {
	var psFile PosterFileWithReason
	psFile.File = file
	psFile.ReasonTitle = ""
	psFile.ReasonDetails = ""

	// Do the following checks to see if the file should be downloaded:
	// 1. Check if the File has been updated since the last download
	//		If has been updated:
	//			If the file is a Poster or Backdrop, download the file
	//			If the file is a Season Poster or Titlecard, check to see if the episode exists in the latest media item. If it does, download the file
	//
	// 2. For Season Posters and Titlecards, check if new seasons/episodes were added
	//		If new seasons/episodes were added, download the file

	// Check if the File has been updated since the last download
	fileUpdated := Time_IsLastDownloadedBeforeLatestPosterSetDate(dbPosterSet.LastDownloaded, file.Modified)

	formattedLastDownloaded := formatDateString(dbPosterSet.LastDownloaded)

	switch file.Type {
	case "poster", "backdrop":
		if fileUpdated {
			psFile.ReasonTitle = "Downloading - File Updated"
			psFile.ReasonDetails = fmt.Sprintf("File Updated:\t%s\nLast Download:\t%s", file.Modified.Format("2006-01-02 15:04:05"), formattedLastDownloaded)
			return psFile
		}
	case "seasonPoster", "specialSeasonPoster":
		// For season posters, check if the season exists in the latest media item
		existsInDB, existsInLatest := CheckSeasonExistsAndAdded(file.Season.Number, dbSavedItem, latestMediaItem)

		if fileUpdated && existsInLatest {
			psFile.ReasonTitle = "Downloading - File Updated"
			psFile.ReasonDetails = fmt.Sprintf("File Updated:\t%s\nLast Download:\t%s", file.Modified.Format("2006-01-02 15:04:05"), formattedLastDownloaded)
			return psFile
		} else if !fileUpdated && !existsInDB && existsInLatest {
			psFile.ReasonTitle = "Downloading - New Season Added"
			psFile.ReasonDetails = fmt.Sprintf("Season %d added to %s", file.Season.Number, latestMediaItem.Title)
			return psFile
		}
	case "titlecard":
		// For titlecards, check if the episode exists in the latest media item
		existsInDB, existsInLatest, pathChanged := CheckEpisodeExistsAddedAndPath(file.Episode.SeasonNumber, file.Episode.EpisodeNumber, dbSavedItem, latestMediaItem)
		if fileUpdated && existsInLatest {
			psFile.ReasonTitle = "Downloading - File Updated"
			psFile.ReasonDetails = fmt.Sprintf("File Updated:\t%s\nLast Download:\t%s", file.Modified.Format("2006-01-02 15:04:05"), formattedLastDownloaded)
			return psFile
		} else if !fileUpdated && !existsInDB && existsInLatest {
			psFile.ReasonTitle = "Downloading - New Episode Added"
			psFile.ReasonDetails = fmt.Sprintf("Episode S%02dE%02d added to %s", file.Episode.SeasonNumber, file.Episode.EpisodeNumber, latestMediaItem.Title)
			return psFile
		} else if !fileUpdated && existsInDB && existsInLatest && pathChanged != "" {
			psFile.ReasonTitle = "Downloading - Episode Path Changed"
			psFile.ReasonDetails = pathChanged
			return psFile
		}
	}

	return psFile
}

func AutoDownload_GetLatestMediaItemAndCheckForRatingKeyChanges(ctx context.Context, dbSavedItem DBMediaItemWithPosterSets) (MediaItem, bool, logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, fmt.Sprintf("Checking '%s' (%s) for Rating Key changes", dbSavedItem.MediaItem.Title, dbSavedItem.MediaItem.LibraryTitle), logging.LevelDebug)
	defer logAction.Complete()

	// Get the latest Media Item information from the Media Server
	latestMediaItem, Err := CallFetchItemContent(ctx, dbSavedItem.MediaItem.RatingKey, dbSavedItem.MediaItem.LibraryTitle)
	if Err.Message != "" {

		// Try and get the item from the cache
		cacheItem, found := Global_Cache_LibraryStore.GetMediaItemFromSectionByTMDBID(dbSavedItem.MediaItem.LibraryTitle, dbSavedItem.MediaItem.TMDB_ID)
		if !found {
			return MediaItem{}, false, Err
		} else {
			latestMediaItem = *cacheItem
			logAction.AppendResult("message", "Using cached media item data")
		}
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

	if len(Global_Config.Notifications.Providers) == 0 || Global_Config.Notifications.Enabled == false {
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
			case "Webhook":
				Notification_SendWebhookMessage(
					ctx,
					provider.Webhook,
					messageBody,
					imageURL,
					psFile.ReasonTitle,
				)
			}
		}
	}
}

// Helper Function to format a date string from various formats to "YYYY-MM-DD HH:MM:SS"
// If parsing fails, returns the original string.
func formatDateString(dateStr string) string {
	if dateStr == "" {
		return ""
	}

	formats := []string{
		time.RFC3339,                    // "2006-01-02T15:04:05Z07:00"
		"2006-01-02T15:04:05Z",          // "2025-11-02T16:49:55Z"
		"2006-01-02 15:04:05",           // "2025-11-02 16:49:55"
		"2006-01-02",                    // "2025-11-02"
		"2006-01-02T15:04:05.000Z07:00", // with milliseconds
		"2006-01-02T15:04:05.000Z",      // with milliseconds and Z
	}

	for _, layout := range formats {
		if t, err := time.Parse(layout, dateStr); err == nil {
			return t.Format("2006-01-02 15:04:05")
		}
	}
	return dateStr
}
