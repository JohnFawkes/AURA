package download

import (
	"aura/internal/config"
	"aura/internal/database"
	"aura/internal/logging"
	"aura/internal/mediux"
	"aura/internal/modals"
	"aura/internal/notifications"
	mediaserver "aura/internal/server"
	mediaserver_shared "aura/internal/server/shared"
	"fmt"
	"strings"
	"time"
)

func CheckForUpdatesToPosters() {
	// Get all items from the database
	dbSavedItems, Err := database.GetAllItemsFromDatabase()
	if Err.Message != "" {
		logging.LOG.ErrorWithLog(Err)
		return
	}

	// Loop through each item in the database
	for _, dbSavedItem := range dbSavedItems {
		CheckItemForAutodownload(dbSavedItem)
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

func CheckItemForAutodownload(dbSavedItem modals.DBMediaItemWithPosterSets) AutoDownloadResult {
	var result AutoDownloadResult
	result.MediaItemTitle = dbSavedItem.MediaItem.Title

	// If the is not a show skip it
	if dbSavedItem.MediaItem.Type != "show" {
		result.OverAllResult = "Skipped"
		result.OverAllResultMessage = fmt.Sprintf("Skipping '%s' - Not a show", dbSavedItem.MediaItem.Title)
		return result
	}

	var mediaServer mediaserver_shared.MediaServer
	switch config.Global.MediaServer.Type {
	case "Plex":
		mediaServer = &mediaserver_shared.PlexServer{}
	case "Emby", "Jellyfin":
		mediaServer = &mediaserver_shared.EmbyJellyServer{}
	default:
		logging.LOG.Error(fmt.Sprintf("Unsupported media server type: %s", config.Global.MediaServer.Type))
		result.OverAllResult = "Error"
		result.OverAllResultMessage = fmt.Sprintf("Unsupported media server type: %s", config.Global.MediaServer.Type)
		return result
	}

	logging.LOG.Debug(fmt.Sprintf("Checking for updates to posters for '%s'", dbSavedItem.MediaItem.Title))

	// Get the latest media item from the media server using the rating key
	latestMediaItem, Err := mediaServer.FetchItemContent(dbSavedItem.MediaItem.RatingKey, dbSavedItem.MediaItem.LibraryTitle)
	if Err.Message != "" {
		logging.LOG.ErrorWithLog(Err)
		result.OverAllResult = "Error"
		result.OverAllResultMessage = "Error fetching latest media item"
		return result
	}

	result.MediaItemTitle = dbSavedItem.MediaItem.Title

	// Loop through each poster set for the item
	for _, dbPosterSet := range dbSavedItem.PosterSets {
		var setResult AutoDownloadSetResult
		setResult.PosterSetID = dbPosterSet.PosterSetID
		// If the poster set is not auto download, skip it
		if !dbPosterSet.AutoDownload {
			setResult.Result = "Skipped"
			setResult.Reason = "Auto download is not selected"
			result.Sets = append(result.Sets, setResult)
			logging.LOG.Trace(fmt.Sprintf("Skipping poster set '%s' for '%s' - Auto download is not selected", dbPosterSet.PosterSetID, dbSavedItem.MediaItem.Title))
			continue
		}
		// If selected types are empty, skip it
		if len(dbPosterSet.SelectedTypes) == 0 {
			setResult.Result = "Skipped"
			setResult.Reason = "No selected types"
			result.Sets = append(result.Sets, setResult)
			logging.LOG.Trace(fmt.Sprintf("Skipping poster set '%s' for '%s' - No selected types", dbPosterSet.PosterSetID, dbSavedItem.MediaItem.Title))
			continue
		}

		// Get the TMDB ID from the media item
		tmdbID := ""
		for _, guid := range dbSavedItem.MediaItem.Guids {
			if guid.Provider == "tmdb" {
				tmdbID = guid.ID
				break
			}
		}
		// If the TMDB ID is not found, skip the item
		if tmdbID == "" {
			logging.LOG.Error(fmt.Sprintf("TMDB ID not found for '%s'", dbSavedItem.MediaItem.Title))
			setResult.Result = "Error"
			setResult.Reason = "No TMDB ID found"
			result.Sets = append(result.Sets, setResult)
			continue
		}

		logging.LOG.Trace(fmt.Sprintf("Checking poster set '%s' for '%s' with TMDB ID '%s'", dbPosterSet.PosterSetID, dbSavedItem.MediaItem.Title, tmdbID))

		// Fetch the updated set from Mediux using the TMDB ID
		updatedSet, Err := mediux.FetchShowSetByID(dbSavedItem.MediaItem.LibraryTitle, dbSavedItem.MediaItemID, dbPosterSet.PosterSetID)
		if Err.Message != "" {
			logging.LOG.ErrorWithLog(Err)
			setResult.Result = "Error"
			setResult.Reason = fmt.Sprintf("Error fetching updated set - %s", Err.Message)
			result.Sets = append(result.Sets, setResult)
			continue
		}
		posterSetUpdated := compareLastUpdateToUpdateSetDateUpdated(dbPosterSet.LastDownloaded, updatedSet.DateUpdated)
		var formattedLastDownloaded string
		lastDownloadedTime, err := time.Parse("2006-01-02T15:04:05Z07:00", dbPosterSet.LastDownloaded)
		if err == nil {
			formattedLastDownloaded = lastDownloadedTime.Format("2006-01-02 15:04:05")
		} else {
			formattedLastDownloaded = dbPosterSet.LastDownloaded
		}

		addedSeasonOrEpisodes := AddedMoreSeasonsOrEpisodes(dbSavedItem.MediaItem, latestMediaItem)
		if !posterSetUpdated && !addedSeasonOrEpisodes {
			logging.LOG.Debug(fmt.Sprintf("Skipping '%s' - Set '%s' | Last update: %s < Last download: %s | No new seasons or episodes added",
				dbSavedItem.MediaItem.Title,
				dbPosterSet.PosterSetID,
				updatedSet.DateUpdated.Format("2006-01-02 15:04:05"),
				formattedLastDownloaded))
			setResult.Result = "Skipped"
			setResult.Reason = fmt.Sprintf("No updates since last download on %s", lastDownloadedTime.Local().Format("1/2/2006 at 3:04:05 PM"))
			result.Sets = append(result.Sets, setResult)
			continue
		}
		updateReasons := []string{}
		if posterSetUpdated {
			updateReasons = append(updateReasons, "Poster set updated")
		}
		if addedSeasonOrEpisodes {
			updateReasons = append(updateReasons, "New seasons or episodes added")
		}
		updateReason := strings.Join(updateReasons, " and ")
		logging.LOG.Debug(fmt.Sprintf("'%s' - Set '%s' updated. %s", dbSavedItem.MediaItem.Title, dbPosterSet.PosterSetID, updateReason))

		// Check if selectedTypes contains "poster"
		posterSet := false
		backdropSet := false
		seasonSet := false
		specialSeasonSet := false
		titlecardSet := false
		for _, selectedType := range dbPosterSet.SelectedTypes {
			switch selectedType {
			case "poster":
				posterSet = true
			case "backdrop":
				backdropSet = true
			case "seasonPoster":
				seasonSet = true
			case "specialSeasonPoster":
				specialSeasonSet = true
			case "titlecard":
				titlecardSet = true
			}
		}
		logging.LOG.Debug(fmt.Sprintf("Downloading selected types: %s", strings.Join(dbPosterSet.SelectedTypes, ", ")))

		filesToDownload := []modals.PosterFile{}
		if posterSet {
			if shouldDownloadFile(dbPosterSet, *updatedSet.Poster, dbSavedItem.MediaItem, latestMediaItem) {
				filesToDownload = append(filesToDownload, *updatedSet.Poster)
			}
		}
		if backdropSet {
			if shouldDownloadFile(dbPosterSet, *updatedSet.Backdrop, dbSavedItem.MediaItem, latestMediaItem) {
				filesToDownload = append(filesToDownload, *updatedSet.Backdrop)
			}
		}
		if seasonSet {
			for _, season := range updatedSet.SeasonPosters {
				if shouldDownloadFile(dbPosterSet, season, dbSavedItem.MediaItem, latestMediaItem) {
					filesToDownload = append(filesToDownload, season)
				}
			}
		}
		if specialSeasonSet {
			for _, season := range updatedSet.SeasonPosters {
				if season.Season.Number != 0 {
					continue
				}
				if shouldDownloadFile(dbPosterSet, season, dbSavedItem.MediaItem, latestMediaItem) {
					filesToDownload = append(filesToDownload, season)
				}
			}
		}
		if titlecardSet {
			for _, titlecard := range updatedSet.TitleCards {
				if shouldDownloadFile(dbPosterSet, titlecard, dbSavedItem.MediaItem, latestMediaItem) {
					filesToDownload = append(filesToDownload, titlecard)
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

		for _, file := range filesToDownload {
			logging.LOG.Info(fmt.Sprintf("Downloading new '%s' for '%s' in set '%s'", file.Type, dbSavedItem.MediaItem.Title, dbPosterSet.PosterSetID))
			logging.LOG.Debug(fmt.Sprintf("File modified: %s | Last download: %s",
				file.Modified.Format("2006-01-02 15:04:05"),
				dbPosterSet.LastDownloaded,
			))

			Err = mediaServer.DownloadAndUpdatePosters(latestMediaItem, file)
			if Err.Message != "" {
				logging.LOG.ErrorWithLog(Err)
				setResult.Result = "Error"
				setResult.Reason = fmt.Sprintf("Error downloading '%s' for '%s' in set '%s' - %s", file.Type, dbSavedItem.MediaItem.Title, dbPosterSet.PosterSetID, Err.Message)
				result.Sets = append(result.Sets, setResult)
				continue
			}
			logging.LOG.Debug(fmt.Sprintf("File '%s' for '%s' in set '%s' downloaded successfully", file.Type, dbSavedItem.MediaItem.Title, dbPosterSet.PosterSetID))

			// Send a notification to all configured providers
			// Do this using go func so that it runs asynchronously
			go func() {
				for _, provider := range config.Global.Notifications.Providers {
					if provider.Enabled {
						switch provider.Provider {
						case "Discord":
							notifications.SendDiscordNotification(
								provider.Discord,
								fmt.Sprintf(
									"%s has been updated for %s in set %s",
									mediaserver.GetFileDownloadName(file),
									dbSavedItem.MediaItem.Title,
									dbPosterSet.PosterSetID,
								),
								fmt.Sprintf("%s/%s?v=%s&key=jpg",
									"https://images.mediux.io/assets",
									file.ID,
									file.Modified.Format("20060102150405"),
								),
								"Image Updated",
							)
						case "Pushover":
							notifications.SendPushoverNotification(
								provider.Pushover,
								fmt.Sprintf(
									"%s has been updated for %s in set %s",
									mediaserver.GetFileDownloadName(file),
									dbSavedItem.MediaItem.Title,
									dbPosterSet.PosterSetID,
								),
								fmt.Sprintf("%s/%s?v=%s&key=jpg",
									"https://images.mediux.io/assets",
									file.ID,
									file.Modified.Format("20060102150405"),
								),
								"Image Updated",
							)
						}
					}
				}
			}()
		}

		// Update the item in the database with the new info
		dbSaveItem := modals.DBSavedItem{
			MediaItemID:   dbSavedItem.MediaItemID,
			MediaItem:     latestMediaItem,
			PosterSetID:   updatedSet.ID,
			PosterSet:     updatedSet,
			SelectedTypes: dbPosterSet.SelectedTypes,
			AutoDownload:  dbPosterSet.AutoDownload,
		}
		Err = database.UpdateItemInDatabase(dbSaveItem)
		if Err.Message != "" {
			logging.LOG.ErrorWithLog(Err)
			setResult.Result = "Error"
			setResult.Reason = fmt.Sprintf("Error updating database for '%s' - %s", dbSavedItem.MediaItem.Title, Err.Message)
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

func shouldDownloadFile(dbPosterSet modals.DBPosterSetDetail, file modals.PosterFile, dbSavedItem, latestMediaItem modals.MediaItem) bool {
	// Check if file was modified after last download
	fileUpdated := compareLastUpdateToUpdateSetDateUpdated(dbPosterSet.LastDownloaded, file.Modified)

	// If the file was updated, download it
	if fileUpdated {
		return true
	}

	// For season posters and titlecards, also check if new seasons/episodes were added
	switch file.Type {
	case "seasonPoster":
		return fileUpdated || CheckSeasonAdded(file.Season.Number, dbSavedItem, latestMediaItem)
	case "titlecard":
		return fileUpdated || CheckEpisodeAdded(file.Episode.SeasonNumber, file.Episode.EpisodeNumber, dbSavedItem, latestMediaItem)
	default:
		return fileUpdated
	}
}
