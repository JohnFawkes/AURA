package api

import (
	"aura/internal/logging"
	"fmt"
	"strings"
	"time"
)

func AutoDownload_CheckForUpdatesToPosters() {
	// Get all items from the database
	dbSavedItems, Err := DB_GetAllItems()
	if Err.Message != "" {
		logging.LOG.ErrorWithLog(Err)
		return
	}

	// Loop through each item in the database
	for _, dbSavedItem := range dbSavedItems {
		AutoDownload_CheckItem(dbSavedItem)
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

func AutoDownload_CheckItem(dbSavedItem DBMediaItemWithPosterSets) AutoDownloadResult {
	var result AutoDownloadResult
	result.MediaItemTitle = dbSavedItem.MediaItem.Title

	// If the item is a movie, then just check for rating key changes
	if dbSavedItem.MediaItem.Type == "movie" {
		return AutoDownload_CheckMovieForKeyChanges(dbSavedItem)
	}

	// If item is not a show, Skip
	if dbSavedItem.MediaItem.Type != "show" {
		result.OverAllResult = "Skipped"
		result.OverAllResultMessage = fmt.Sprintf("Skipping '%s' - Not a show", dbSavedItem.MediaItem.Title)
		return result
	}

	// If there is no TMDB ID, Skip
	if dbSavedItem.MediaItem.TMDB_ID == "" {
		result.OverAllResult = "Skipped"
		result.OverAllResultMessage = fmt.Sprintf("Skipping '%s' - No TMDB ID", dbSavedItem.MediaItem.Title)
		return result
	}

	logging.LOG.Debug(fmt.Sprintf("Checking for updates to posters for '%s'", dbSavedItem.MediaItem.Title))

	// Get the latest Media Item from the Server
	latestMediaItem, ratingKeyChanged, Err := AutoDownload_GetLatestMediaItemAndCheckForChanges(dbSavedItem)
	if Err.Message != "" {
		logging.LOG.ErrorWithLog(Err)
		result.OverAllResult = "Error"
		result.OverAllResultMessage = Err.Message
		return result
	}
	if ratingKeyChanged {
		logging.LOG.Info(fmt.Sprintf("Rating key changed for '%s' - %s to %s...redownloading", dbSavedItem.MediaItem.Title, dbSavedItem.MediaItem.RatingKey, latestMediaItem.RatingKey))
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
			logging.LOG.Trace(fmt.Sprintf("Skipping poster set '%s' for '%s' - Auto download is not selected", dbPosterSet.PosterSetID, dbSavedItem.MediaItem.Title))
			continue
		}

		// If Selected Types are empty, Skip
		if len(dbPosterSet.SelectedTypes) == 0 {
			setResult.Result = "Skipped"
			setResult.Reason = "No selected types"
			result.Sets = append(result.Sets, setResult)
			logging.LOG.Trace(fmt.Sprintf("Skipping poster set '%s' for '%s' - No selected types", dbPosterSet.PosterSetID, dbSavedItem.MediaItem.Title))
			continue
		}

		logging.LOG.Trace(fmt.Sprintf("Checking Set '%s' by '%s' for '%s'", dbPosterSet.PosterSetID, dbPosterSet.PosterSet.User.Name, dbSavedItem.MediaItem.Title))

		// Get the latest set information from MediUX
		latestSet, Err := Mediux_FetchShowSetByID(dbSavedItem.MediaItem.LibraryTitle, dbSavedItem.MediaItem.TMDB_ID, dbPosterSet.PosterSetID)
		if Err.Message != "" {
			logging.LOG.ErrorWithLog(Err)
			logging.LOG.Warn(fmt.Sprintf("Set '%s' for '%s' possibly deleted from Mediux - %s", dbPosterSet.PosterSetID, dbSavedItem.MediaItem.Title, Err.Message))
			setResult.Result = "Error"
			setResult.Reason = fmt.Sprintf("Error fetching updated set - %s", Err.Message)
			result.Sets = append(result.Sets, setResult)
			continue
		}

		// Check to see if Last Downloaded is greater than Poster Set Last Updated
		posterSetDateNewer := Time_IsLastDownloadedBeforeLatestPosterSetDate(dbPosterSet.LastDownloaded, latestSet.DateUpdated)
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
			logging.LOG.Debug(fmt.Sprintf("Skipping '%s' - Set '%s'\nLast Update: %s\nLast Download: %s\nNo update to poster set, seasons/episodes or rating key",
				dbSavedItem.MediaItem.Title,
				dbPosterSet.PosterSetID,
				latestSet.DateUpdated.Format("2006-01-02 15:04:05"),
				formattedLastDownloaded,
			))
			setResult.Result = "Skipped"
			setResult.Reason = "No updates to poster set, seasons/episodes or rating key"
			result.Sets = append(result.Sets, setResult)
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
		logging.LOG.Debug(fmt.Sprintf("'%s' - Set '%s' updated. %s", dbSavedItem.MediaItem.Title, dbPosterSet.PosterSetID, updateReason))

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
		logging.LOG.Debug(fmt.Sprintf("Downloading selected types: %s", strings.Join(dbPosterSet.SelectedTypes, ", ")))

		filesToDownload := []PosterFileWithReason{}
		if posterSelected {
			posterFileReason := AutoDownload_ShouldDownloadFile(dbPosterSet, *latestSet.Poster, dbSavedItem.MediaItem, latestMediaItem)
			if posterFileReason.ReasonTitle != "" && posterFileReason.ReasonDetails != "" {
				if ratingKeyChanged {
					posterFileReason.ReasonTitle = "Redownloading"
					posterFileReason.ReasonDetails = "Rating key changed and " + posterFileReason.ReasonDetails
				}
				if posterFileReason.ReasonDetails != "" {
					filesToDownload = append(filesToDownload, posterFileReason)
				}
			}
		}
		if backdropSelected {
			backdropFileReason := AutoDownload_ShouldDownloadFile(dbPosterSet, *latestSet.Backdrop, dbSavedItem.MediaItem, latestMediaItem)
			if backdropFileReason.ReasonTitle != "" && backdropFileReason.ReasonDetails != "" {
				if ratingKeyChanged {
					backdropFileReason.ReasonTitle = "Redownloading"
					backdropFileReason.ReasonDetails = "Rating key changed and " + backdropFileReason.ReasonDetails
				}
				if backdropFileReason.ReasonDetails != "" {
					filesToDownload = append(filesToDownload, backdropFileReason)
				}
			}
		}
		for _, season := range latestSet.SeasonPosters {
			if season.Season.Number == 0 {
				if specialSeasonSelected {
					specialSeasonFileReason := AutoDownload_ShouldDownloadFile(dbPosterSet, season, dbSavedItem.MediaItem, latestMediaItem)
					if specialSeasonFileReason.ReasonTitle != "" && specialSeasonFileReason.ReasonDetails != "" {
						if ratingKeyChanged {
							specialSeasonFileReason.ReasonTitle = "Redownloading"
							specialSeasonFileReason.ReasonDetails = "Rating key changed and " + specialSeasonFileReason.ReasonDetails
						}
						if specialSeasonFileReason.ReasonDetails != "" {
							filesToDownload = append(filesToDownload, specialSeasonFileReason)
						}
					}
				}
			} else {
				if seasonSelected {
					seasonFileReason := AutoDownload_ShouldDownloadFile(dbPosterSet, season, dbSavedItem.MediaItem, latestMediaItem)
					if seasonFileReason.ReasonTitle != "" && seasonFileReason.ReasonDetails != "" {
						if ratingKeyChanged {
							seasonFileReason.ReasonTitle = "Redownloading"
							seasonFileReason.ReasonDetails = "Rating key changed and " + seasonFileReason.ReasonDetails
						}
						if seasonFileReason.ReasonDetails != "" {
							filesToDownload = append(filesToDownload, seasonFileReason)
						}
					}
				}
			}
		}
		if titlecardSelected {
			for _, titlecard := range latestSet.TitleCards {
				titlecardFileReason := AutoDownload_ShouldDownloadFile(dbPosterSet, titlecard, dbSavedItem.MediaItem, latestMediaItem)
				if titlecardFileReason.ReasonTitle != "" && titlecardFileReason.ReasonDetails != "" {
					if ratingKeyChanged {
						titlecardFileReason.ReasonTitle = "Redownloading"
						titlecardFileReason.ReasonDetails = "Rating key changed and " + titlecardFileReason.ReasonDetails
					}
					if titlecardFileReason.ReasonDetails != "" {
						filesToDownload = append(filesToDownload, titlecardFileReason)
					}
				}
			}
		}

		if len(filesToDownload) == 0 {
			logging.LOG.Debug(fmt.Sprintf("No files to download for '%s' in set '%s'", dbSavedItem.MediaItem.Title, dbPosterSet.PosterSetID))
			setResult.Result = "Skipped"
			setResult.Reason = fmt.Sprintf("No files to download for '%s' in set '%s'", dbSavedItem.MediaItem.Title, dbPosterSet.PosterSetID)
			result.Sets = append(result.Sets, setResult)
			continue
		}

		// Go through each file and download it
		for _, file := range filesToDownload {
			logging.LOG.Info(fmt.Sprintf("Downloading new '%s' for '%s' in set '%s'\nReason: %s", file.File.Type, dbSavedItem.MediaItem.Title, dbPosterSet.PosterSetID, file.ReasonDetails))
			logging.LOG.Debug(fmt.Sprintf("File modified: %s | Last download: %s",
				file.File.Modified.Format("2006-01-02 15:04:05"),
				dbPosterSet.LastDownloaded,
			))

			Err = CallDownloadAndUpdatePosters(latestMediaItem, file.File)
			if Err.Message != "" {
				logging.LOG.ErrorWithLog(Err)
				setResult.Result = "Error"
				setResult.Reason = fmt.Sprintf("Error downloading '%s' for '%s' in set '%s' - %s", file.File.Type, dbSavedItem.MediaItem.Title, dbPosterSet.PosterSetID, Err.Message)
				result.Sets = append(result.Sets, setResult)
				continue
			}
			logging.LOG.Debug(fmt.Sprintf("File '%s' for '%s' in set '%s' downloaded successfully", file.File.Type, dbSavedItem.MediaItem.Title, dbPosterSet.PosterSetID))

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
		Err = DB_InsertAllInfoIntoTables(newDBItem)
		if Err.Message != "" {
			logging.LOG.ErrorWithLog(Err)
			setResult.Result = "Error"
			setResult.Reason = fmt.Sprintf("Error updating database for set '%s' - %s", dbPosterSet.PosterSetID, Err.Message)
			result.Sets = append(result.Sets, setResult)
			continue
		}
		setResult.Result = "Success"
		setResult.Reason = fmt.Sprintf("Successfully downloaded files for '%s' in set '%s'", dbSavedItem.MediaItem.Title, dbPosterSet.PosterSetID)
		result.Sets = append(result.Sets, setResult)
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

func AutoDownload_ShouldDownloadFile(dbPosterSet DBPosterSetDetail, file PosterFile, dbSavedItem, latestMediaItem MediaItem) PosterFileWithReason {
	var psFile PosterFileWithReason
	psFile.File = file
	psFile.ReasonTitle = ""
	psFile.ReasonDetails = ""

	// Check if file was modified after last download
	fileUpdated := Time_IsLastDownloadedBeforeLatestPosterSetDate(dbPosterSet.LastDownloaded, file.Modified)

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
		CheckSeasonAdded(file.Season.Number, dbSavedItem, latestMediaItem, &psFile)
		return psFile
	case "titlecard":
		CheckEpisodeAdded(file.Episode.SeasonNumber, file.Episode.EpisodeNumber, dbSavedItem, latestMediaItem, &psFile)
	default:
		return PosterFileWithReason{}
	}
	return psFile
}

func AutoDownload_CheckMovieForKeyChanges(dbSavedItem DBMediaItemWithPosterSets) AutoDownloadResult {
	var result AutoDownloadResult
	result.MediaItemTitle = dbSavedItem.MediaItem.Title

	latestMediaItem, ratingKeyChanged, Err := AutoDownload_GetLatestMediaItemAndCheckForChanges(dbSavedItem)
	if Err.Message != "" {
		logging.LOG.ErrorWithLog(Err)
		result.OverAllResult = "Error"
		result.OverAllResultMessage = Err.Message
		return result
	}

	ratingKeyString := ""
	if ratingKeyChanged {
		ratingKeyString = fmt.Sprintf("Rating key changed from %s to %s", dbSavedItem.MediaItem.RatingKey, latestMediaItem.RatingKey)
		logging.LOG.Info(fmt.Sprintf("%s - %s", dbSavedItem.MediaItem.Title, ratingKeyString))
	}

	// Check to see if the Path has changed
	pathChanged := false
	pathChangedString := ""
	if dbSavedItem.MediaItem.Movie != nil && latestMediaItem.Movie != nil {
		if dbSavedItem.MediaItem.Movie.File.Path != latestMediaItem.Movie.File.Path {
			pathChanged = true
			pathChangedString = fmt.Sprintf("Path changed from '%s' to '%s'. ", dbSavedItem.MediaItem.Movie.File.Path, latestMediaItem.Movie.File.Path)
			logging.LOG.Info(fmt.Sprintf("%s - %s", dbSavedItem.MediaItem.Title, pathChangedString))
		}
	}

	if !ratingKeyChanged && !pathChanged {
		result.OverAllResult = "Skipped"
		result.OverAllResultMessage = fmt.Sprintf("Skipping '%s' - No changes to rating key or path", dbSavedItem.MediaItem.Title)
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
			logging.LOG.Trace(fmt.Sprintf("Skipping poster set '%s' for '%s' - No files to download for selected types", dbPosterSet.PosterSetID, dbSavedItem.MediaItem.Title))
			continue
		}

		// Download and update the media item with the new files
		for _, file := range filesToDownload {
			logging.LOG.Info(fmt.Sprintf("Downloading '%s' for '%s'", file.Type, dbSavedItem.MediaItem.Title))
			Err := CallDownloadAndUpdatePosters(latestMediaItem, file)
			if Err.Message != "" {
				logging.LOG.ErrorWithLog(Err)
			}
			logging.LOG.Debug(fmt.Sprintf("File '%s' for '%s' in set '%s' downloaded successfully", file.Type, dbSavedItem.MediaItem.Title, dbPosterSet.PosterSetID))
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
	Err = DB_InsertAllInfoIntoTables(newDBItem)
	if Err.Message != "" {
		logging.LOG.ErrorWithLog(Err)
	}

	result.OverAllResult = "Redownloaded"
	result.OverAllResultMessage = fmt.Sprintf("Redownloaded %d images for '%s' - Rating key changed from %s to %s", len(dbSavedItem.PosterSets), dbSavedItem.MediaItem.Title, dbSavedItem.MediaItem.RatingKey, latestMediaItem.RatingKey)
	logging.LOG.Info(result.OverAllResultMessage)
	return result
}

func AutoDownload_GetLatestMediaItemAndCheckForChanges(dbSavedItem DBMediaItemWithPosterSets) (MediaItem, bool, logging.StandardError) {
	// Get the latest Media Item from the Server
	latestMediaItem, Err := CallFetchItemContent(dbSavedItem.MediaItem.RatingKey, dbSavedItem.MediaItem.LibraryTitle)
	if Err.Message != "" {
		logging.LOG.ErrorWithLog(Err)
		return MediaItem{}, false, Err
	}

	// If the TMDB ID or Library Title do not match, Skip
	if dbSavedItem.MediaItem.TMDB_ID != latestMediaItem.TMDB_ID || dbSavedItem.MediaItem.LibraryTitle != latestMediaItem.LibraryTitle {
		Err.Message = fmt.Sprintf("Skipping '%s' - TMDB ID or Library Title do not match", dbSavedItem.MediaItem.Title)
		Err.Details = map[string]any{
			"MediaItem_TMDB_ID":            dbSavedItem.MediaItem.TMDB_ID,
			"LatestMediaItem_TMDB_ID":      latestMediaItem.TMDB_ID,
			"MediaItem_LibraryTitle":       dbSavedItem.MediaItem.LibraryTitle,
			"LatestMediaItem_LibraryTitle": latestMediaItem.LibraryTitle,
		}
		return MediaItem{}, false, Err
	}

	// Check to see if the Rating Key has changed
	ratingKeyChanged := dbSavedItem.MediaItem.RatingKey != latestMediaItem.RatingKey

	return latestMediaItem, ratingKeyChanged, logging.StandardError{}
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

	// Send a notification to all configured providers
	for _, provider := range Global_Config.Notifications.Providers {
		if provider.Enabled {
			switch provider.Provider {
			case "Discord":
				Notification_SendDiscordMessage(
					provider.Discord,
					messageBody,
					imageURL,
					psFile.ReasonTitle,
				)
			case "Pushover":
				Notification_SendPushoverMessage(
					provider.Pushover,
					messageBody,
					imageURL,
					psFile.ReasonTitle,
				)
			case "Gotify":
				Notification_SendGotifyMessage(
					provider.Gotify,
					messageBody,
					imageURL,
					psFile.ReasonTitle,
				)
			}
		}
	}
}
