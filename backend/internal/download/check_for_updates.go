package download

import (
	"fmt"
	"poster-setter/internal/database"
	"poster-setter/internal/logging"
	"poster-setter/internal/mediux"
	"poster-setter/internal/modals"
	"poster-setter/internal/server/plex"
	mediaserver_shared "poster-setter/internal/server/shared"
)

func CheckForUpdatesToPosters() {

	items, logErr := database.GetAllItemsFromDatabase()
	if logErr.Err != nil {
		logging.LOG.ErrorWithLog(logErr)
		return
	}

	for _, item := range items {
		if item.AutoDownload {
			logging.LOG.Debug(fmt.Sprintf("Checking for updates to posters for '%s'", item.MediaItem.Title))
			var updatedSet modals.PosterSet
			var logErr logging.ErrorLog
			if item.Set.Type == "movie" || item.Set.Type == "collection" || item.Set.Type == "show" {
				// Get the TMDB ID from the MediaItem.Guids
				tmdbID := ""
				for _, guid := range item.MediaItem.Guids {
					if guid.Provider == "tmdb" {
						tmdbID = guid.ID
						break
					}
				}
				if tmdbID == "" {
					logging.LOG.Error(fmt.Sprintf("TMDB ID not found for '%s'", item.MediaItem.Title))
					continue
				}
				updatedSet, logErr = mediux.FetchSetByID(item.Set, tmdbID)
			} else {
				logging.LOG.Error(fmt.Sprintf("Set for '%s' is not a valid type: %s", item.MediaItem.Title, item.Set.Type))
			}
			if logErr.Err != nil {
				logging.LOG.ErrorWithLog(logErr)
				continue
			}

			updated := compareLastUpdateToUpdateSetDateUpdated(item.LastUpdate, updatedSet.DateUpdated)
			if updated {
				logging.LOG.Info(fmt.Sprintf("Posters for '%s' have been updated. Downloading new posters...", item.MediaItem.Title))
				// Download the new posters and update Media Server
				item.Set.Files = mediaserver_shared.FilterAndSortFiles(updatedSet.Files, item.SelectedTypes)
				for _, file := range item.Set.Files {
					fileUpdated := compareLastUpdateToUpdateSetDateUpdated(item.LastUpdate, file.Modified)
					if !fileUpdated {
						logging.LOG.Debug(fmt.Sprintf("File '%s' for '%s' has not been updated. Skipping...", file.Type, item.MediaItem.Title))
						continue
					}
					logging.LOG.Info(fmt.Sprintf("Downloading new '%s' for '%s'", file.Type, item.MediaItem.Title))
					logErr := plex.DownloadAndUpdatePosters(item.MediaItem, file)
					if logErr.Err != nil {
						logging.LOG.ErrorWithLog(logErr)
						continue
					}
				}
				// Update the item in the database with the new info
				logErr = database.UpdateAutoDownloadItem(modals.ClientMessage{
					MediaItem:     item.MediaItem,
					Set:           updatedSet,
					AutoDownload:  item.AutoDownload,
					SelectedTypes: item.SelectedTypes,
				})
				if logErr.Err != nil {
					logging.LOG.ErrorWithLog(logErr)
					continue
				}
			} else {
				logging.LOG.Debug(fmt.Sprintf("Posters for '%s' have not been updated. Skipping...", item.MediaItem.Title))
			}
		}
	}
}
