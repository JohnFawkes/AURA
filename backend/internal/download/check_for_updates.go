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
	dbSavedItems, logErr := database.GetAllItemsFromDatabase()
	if logErr.Err != nil {
		logging.LOG.ErrorWithLog(logErr)
		return
	}

	// Loop through each item in the database
	for _, dbSavedItem := range dbSavedItems {
		// If the is not a show skip it
		if dbSavedItem.MediaItem.Type != "show" {
			continue
		}

		var mediaServer mediaserver_shared.MediaServer
		switch config.Global.MediaServer.Type {
		case "Plex":
			mediaServer = &mediaserver_shared.PlexServer{}
		case "Emby", "Jellyfin":
			mediaServer = &mediaserver_shared.EmbyJellyServer{}
		default:
			logErr := logging.ErrorLog{
				Err: fmt.Errorf("unsupported media server type: %s", config.Global.MediaServer.Type),
				Log: logging.Log{Message: fmt.Sprintf("Unsupported media server type: %s", config.Global.MediaServer.Type)},
			}
			logging.LOG.ErrorWithLog(logErr)
			continue
		}

		// Loop through each poster set for the item
		for _, dbPosterSet := range dbSavedItem.PosterSets {
			// If the poster set is not auto download, skip it
			if !dbPosterSet.AutoDownload {
				continue
			}
			// If selected types are empty, skip it
			if len(dbPosterSet.SelectedTypes) == 0 {
				continue
			}
			logging.LOG.Debug(fmt.Sprintf("Checking for updates to posters for '%s' on set '%s'", dbSavedItem.MediaItem.Title, dbPosterSet.PosterSetID))

			// Get the latest media item from the media server using the rating key
			latestMediaItem, logErr := mediaServer.FetchItemContent(dbSavedItem.MediaItem.RatingKey, dbSavedItem.MediaItem.LibraryTitle)
			if logErr.Err != nil {
				logging.LOG.ErrorWithLog(logErr)
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
				continue
			}
			// Fetch the updated set from Mediux using the TMDB ID
			updatedSet, logErr := mediux.FetchShowSetByID(dbPosterSet.PosterSetID)
			if logErr.Err != nil {
				logging.LOG.ErrorWithLog(logErr)
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
			logging.LOG.Trace(fmt.Sprintf("'%s' - Set '%s' updated. %s", dbSavedItem.MediaItem.Title, dbPosterSet.PosterSetID, updateReason))

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

			for _, file := range filesToDownload {
				logging.LOG.Info(fmt.Sprintf("Downloading new '%s' for '%s' in set '%s'", file.Type, dbSavedItem.MediaItem.Title, dbPosterSet.PosterSetID))

				logErr = mediaServer.DownloadAndUpdatePosters(latestMediaItem, file)
				if logErr.Err != nil {
					logging.LOG.ErrorWithLog(logErr)
					continue
				}
				logging.LOG.Debug(fmt.Sprintf("File '%s' for '%s' in set '%s' downloaded successfully", file.Type, dbSavedItem.MediaItem.Title, dbPosterSet.PosterSetID))

				// Send a notification with the following information:
				// - 'File Type' has been updated for 'Media Item' in 'Set'
				// - Image URL
				notifications.SendDiscordNotification(
					fmt.Sprintf(
						"%s has been updated for %s in set %s",
						mediaserver.GetFileDownloadName(file),
						dbSavedItem.MediaItem.Title,
						dbPosterSet.PosterSetID,
					),
					fmt.Sprintf("%s/%s?%s",
						"https://staged.mediux.io/assets",
						file.ID,
						file.Modified.Format("20060102"),
					),
				)
			}

			// Update the item in the database with the new info
			dbSaveItem := modals.DBSavedItem{
				MediaItemID:    dbSavedItem.MediaItemID,
				MediaItem:      latestMediaItem,
				PosterSetID:    updatedSet.ID,
				PosterSet:      updatedSet,
				LastDownloaded: time.Now().Format("2006-01-02T15:04:05Z07:00"),
				SelectedTypes:  dbPosterSet.SelectedTypes,
				AutoDownload:   dbPosterSet.AutoDownload,
			}
			logErr = database.UpdateItemInDatabase(dbSaveItem)
			if logErr.Err != nil {
				logging.LOG.ErrorWithLog(logErr)
				continue
			}
		}
	}
}

func shouldDownloadFile(dbPosterSet modals.DBPosterSetDetail, file modals.PosterFile, dbSavedItem, latestMediaItem modals.MediaItem) bool {
	// Check if file was modified after last download
	fileUpdated := compareLastUpdateToUpdateSetDateUpdated(dbPosterSet.LastDownloaded, file.Modified)

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
