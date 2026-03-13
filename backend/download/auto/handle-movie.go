package autodownload

import (
	"aura/cache"
	"aura/database"
	"aura/logging"
	"aura/mediaserver"
	"aura/mediux"
	"aura/models"
	"aura/utils"
	"context"
	"fmt"
	"runtime/debug"
	"strings"
	"time"
)

func handleMovie(ctx context.Context, dbItem models.DBSavedItem) (result AutoDownloadResult) {
	result = AutoDownloadResult{}
	result.Item = utils.MediaItemInfo(dbItem.MediaItem)

	defer func() {
		if r := recover(); r != nil {
			logging.LOGGER.Error().
				Timestamp().
				Str("item", utils.MediaItemInfo(dbItem.MediaItem)).
				Interface("recover", r).
				Str("stack", string(debug.Stack())).
				Msg("PANIC: in handleMovie for AutoDownload Check")
			result = AutoDownloadResult{
				Item:           utils.MediaItemInfo(dbItem.MediaItem),
				OverallResult:  "error",
				OverallMessage: fmt.Sprintf("Panic occurred: %v", r),
			}
		}
	}()

	// Get the base Movie Media Item from the cache
	_, actionGetFromCache := logging.AddSubActionToContext(ctx, fmt.Sprintf("Getting %s Item from cache", utils.MediaItemInfo(dbItem.MediaItem)), logging.LevelTrace)
	mediaItem, found := cache.LibraryStore.GetMediaItemFromSectionByTMDBID(dbItem.MediaItem.LibraryTitle, dbItem.MediaItem.TMDB_ID)
	if !found || mediaItem == nil {
		result.OverallResult = "error"
		result.OverallMessage = "Media Item not found in cache"
		actionGetFromCache.SetError("Media Item not found in cache", "Try refreshing the cache if this issue persists", nil)
		return result
	}
	actionGetFromCache.Complete()

	// Get the latest Movie Media Item from the media server
	found, Err := mediaserver.GetMediaItemDetails(ctx, mediaItem)
	if Err.Message != "" || !found {
		result.OverallResult = "error"
		result.OverallMessage = "Failed to get latest Media Item details from media server"
		return result
	}

	// If there are no sets in the database for this item, we will skip the check process as there is no existing data to compare against, and just return the result
	if len(dbItem.PosterSets) == 0 {
		result.OverallResult = "skipped"
		result.OverallMessage = "No sets in this item, try deleting and re-adding if this is an error"
		return result
	}

	// Indicators that something has changed with the media item
	// If any of these are true, we will redownload the selected types for each set
	// If none of these are true, we will compare get the latest set and check the dates to see if there has been an update to an image
	itemChanges := make(map[string]bool)
	itemChanges["rating_key"] = false
	itemChanges["duration"] = false
	itemChanges["path"] = false
	itemChanges["size"] = false

	_, actionCheckChanges := logging.AddSubActionToContext(ctx, "Checking if Media Item has changed", logging.LevelTrace)
	// Check to see if the Rating Key has changed
	if mediaItem.RatingKey != dbItem.MediaItem.RatingKey {
		itemChanges["rating_key"] = true
	}
	// Check to see if the file details have changed
	if mediaItem.Movie != nil && dbItem.MediaItem.Movie != nil {
		if durationReallyChanged(mediaItem.Movie.File.Duration, dbItem.MediaItem.Movie.File.Duration) {
			itemChanges["duration"] = true
		}
		if mediaItem.Movie.File.Path != dbItem.MediaItem.Movie.File.Path {
			itemChanges["path"] = true
		}
		if sizeReallyChanged(mediaItem.Movie.File.Size, dbItem.MediaItem.Movie.File.Size) {
			itemChanges["size"] = true
		}
	}
	actionCheckChanges.AppendResult("item_changes", itemChanges)
	actionCheckChanges.Complete()

	for _, dbSet := range dbItem.PosterSets {
		var setResult AutoDownloadSetResult
		setResult.ID = dbSet.ID
		setResult.Title = dbSet.Title
		setResult.UserCreated = dbSet.UserCreated

		// If no types are selected, then we skip
		if !dbSet.SelectedTypes.Poster && !dbSet.SelectedTypes.Backdrop {
			setResult.Result = "skipped"
			setResult.Reason = "No image types selected for this set, skipping check for this set"
			result.Sets = append(result.Sets, setResult)
			continue
		}

		mediuxSet := models.SetRef{}
		// Get the latest set details from MediUX
		switch dbSet.Type {
		case "movie":
			mediuxSet, _, Err = mediux.GetMovieSetByID(ctx, dbSet.ID, mediaItem.LibraryTitle)
			if Err.Message != "" {
				setResult.Result = "error"
				setResult.Reason = "Failed to get latest set details from MediUX"
				result.Sets = append(result.Sets, setResult)
				continue
			}
		case "collection":
			mediuxSet, _, Err = mediux.GetMovieCollectionSetByID(ctx, dbSet.ID, mediaItem.TMDB_ID, mediaItem.LibraryTitle)
			if Err.Message != "" {
				setResult.Result = "error"
				setResult.Reason = "Failed to get latest set details from MediUX"
				result.Sets = append(result.Sets, setResult)
				continue
			}
		default:
			setResult.Result = "error"
			setResult.Reason = "Unknown set type"
			result.Sets = append(result.Sets, setResult)
			continue
		}

		imagesToRedownload := []ImageFileWithReason{}

		for idx, image := range mediuxSet.Images {
			// Skip any images that are not selected for download in the set, or if the image type is not poster or backdrop
			if image.Type != "poster" && image.Type != "backdrop" {
				continue
			} else if (image.Type == "poster" && !dbSet.SelectedTypes.Poster) || (image.Type == "backdrop" && !dbSet.SelectedTypes.Backdrop) {
				continue
			}

			// Run the relevant checks based on the item changes and set updates
			if itemChanges["rating_key"] || itemChanges["duration"] || itemChanges["path"] || itemChanges["size"] {
				// If the item details have changed, we will redownload the selected types for each set
				imagesToRedownload = append(imagesToRedownload, ImageFileWithReason{ImageFile: image, ReasonTitle: "Movie Info Changed", Reason: getMovieInfoReasons(itemChanges)})
			} else {
				// If the item details have not changed, we will compare get the latest set and check the dates to see if there has been an update to an image
				imageFileCheckResult := make(map[string]any)
				imageFileCheckResult["image"] = utils.GetFileDownloadName(mediaItem.Title, image)
				imageFileCheckResult["image_type"] = image.Type
				var matchingOldImage *models.ImageFile
				for _, oldImage := range dbSet.Images {
					if oldImage.Type == image.Type && oldImage.ID == image.ID {
						matchingOldImage = &oldImage
						break
					}
				}

				if matchingOldImage == nil {
					// No matching image in the old set, we will redownload this image
					imagesToRedownload = append(imagesToRedownload, ImageFileWithReason{ImageFile: image, ReasonTitle: "New Image in Set", Reason: fmt.Sprintf("New image added to set since last download\n%s", dbSet.LastDownloaded.Format("2006-01-02 15:04:05"))})
					imageFileCheckResult["check_result"] = "No matching image in the old set, redownloading image"
				} else if !image.Modified.Equal(matchingOldImage.Modified) {
					// The dates don't match, we will check to see if the new image is newer than the old image, if it is we will redownload this image
					if image.Modified.After(matchingOldImage.Modified) {
						imagesToRedownload = append(imagesToRedownload, ImageFileWithReason{ImageFile: image,
							ReasonTitle: "Image Updated",
							Reason: fmt.Sprintf("Image modified date is newer than the previously downloaded image\nNew Image Updated: %s\nPrevious Image Updated: %s",
								image.Modified.Format("2006-01-02 15:04:05"), matchingOldImage.Modified.Format("2006-01-02 15:04:05"))})
						imageFileCheckResult["check_result"] = "Image modified date is newer than the previously downloaded image, redownloading image"
						imageFileCheckResult["image_modified_new"] = image.Modified.Format("2006-01-02 15:04:05")
						imageFileCheckResult["image_modified_previous"] = matchingOldImage.Modified.Format("2006-01-02 15:04:05")
						imageFileCheckResult["last_downloaded"] = dbSet.LastDownloaded.Format("2006-01-02 15:04:05")
					}
				} else if image.Modified.After(dbSet.LastDownloaded) {
					// The image has been modified since the last download, we will redownload this image
					imagesToRedownload = append(imagesToRedownload, ImageFileWithReason{ImageFile: image,
						ReasonTitle: "Image Updated Since Last Download",
						Reason: fmt.Sprintf("Image modified date is newer than the last downloaded date\nNew Image Updated: %s\nLast Downloaded: %s",
							image.Modified.Format("2006-01-02 15:04:05"), dbSet.LastDownloaded.Format("2006-01-02 15:04:05"))})
					imageFileCheckResult["check_result"] = "Image modified date is newer than the last downloaded date, redownloading image"
					imageFileCheckResult["image_modified_new"] = image.Modified.Format("2006-01-02 15:04:05")
					imageFileCheckResult["image_modified_previous"] = matchingOldImage.Modified.Format("2006-01-02 15:04:05")
					imageFileCheckResult["last_downloaded"] = dbSet.LastDownloaded.Format("2006-01-02 15:04:05")
				} else {
					imageFileCheckResult["check_result"] = "No changes detected for this image, skipping redownload"
					imageFileCheckResult["image_modified_new"] = image.Modified.Format("2006-01-02 15:04:05")
					imageFileCheckResult["image_modified_previous"] = matchingOldImage.Modified.Format("2006-01-02 15:04:05")
					imageFileCheckResult["last_downloaded"] = dbSet.LastDownloaded.Format("2006-01-02 15:04:05")
				}
				actionCheckChanges.AppendResult(fmt.Sprintf("image_check_%d", idx+1), imageFileCheckResult)
			}
		}
		actionCheckChanges.Complete()

		// If no images need to be redownloaded, we will skip the redownload process and move on to the next set
		if len(imagesToRedownload) == 0 {
			setResult.Result = "skipped"
			setResult.Reason = "No changes detected that require redownloading images"
			result.Sets = append(result.Sets, setResult)
			continue
		}

		_, imageRedownloadsAction := logging.AddSubActionToContext(ctx, fmt.Sprintf("Downloading %d updated images for set %s (ID: %s)", len(imagesToRedownload), dbSet.Title, dbSet.ID), logging.LevelInfo)
		for idx, image := range imagesToRedownload {
			// Redownload the image
			imageRedownloadResult := make(map[string]any)
			imageRedownloadResult["image"] = utils.GetFileDownloadName(mediaItem.Title, image.ImageFile)
			imageRedownloadResult["image_type"] = image.Type
			imageRedownloadResult["redownload_reason"] = image.Reason

			Err := mediaserver.DownloadApplyImageToMediaItem(ctx, mediaItem, image.ImageFile)
			if Err.Message != "" {
				imageRedownloadResult["redownload_result"] = "error"
				imageRedownloadResult["redownload_error"] = Err.Message
				imageRedownloadsAction.AppendResult(fmt.Sprintf("image_redownload_%d", idx+1), imageRedownloadResult)
				setResult.Result = "error"
				setResult.Reason = fmt.Sprintf("Failed to redownload image %s: %s", utils.GetFileDownloadName(mediaItem.Title, image.ImageFile), Err.Message)
				result.Sets = append(result.Sets, setResult)
				logging.LOGGER.Error().Timestamp().Str("item", utils.MediaItemInfo(*mediaItem)).Str("set_id", dbSet.ID).Str("image", utils.GetFileDownloadName(mediaItem.Title, image.ImageFile)).Str("error", Err.Message).Msg("Failed to redownload image for AutoDownload Check")
				continue
			} else {
				// Send a notification to all configured notification services
				// We do this asynchronously and don't wait for the result
				go func(image ImageFileWithReason) {
					sendFileDownloadNotification(utils.MediaItemInfo(*mediaItem), dbSet.ID, image)
				}(image)
			}
		}
		imageRedownloadsAction.Complete()

		dbItem.MediaItem = *mediaItem
		newSetInfo := models.DBPosterSetDetail{
			PosterSet: models.PosterSet{
				BaseSetInfo: models.BaseSetInfo{
					ID:               dbSet.ID,
					Type:             dbSet.Type,
					Title:            dbSet.Title,
					UserCreated:      dbSet.UserCreated,
					DateCreated:      dbSet.DateCreated,
					DateUpdated:      dbSet.DateUpdated,
					Popularity:       dbSet.Popularity,
					PopularityGlobal: dbSet.PopularityGlobal,
				},
				Images: mediuxSet.Images,
			},
			LastDownloaded: time.Now(),
			SelectedTypes:  dbSet.SelectedTypes,
			AutoDownload:   dbSet.AutoDownload,
			ToDelete:       false,
		}
		found, dbItem.PosterSets = utils.UpdatePosterSetInDBItem(dbItem.PosterSets, newSetInfo)
		if !found {
			logging.LOGGER.Error().Timestamp().Str("item", utils.MediaItemInfo(*mediaItem)).Str("set_id", dbSet.ID).Msg("Failed to update set info in DB item after redownloading images for AutoDownload Check, set not found in DB item")
		} else {
			// Update the set info in the database for this item
			Err = database.UpsertSavedItem(ctx, dbItem)
			if Err.Message != "" {
				setResult.Result = "error"
				setResult.Reason = fmt.Sprintf("Error updating database for '%s' in set '%s' - %s", utils.MediaItemInfo(dbItem.MediaItem), dbSet.ID, Err.Message)
				result.Sets = append(result.Sets, setResult)
				logging.LOGGER.Error().Timestamp().Str("item", utils.MediaItemInfo(*mediaItem)).Str("set_id", dbSet.ID).Str("error", Err.Message).Msg("Failed to update DB item after redownloading images for AutoDownload Check")
				continue
			}
		}

		setResult.Result = "success"
		setResult.Reason = fmt.Sprintf("%d images need to be redownloaded", len(imagesToRedownload))
		result.Sets = append(result.Sets, setResult)
	}

	if len(result.Sets) == 0 {
		result.OverallResult = "skipped"
		result.OverallMessage = "No sets in this item, try deleting and re-adding if this is an error"
	} else {
		errorCount := 0
		warningCount := 0
		successCount := 0
		skippedCount := 0
		for _, setResult := range result.Sets {
			switch setResult.Result {
			case "error":
				errorCount++
			case "warning":
				warningCount++
			case "success":
				successCount++
			case "skipped":
				skippedCount++
			}
		}
		if errorCount > 0 {
			result.OverallResult = "error"
			result.OverallMessage = fmt.Sprintf("%d sets had errors, %d sets were successful, %d sets were skipped", errorCount, successCount, skippedCount)
		} else if warningCount > 0 {
			result.OverallResult = "warning"
			result.OverallMessage = fmt.Sprintf("%d sets had warnings, %d sets were successful, %d sets were skipped", warningCount, successCount, skippedCount)
		} else if successCount > 0 {
			result.OverallResult = "success"
			result.OverallMessage = fmt.Sprintf("%d sets were successful, %d sets were skipped", successCount, skippedCount)
		} else {
			result.OverallResult = "skipped"
			result.OverallMessage = "All sets were skipped, no changes detected that require redownloading images"
		}
	}

	return result
}

func getMovieInfoReasons(itemChanges map[string]bool) string {
	reason := "Changes in Movie Info: "
	if itemChanges["rating_key"] {
		reason += "Rating Key"
	}
	if itemChanges["duration"] {
		if reason != "Changes in Movie Info: " {
			reason += ", "
		}
		reason += "Duration"
	}
	if itemChanges["path"] {
		if reason != "Changes in Movie Info: " {
			reason += ", "
		}
		reason += "Path"
	}
	if itemChanges["size"] {
		if reason != "Changes in Movie Info: " {
			reason += ", "
		}
		reason += "Size"
	}
	// Replace the last comma with 'and' if there are multiple reasons
	if strings.Count(reason, ",") > 0 {
		lastCommaIndex := strings.LastIndex(reason, ",")
		reason = reason[:lastCommaIndex] + " and" + reason[lastCommaIndex+1:]
	}
	return reason
}
