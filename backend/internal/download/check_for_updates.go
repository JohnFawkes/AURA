package download

import (
	"fmt"
	"poster-setter/internal/config"
	"poster-setter/internal/database"
	"poster-setter/internal/logging"
	"poster-setter/internal/mediux"
	"poster-setter/internal/modals"
	"poster-setter/internal/notifications"
	mediaserver "poster-setter/internal/server"
	mediaserver_shared "poster-setter/internal/server/shared"
	"time"
)

func CheckForUpdatesToPosters() {
	dbSets, logErr := database.GetAllItemsFromDatabase()
	if logErr.Err != nil {
		logging.LOG.ErrorWithLog(logErr)
		return
	}

	for _, dbSet := range dbSets {
		for _, dbPosterSet := range dbSet.Sets {
			if !dbPosterSet.AutoDownload {
				continue
			}
			logging.LOG.Debug(fmt.Sprintf("Checking for updates to posters for '%s' on set '%s'", dbSet.MediaItem.Title, dbPosterSet.Set.ID))
			var updatedSet modals.PosterSet
			if dbPosterSet.Set.Type == "movie" || dbPosterSet.Set.Type == "collection" || dbPosterSet.Set.Type == "show" {
				tmdbID := ""
				for _, guid := range dbSet.MediaItem.Guids {
					if guid.Provider == "tmdb" {
						tmdbID = guid.ID
						break
					}
				}
				if tmdbID == "" {
					logging.LOG.Error(fmt.Sprintf("TMDB ID not found for '%s'", dbSet.MediaItem.Title))
					continue
				}
				updatedSet, logErr = mediux.FetchSetByID(dbPosterSet.Set, tmdbID)
				if logErr.Err != nil {
					logging.LOG.ErrorWithLog(logErr)
					continue
				}
			} else {
				logging.LOG.Error(fmt.Sprintf("Set '%s' for '%s' is not a valid type: %s", dbPosterSet.Set.ID, dbSet.MediaItem.Title, dbPosterSet.Set.Type))
				continue
			}

			updated := compareLastUpdateToUpdateSetDateUpdated(dbPosterSet.LastUpdate, updatedSet.DateUpdated)
			updatedSetDateUpdated := updatedSet.DateUpdated.Format("2006-01-02 15:04:05")
			dbPosterSetLastUpdateTime, err := time.Parse("2006-01-02T15:04:05Z07:00", dbPosterSet.LastUpdate)
			dbPosterSetDateUpdated := dbPosterSet.LastUpdate
			if err == nil {
				dbPosterSetDateUpdated = dbPosterSetLastUpdateTime.Format("2006-01-02 15:04:05")
			}
			if updated {
				logging.LOG.Trace(fmt.Sprintf("'%s' - Set '%s' updated. Downloading new images...", dbSet.MediaItem.Title, dbPosterSet.Set.ID))
				// Filter and sort the files based on the selected types
				updatedSet.Files = mediaserver_shared.FilterAndSortFiles(updatedSet.Files, dbPosterSet.SelectedTypes)

				// Go through each file and download it if it has been updated
				for _, file := range updatedSet.Files {
					fileUpdated := compareLastUpdateToUpdateSetDateUpdated(dbPosterSet.LastUpdate, file.Modified)
					if !fileUpdated {
						logging.LOG.Trace(fmt.Sprintf("File '%s' for '%s' in set '%s' has not been updated. Skipping...", file.Type, dbSet.MediaItem.Title, dbPosterSet.Set.ID))
						continue
					}
					logging.LOG.Info(fmt.Sprintf("Downloading new '%s' for '%s' in set '%s'", file.Type, dbSet.MediaItem.Title, dbPosterSet.Set.ID))
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
					logErr = mediaServer.DownloadAndUpdatePosters(dbSet.MediaItem, file)
					if logErr.Err != nil {
						logging.LOG.ErrorWithLog(logErr)
						continue
					}
					logging.LOG.Debug(fmt.Sprintf("File '%s' for '%s' in set '%s' downloaded successfully", file.Type, dbSet.MediaItem.Title, dbPosterSet.Set.ID))

					// Send a notification with the following information:
					// - 'File Type' has been updated for 'Media Item' in 'Set'
					// - Image URL
					notifications.SendDiscordNotification(
						fmt.Sprintf(
							"%s has been updated for %s in set %s",
							mediaserver.GetFileDownloadName(file),
							dbSet.MediaItem.Title,
							dbPosterSet.Set.ID,
						),
						fmt.Sprintf("%s/%s?%s",
							"https://staged.mediux.io/assets",
							file.ID,
							file.Modified.Format("20060102"),
						),
					)
				}

				// Update the item in the database with the new info
				logErr = database.UpdatePosterSetInDatabase(dbPosterSet.Set, dbSet.MediaItem.RatingKey, dbPosterSet.SelectedTypes, dbPosterSet.AutoDownload)
				if logErr.Err != nil {
					logging.LOG.ErrorWithLog(logErr)
					continue
				}
			} else {
				logging.LOG.Trace(fmt.Sprintf("Skipping '%s' - Set '%s' (No Update). Last update: %s, Last download: %s", dbSet.MediaItem.Title, dbPosterSet.Set.ID, updatedSetDateUpdated, dbPosterSetDateUpdated))
				continue
			}
		}
	}
}
