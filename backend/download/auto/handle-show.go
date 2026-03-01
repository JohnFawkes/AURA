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
	"time"
)

func handleShow(ctx context.Context, dbItem models.DBSavedItem) (result AutoDownloadResult) {
	result = AutoDownloadResult{}
	result.Item = utils.MediaItemInfo(dbItem.MediaItem)

	defer func() {
		if r := recover(); r != nil {
			logging.LOGGER.Error().
				Timestamp().
				Str("item", utils.MediaItemInfo(dbItem.MediaItem)).
				Interface("recover", r).
				Str("stack", string(debug.Stack())).
				Msg("Panic in handleShow for AutoDownload Check")
			result = AutoDownloadResult{
				Item:           utils.MediaItemInfo(dbItem.MediaItem),
				OverallResult:  "error",
				OverallMessage: fmt.Sprintf("Panic occurred: %v", r),
			}
		}
	}()

	// Get the base Show Media Item from the cache
	_, actionGetFromCache := logging.AddSubActionToContext(ctx, fmt.Sprintf("Getting %s Item from cache", utils.MediaItemInfo(dbItem.MediaItem)), logging.LevelTrace)
	mediaItem, found := cache.LibraryStore.GetMediaItemFromSectionByTMDBID(dbItem.MediaItem.LibraryTitle, dbItem.MediaItem.TMDB_ID)
	if !found || mediaItem == nil {
		result.OverallResult = "error"
		result.OverallMessage = "Media Item not found in cache"
		actionGetFromCache.SetError("Media Item not found in cache", "Try refreshing the cache if this issue persists", nil)
		return result
	}
	actionGetFromCache.Complete()

	// Get the latest Show Media Item from the media server
	found, Err := mediaserver.GetMediaItemDetails(ctx, mediaItem)
	if Err.Message != "" {
		result.OverallResult = "error"
		result.OverallMessage = "Failed to get latest Media Item details from media server"
		return result
	}
	if !found {
		result.OverallResult = "error"
		result.OverallMessage = "Media Item not found on media server"
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
	addedSeasons := make(map[int]models.MediaItemSeason)
	addedEpisodes := make(map[int]map[int]models.MediaItemEpisode)
	episodesChanged := make(map[int]map[int]models.MediaItemEpisode)

	itemChanges["rating_key"] = false
	itemChanges["path"] = false
	itemChanges["seasons_added"] = false
	itemChanges["episodes_added"] = false
	itemChanges["episodes_changed"] = false

	_, actionCheckChanges := logging.AddSubActionToContext(ctx, "Checking if Media Item has changed", logging.LevelTrace)
	// Check to see if the Rating Key has changed
	if mediaItem.RatingKey != dbItem.MediaItem.RatingKey {
		itemChanges["rating_key"] = true
	}
	// Check to see if Path has changed
	if mediaItem.Series != nil && dbItem.MediaItem.Series != nil {
		if mediaItem.Series.Location != dbItem.MediaItem.Series.Location {
			itemChanges["path"] = true
		}
		dbSeasons := make(map[int]models.MediaItemSeason)
		for _, s := range dbItem.MediaItem.Series.Seasons {
			dbSeasons[s.SeasonNumber] = s
		}

		for _, season := range mediaItem.Series.Seasons {
			dbSeason, found := dbSeasons[season.SeasonNumber]
			if !found {
				itemChanges["seasons_added"] = true
				addedSeasons[season.SeasonNumber] = season
				// Add all episodes in the new season
				addedEpisodes[season.SeasonNumber] = make(map[int]models.MediaItemEpisode)
				for _, episode := range season.Episodes {
					addedEpisodes[season.SeasonNumber][episode.EpisodeNumber] = episode
					itemChanges["episodes_added"] = true
				}
				continue
			}

			dbEpisodes := make(map[int]models.MediaItemEpisode)
			for _, e := range dbSeason.Episodes {
				dbEpisodes[e.EpisodeNumber] = e
			}

			for _, episode := range season.Episodes {
				dbEpisode, found := dbEpisodes[episode.EpisodeNumber]
				if !found {
					itemChanges["episodes_added"] = true
					if addedEpisodes[season.SeasonNumber] == nil {
						addedEpisodes[season.SeasonNumber] = make(map[int]models.MediaItemEpisode)
					}
					addedEpisodes[season.SeasonNumber][episode.EpisodeNumber] = episode
					continue
				}
				if episode.File.Path != dbEpisode.File.Path ||
					episode.File.Size != dbEpisode.File.Size ||
					episode.File.Duration != dbEpisode.File.Duration {
					itemChanges["episodes_changed"] = true
					if episodesChanged[season.SeasonNumber] == nil {
						episodesChanged[season.SeasonNumber] = make(map[int]models.MediaItemEpisode)
					}
					episodesChanged[season.SeasonNumber][episode.EpisodeNumber] = episode
				}
			}
		}

	}
	actionCheckChanges.AppendResult("item_changes", itemChanges)
	actionCheckChanges.AppendResult("added_seasons", len(addedSeasons))
	actionCheckChanges.AppendResult("added_episodes", len(addedEpisodes))
	actionCheckChanges.AppendResult("episodes_changed", len(episodesChanged))
	actionCheckChanges.Complete()

	for _, dbSet := range dbItem.PosterSets {
		var setResult AutoDownloadSetResult
		setResult.ID = dbSet.ID
		setResult.Title = dbSet.Title
		setResult.UserCreated = dbSet.UserCreated

		// If no types are selected, then we skip
		if !dbSet.SelectedTypes.Poster && !dbSet.SelectedTypes.Backdrop && !dbSet.SelectedTypes.SeasonPoster && !dbSet.SelectedTypes.SpecialSeasonPoster && !dbSet.SelectedTypes.Titlecard {
			setResult.Result = "skipped"
			setResult.Reason = "No types selected for this set"
			result.Sets = append(result.Sets, setResult)
			continue
		}

		// Get the latest set details from MediUX
		mediuxSet, _, Err := mediux.GetShowSetByID(ctx, dbSet.ID, mediaItem.LibraryTitle)
		if Err.Message != "" {
			setResult.Result = "error"
			setResult.Reason = "Failed to get latest set details from MediUX"
			result.Sets = append(result.Sets, setResult)
			continue
		}

		imagesToRedownload := []ImageFileWithReason{}

		for idx, image := range mediuxSet.Images {
			imageFileCheckResult := make(map[string]any)
			imageFileCheckResult["image"] = utils.GetFileDownloadName(mediaItem.Title, image)
			imageFileCheckResult["image_type"] = image.Type

			if image.Type == "poster" && !dbSet.SelectedTypes.Poster {
				imageFileCheckResult["check_result"] = "Skipping this image as it is a poster and posters are not selected for this set"
				actionCheckChanges.AppendResult(fmt.Sprintf("image_check_%d", idx+1), imageFileCheckResult)
				continue
			} else if image.Type == "backdrop" && !dbSet.SelectedTypes.Backdrop {
				imageFileCheckResult["check_result"] = "Skipping this image as it is a backdrop and backdrops are not selected for this set"
				actionCheckChanges.AppendResult(fmt.Sprintf("image_check_%d", idx+1), imageFileCheckResult)
				continue
			} else if image.Type == "season_poster" && !dbSet.SelectedTypes.SeasonPoster && !dbSet.SelectedTypes.SpecialSeasonPoster {
				imageFileCheckResult["check_result"] = "Skipping this image as it is a season poster and season posters are not selected for this set"
				actionCheckChanges.AppendResult(fmt.Sprintf("image_check_%d", idx+1), imageFileCheckResult)
				continue
			} else if image.Type == "titlecard" && !dbSet.SelectedTypes.Titlecard {
				imageFileCheckResult["check_result"] = "Skipping this image as it is a titlecard and titlecards are not selected for this set"
				actionCheckChanges.AppendResult(fmt.Sprintf("image_check_%d", idx+1), imageFileCheckResult)
				continue
			}

			// Skip any season posters and titlecards if there is no season or episode information in the Media Item for this image, as there is nothing to match against and download for
			if (image.Type == "season_poster" || image.Type == "titlecard") && (image.SeasonNumber == nil || (image.Type == "titlecard" && image.EpisodeNumber == nil)) {
				imageFileCheckResult["check_result"] = "Skipping this image as it is a season poster or titlecard and there is no season or episode information in the Media Item for this image, so there is nothing to match against and download for"
				actionCheckChanges.AppendResult(fmt.Sprintf("image_check_%d", idx+1), imageFileCheckResult)
				continue
			}

			handled := false
			// Run the relevant checks based on the item changes and set updates
			if itemChanges["rating_key"] || itemChanges["path"] {
				// If the item details have changed, we will redownload the selected types for each set
				imagesToRedownload = append(imagesToRedownload, ImageFileWithReason{
					ImageFile:   image,
					ReasonTitle: "Show Info Changed",
					Reason:      getShowInfoReasons(itemChanges),
				})
				imageFileCheckResult["check_result"] = "Show info has changed, redownloading this image based on the detected changes in show info"
				imageFileCheckResult["item_changes"] = itemChanges
				actionCheckChanges.AppendResult(fmt.Sprintf("image_check_%d", idx+1), imageFileCheckResult)
				handled = true
			} else if itemChanges["seasons_added"] || itemChanges["episodes_added"] || itemChanges["episodes_changed"] {
				// If there are changes related to seasons or episodes, we will redownload season posters and titlecards, but not regular posters or backdrops
				// We will also check to ensure that the MediUX set contains season posters or titlecards before redownloading, as if there are no season posters or titlecards in the set or the set doesn't contain an image for the Season/Episode that was changed, then there is no need to redownload
				if image.Type == "season_poster" && itemChanges["seasons_added"] {
					if image.SeasonNumber != nil {
						// Check to see if the season number for this image matches any of the added seasons
						if _, found := addedSeasons[*image.SeasonNumber]; found {
							seasonStr := fmt.Sprintf("Season %d added since last download", *image.SeasonNumber)
							if *image.SeasonNumber == 0 {
								seasonStr = "Special Season added since last download"
							}
							imagesToRedownload = append(imagesToRedownload, ImageFileWithReason{
								ImageFile:   image,
								ReasonTitle: "Season Added",
								Reason:      seasonStr,
							})
							imageFileCheckResult["check_result"] = "Season added since last download, redownloading this season poster based on the detected season added"
							imageFileCheckResult["added_seasons"] = addedSeasons
							actionCheckChanges.AppendResult(fmt.Sprintf("image_check_%d", idx+1), imageFileCheckResult)
							handled = true
						}
					}
				} else if image.Type == "titlecard" && itemChanges["episodes_added"] {
					if image.SeasonNumber != nil && image.EpisodeNumber != nil {
						// Check to see if the season and episode number for this image matches any of the added episodes
						if seasonEpisodes, found := addedEpisodes[*image.SeasonNumber]; found {
							if _, found := seasonEpisodes[*image.EpisodeNumber]; found {
								imagesToRedownload = append(imagesToRedownload, ImageFileWithReason{
									ImageFile:   image,
									ReasonTitle: "Episode Added",
									Reason:      "New episode added since last download",
								})
								imageFileCheckResult["check_result"] = "Episode added since last download, redownloading this titlecard based on the detected episode added"
								imageFileCheckResult["added_episodes"] = addedEpisodes
								actionCheckChanges.AppendResult(fmt.Sprintf("image_check_%d", idx+1), imageFileCheckResult)
								handled = true
							}
						}
					}
				} else if image.Type == "titlecard" && itemChanges["episodes_changed"] {
					if image.SeasonNumber != nil && image.EpisodeNumber != nil {
						// Check to see if the season and episode number for this image matches any of the changed episodes
						if seasonEpisodes, found := episodesChanged[*image.SeasonNumber]; found {
							if _, found := seasonEpisodes[*image.EpisodeNumber]; found {
								episodeStr := fmt.Sprintf("Season %d Episode %d", *image.SeasonNumber, *image.EpisodeNumber)
								imagesToRedownload = append(imagesToRedownload, ImageFileWithReason{
									ImageFile:   image,
									ReasonTitle: "Episode Changed",
									Reason:      fmt.Sprintf("Episode changed since last download\n%s", episodeStr),
								})
								imageFileCheckResult["check_result"] = "Episode changed since last download, redownloading this titlecard based on the detected episode change"
								imageFileCheckResult["episodes_changed"] = episodesChanged
								actionCheckChanges.AppendResult(fmt.Sprintf("image_check_%d", idx+1), imageFileCheckResult)
								handled = true
							}
						}
					}
				}

			}
			if !handled {
				checkImageDate(idx, image, mediaItem, &dbSet, actionCheckChanges, &imagesToRedownload)
				continue
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
					sendFileDownloadNotification(mediaItem.Title, dbSet.ID, image)
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

func getShowInfoReasons(itemChanges map[string]bool) string {
	reason := "Changed in Show Info: "
	if itemChanges["rating_key"] {
		reason += "Rating Key"
	}
	if itemChanges["path"] {
		if reason != "Changed in Show Info: " {
			reason += " and "
		}
		reason += "Path"
	}
	return reason
}

func checkImageDate(idx int, image models.ImageFile, mediaItem *models.MediaItem, dbSet *models.DBPosterSetDetail, actionCheckChanges *logging.LogAction, imagesToRedownload *[]ImageFileWithReason) {
	// If the item details have not changed, we will compare get the latest set and check the dates to see if there has been an update to an image
	// But first we check to make sure that if it is a season_poster or titlecard, that the Media Item has the needed season/episode details for this image, if not we skip as there is nothing to download against
	imageFileCheckResult := make(map[string]any)
	imageFileCheckResult["image"] = utils.GetFileDownloadName(mediaItem.Title, image)
	imageFileCheckResult["image_type"] = image.Type

	if (image.Type == "season_poster" && image.SeasonNumber != nil && !seasonExists(*mediaItem, *image.SeasonNumber)) ||
		(image.Type == "titlecard" && image.SeasonNumber != nil && image.EpisodeNumber != nil && !episodeExists(*mediaItem, *image.SeasonNumber, *image.EpisodeNumber)) {
		imageFileCheckResult["check_result"] = "Skipping date checks for this image as the Media Item does not contain the needed season/episode details for this image, so there is nothing to match against and download for"
		actionCheckChanges.AppendResult(fmt.Sprintf("image_check_%d", idx+1), imageFileCheckResult)
		return
	}

	var matchingOldImage *models.ImageFile
	for _, oldImage := range dbSet.Images {
		if oldImage.Type == image.Type && oldImage.ID == image.ID {
			matchingOldImage = &oldImage
			break
		}
	}
	// If there is no matching image in the old set, we will redownload the image
	if matchingOldImage == nil {
		*imagesToRedownload = append(*imagesToRedownload, ImageFileWithReason{ImageFile: image,
			ReasonTitle: "New Image in Set",
			Reason: fmt.Sprintf("New image added to set since last download\nImage Updated New: %s\nLast Downloaded: %s",
				image.Modified.Format("2006-01-02 15:04:05"), dbSet.LastDownloaded.Format("2006-01-02 15:04:05"))})
		imageFileCheckResult["check_result"] = "No matching image in the old set, redownloading image"
	} else if !image.Modified.Equal(matchingOldImage.Modified) {
		// The dates don't match, we will check to see if the new image is newer than the old image, if it is we will redownload this image
		if image.Modified.After(matchingOldImage.Modified) {
			*imagesToRedownload = append(*imagesToRedownload, ImageFileWithReason{ImageFile: image,
				ReasonTitle: "Image Updated",
				Reason: fmt.Sprintf("Image modified date is newer than the previously downloaded image\nImage Updated New: %s\nImage Updated Old: %s",
					image.Modified.Format("2006-01-02 15:04:05"), matchingOldImage.Modified.Format("2006-01-02 15:04:05"))})
			imageFileCheckResult["check_result"] = "Image modified date is newer than the previously downloaded image, redownloading image"
			imageFileCheckResult["image_modified_new"] = image.Modified.Format("2006-01-02 15:04:05")
			imageFileCheckResult["image_modified_previous"] = matchingOldImage.Modified.Format("2006-01-02 15:04:05")
			imageFileCheckResult["last_downloaded"] = dbSet.LastDownloaded.Format("2006-01-02 15:04:05")
		}
	} else if image.Modified.After(dbSet.LastDownloaded) {
		// The image has been modified since the last download, we will redownload this image
		*imagesToRedownload = append(*imagesToRedownload, ImageFileWithReason{ImageFile: image,
			ReasonTitle: "Image Updated Since Last Download",
			Reason: fmt.Sprintf("Image modified date is newer than the last downloaded date\nImage Updated New: %s\nLast Downloaded: %s",
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
