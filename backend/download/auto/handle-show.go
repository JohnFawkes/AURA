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
	"sort"
	"strings"
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
				Msg("PANIC: in handleShow for AutoDownload Check")
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

	// If there are no Poster Sets in the database for this item, we will skip the check process as there is no existing data to compare against, and just return the result
	if len(dbItem.PosterSets) == 0 {
		result.OverallResult = "skipped"
		result.OverallMessage = "No sets in this item, try deleting and re-adding if this is an error"
		return result
	}

	// If none of the Poster Sets are set to be auto-downloaded, we will skip the check process as there is no need to check for updates if we are not going to download anything, and just return the result
	autoDownloadSetExists := false
	for _, s := range dbItem.PosterSets {
		if s.AutoDownload {
			autoDownloadSetExists = true
			break
		}
	}
	if !autoDownloadSetExists {
		result.OverallResult = "skipped"
		result.OverallMessage = "No sets in this item are set to auto-download, try updating the item if this is an error"
		return result
	}

	// Indicators that something has changed with the media item
	// If any of these are true, we will redownload the selected types for each set
	// If none of these are true, we will compare get the latest set and check the dates to see if there has been an update to an image
	changes := ShowChangeDetails{}

	_, actionCheckChanges := logging.AddSubActionToContext(ctx, "Checking if Media Item has changed", logging.LevelTrace)
	// Rating key change
	if mediaItem.RatingKey != dbItem.MediaItem.RatingKey {
		changes.RatingKeyChanged = true
		changes.OldRatingKey = dbItem.MediaItem.RatingKey
		changes.NewRatingKey = mediaItem.RatingKey
	}
	// Path + season/episode changes
	if mediaItem.Series != nil {
		oldPath := ""
		if dbItem.MediaItem.Series != nil {
			oldPath = dbItem.MediaItem.Series.Location
		}
		newPath := mediaItem.Series.Location
		if oldPath != newPath {
			changes.PathChanged = true
			changes.OldPath = oldPath
			changes.NewPath = newPath
		}

		dbSeasons := make(map[int]models.MediaItemSeason)
		if dbItem.MediaItem.Series != nil {
			for _, s := range dbItem.MediaItem.Series.Seasons {
				dbSeasons[s.SeasonNumber] = s
			}
		}

		for _, season := range mediaItem.Series.Seasons {
			dbSeason, seasonFound := dbSeasons[season.SeasonNumber]
			if !seasonFound {
				changes.AddedSeasons = append(changes.AddedSeasons, season.SeasonNumber)
				changes.SeasonsAdded = true
				continue
			}

			dbEpisodes := make(map[int]models.MediaItemEpisode, len(dbSeason.Episodes))
			for _, e := range dbSeason.Episodes {
				dbEpisodes[e.EpisodeNumber] = e
			}

			for _, episode := range season.Episodes {
				dbEpisode, episodeFound := dbEpisodes[episode.EpisodeNumber]
				if !episodeFound {
					changes.AddedEpisodes = append(changes.AddedEpisodes, EpisodeRef{
						SeasonNumber:  episode.SeasonNumber,
						EpisodeNumber: episode.EpisodeNumber,
					})
					changes.EpisodesAdded = true
					continue
				}
				ratingKeyChanged := episode.RatingKey != dbEpisode.RatingKey
				pathChanged := episode.File.Path != dbEpisode.File.Path
				sizeChanged := sizeReallyChanged(episode.File.Size, dbEpisode.File.Size)
				durationChanged := durationReallyChanged(episode.File.Duration, dbEpisode.File.Duration)
				if ratingKeyChanged || pathChanged || sizeChanged || durationChanged {
					episodeChange := EpisodeChangeDetails{
						EpisodeRef: EpisodeRef{
							SeasonNumber:  episode.SeasonNumber,
							EpisodeNumber: episode.EpisodeNumber,
						},
					}
					if ratingKeyChanged {
						episodeChange.RatingKeyChanged = true
						episodeChange.OldRatingKey = dbEpisode.RatingKey
						episodeChange.NewRatingKey = episode.RatingKey
					}
					if pathChanged {
						episodeChange.PathChanged = true
						episodeChange.OldPath = dbEpisode.File.Path
						episodeChange.NewPath = episode.File.Path
					}
					if sizeChanged {
						episodeChange.SizeChanged = true
						episodeChange.OldSize = dbEpisode.File.Size
						episodeChange.NewSize = episode.File.Size
					}
					if durationChanged {
						episodeChange.DurationChanged = true
						episodeChange.OldDuration = dbEpisode.File.Duration
						episodeChange.NewDuration = episode.File.Duration
					}
					changes.ChangedEpisodes = append(changes.ChangedEpisodes, episodeChange)
					changes.EpisodesChanged = true
				}
			}
		}
	}

	actionCheckChanges.AppendResult("changes", changes)
	actionCheckChanges.AppendResult("changes_summary", map[string]any{
		"rating_key_changed":     changes.RatingKeyChanged,
		"path_changed":           changes.PathChanged,
		"added_seasons_count":    len(changes.AddedSeasons),
		"added_episodes_count":   len(changes.AddedEpisodes),
		"changed_episodes_count": len(changes.ChangedEpisodes),
	})
	actionCheckChanges.Complete()
	logging.Dev().Timestamp().
		Bool("rating_key_changed", changes.RatingKeyChanged).
		Bool("path_changed", changes.PathChanged).
		Int("added_seasons_count", len(changes.AddedSeasons)).
		Int("added_episodes_count", len(changes.AddedEpisodes)).
		Int("changed_episodes_count", len(changes.ChangedEpisodes)).
		Msgf("Change details for %s", utils.MediaItemInfo(dbItem.MediaItem))

	for _, dbSet := range dbItem.PosterSets {
		var setResult AutoDownloadSetResult
		setResult.ID = dbSet.ID
		setResult.Title = dbSet.Title
		setResult.UserCreated = dbSet.UserCreated

		// If the set is not set to auto-download, then we skip it
		if !dbSet.AutoDownload {
			setResult.Result = "skipped"
			setResult.Reason = "Set is not set to auto-download, skipped"
			result.Sets = append(result.Sets, setResult)
			continue
		}

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

		// Build fast lookup maps once (before looping images)
		allSeasonsInMediaItem := make(map[int]struct{})
		if mediaItem.Series != nil {
			for _, s := range mediaItem.Series.Seasons {
				allSeasonsInMediaItem[s.SeasonNumber] = struct{}{}
			}
		}

		addedSeasonSet := make(map[int]struct{}, len(changes.AddedSeasons))
		for _, s := range changes.AddedSeasons {
			addedSeasonSet[s] = struct{}{}
		}

		episodeKey := func(season, episode int) string {
			return fmt.Sprintf("%d:%d", season, episode)
		}

		allEpisodesInMediaItem := make(map[string]struct{})
		if mediaItem.Series != nil {
			for _, s := range mediaItem.Series.Seasons {
				for _, e := range s.Episodes {
					allEpisodesInMediaItem[episodeKey(s.SeasonNumber, e.EpisodeNumber)] = struct{}{}
				}
			}
		}

		addedEpisodeSet := make(map[string]struct{}, len(changes.AddedEpisodes))
		for _, ep := range changes.AddedEpisodes {
			addedEpisodeSet[episodeKey(ep.SeasonNumber, ep.EpisodeNumber)] = struct{}{}
		}

		changedEpisodeSet := make(map[string]struct{}, len(changes.ChangedEpisodes))
		for _, ep := range changes.ChangedEpisodes {
			changedEpisodeSet[episodeKey(ep.SeasonNumber, ep.EpisodeNumber)] = struct{}{}
		}

		oldImageByKey := make(map[string]models.ImageFile, len(dbSet.Images))
		for _, oldImage := range dbSet.Images {
			key := oldImage.Type + "|" + oldImage.ID
			oldImageByKey[key] = oldImage
		}

		_, actionImageChecks := logging.AddSubActionToContext(ctx, "Checking images in set", logging.LevelTrace)
		sort.SliceStable(mediuxSet.Images, func(i, j int) bool {
			a := mediuxSet.Images[i]
			b := mediuxSet.Images[j]

			typeRank := func(t string) int {
				switch t {
				case "poster":
					return 0
				case "backdrop":
					return 1
				case "season_poster":
					return 2
				case "titlecard":
					return 3
				default:
					return 4
				}
			}

			ar, br := typeRank(a.Type), typeRank(b.Type)
			if ar != br {
				return ar < br
			}

			seasonNum := func(img models.ImageFile) int {
				if img.SeasonNumber == nil {
					return -1
				}
				return *img.SeasonNumber
			}

			// For season posters/titlecards, sort by season number
			if a.Type == "season_poster" || a.Type == "titlecard" {
				as, bs := seasonNum(a), seasonNum(b)
				if as != bs {
					return as < bs
				}
			}

			// Optional: keep titlecards ordered within season
			if a.Type == "titlecard" && b.Type == "titlecard" &&
				a.EpisodeNumber != nil && b.EpisodeNumber != nil &&
				*a.EpisodeNumber != *b.EpisodeNumber {
				return *a.EpisodeNumber < *b.EpisodeNumber
			}

			// Stable fallback
			return a.ID < b.ID
		})
		// Now we loop through all the images in the latest set details and do our checks
		for _, image := range mediuxSet.Images {
			imageName := utils.GetFileDownloadName(mediaItem.Title, image)
			check := ImageCheckResult{
				Type:    image.Type,
				Outcome: "skipped",
				Reason:  "",
			}

			if image.Type == "poster" && !dbSet.SelectedTypes.Poster {
				check.Reason = "Poster not selected for this set"
				actionImageChecks.AppendResult(imageName, check)
				continue
			} else if image.Type == "backdrop" && !dbSet.SelectedTypes.Backdrop {
				check.Reason = "Backdrop not selected for this set"
				actionImageChecks.AppendResult(imageName, check)
				continue
			} else if image.Type == "season_poster" && !dbSet.SelectedTypes.SeasonPoster && !dbSet.SelectedTypes.SpecialSeasonPoster {
				check.Reason = "Season Poster not selected for this set"
				actionImageChecks.AppendResult(imageName, check)
				continue
			} else if image.Type == "titlecard" && !dbSet.SelectedTypes.Titlecard {
				check.Reason = "Titlecard not selected for this set"
				actionImageChecks.AppendResult(imageName, check)
				continue
			}

			// Skip any season posters and titlecards if there is no season or episode information in the Media Item for this image, as there is nothing to match against and download for
			if (image.Type == "season_poster" || image.Type == "titlecard") && (image.SeasonNumber == nil || (image.Type == "titlecard" && image.EpisodeNumber == nil)) {
				check.Reason = "Image has no season/episode information to match against in the Media Item, skipping"
				actionImageChecks.AppendResult(imageName, check)
				continue
			}

			// Skip any season posters if the season they are for does not exist in the Media Item
			if image.Type == "season_poster" {
				if _, seasonExists := allSeasonsInMediaItem[*image.SeasonNumber]; !seasonExists {
					continue
				}
			}

			// Skip any titlecards if the season/episode they are for does not exist in the Media Item
			if image.Type == "titlecard" {
				sn := *image.SeasonNumber
				en := *image.EpisodeNumber
				if _, seasonExists := allSeasonsInMediaItem[sn]; !seasonExists {
					continue
				}
				if _, episodeExists := allEpisodesInMediaItem[episodeKey(sn, en)]; !episodeExists {
					continue
				}
			}

			// If we got here, it means the image type is selected for this set, so we check if the image needs to be re-downloaded based on the changes we detected earlier
			// If there are changes that indicate we should re-download, we add it to the list.
			// If there are no relevant changes, we check to see if the image has changed based on the dates and add it to the list if it has
			handled := false

			if changes.RatingKeyChanged || changes.PathChanged {
				check.Outcome = "redownload"
				check.Reason = getShowInfoChangeReason(changes)
				imagesToRedownload = append(imagesToRedownload, ImageFileWithReason{
					ImageFile:   image,
					ReasonTitle: "Show Info Changed",
					Reason:      check.Reason,
				})
				handled = true
				continue
			}

			switch image.Type {
			case "poster", "backdrop":
				// For posters and backdrop, we already checked for Series Path and RatingKey changes which would impact all images
			case "season_poster":
				if _, ok := addedSeasonSet[*image.SeasonNumber]; ok {
					seasonStr := fmt.Sprintf("Season %02d added since last download", *image.SeasonNumber)
					if *image.SeasonNumber == 0 {
						seasonStr = "Special Season added since last download"
					}
					check.Outcome = "redownload"
					check.Reason = seasonStr
					imagesToRedownload = append(imagesToRedownload, ImageFileWithReason{
						ImageFile:   image,
						ReasonTitle: "Season Added",
						Reason:      check.Reason,
					})
					handled = true
				}
			case "titlecard":
				sn := *image.SeasonNumber
				en := *image.EpisodeNumber

				_, seasonAdded := addedSeasonSet[sn]
				_, episodeAdded := addedEpisodeSet[episodeKey(sn, en)]
				_, episodeChanged := changedEpisodeSet[episodeKey(sn, en)]

				if seasonAdded || episodeAdded || episodeChanged {
					reasonTitle := ""
					reason := ""
					switch {
					case seasonAdded:
						reasonTitle = "Season Added"
						reason = fmt.Sprintf("Season %02d added since last download", sn)
						if sn == 0 {
							reasonTitle = "Special Season Added"
							reason = "Special Season added since last download"
						}
					case episodeAdded:
						reasonTitle = "Episode Added"
						reason = fmt.Sprintf("Season %02d Episode %02d added since last download", sn, en)
					default:
						reasonTitle = "Episode Changed"
						reason = fmt.Sprintf("Season %02d Episode %02d changed since last download\n", sn, en)
						// Get the changes episode details to include in the reason
						for _, epChange := range changes.ChangedEpisodes {
							if epChange.SeasonNumber == sn && epChange.EpisodeNumber == en {
								reason += getEpisodeInfoChangeReason(epChange)
								break
							}
						}
					}

					imagesToRedownload = append(imagesToRedownload, ImageFileWithReason{
						ImageFile:   image,
						ReasonTitle: reasonTitle,
						Reason:      reason,
					})
					check.Outcome = "redownload"
					check.Reason = reason
					handled = true
				}
			default:
				check.Reason = "Unsupported image type"
			}

			if !handled {
				checkImageDates(image, &dbSet, oldImageByKey, &imagesToRedownload, &check)
			}
			actionImageChecks.AppendResult(imageName, check)
		}
		actionImageChecks.AppendResult("images_to_redownload_count", len(imagesToRedownload))
		actionImageChecks.Complete()

		if len(imagesToRedownload) == 0 {
			setResult.Result = "skipped"
			setResult.Reason = "No changes detected for any images in this set, skipping download"
			result.Sets = append(result.Sets, setResult)
			continue
		}
		logging.Dev().Timestamp().
			Int("total_images_in_set", len(mediuxSet.Images)).
			Int("images_to_redownload", len(imagesToRedownload)).
			Msgf("Image check results for set %s", dbSet.ID)

		_, imageRedownloadsAction := logging.AddSubActionToContext(ctx, fmt.Sprintf("Downloading %d updated images for set %s (ID: %s)", len(imagesToRedownload), dbSet.Title, dbSet.ID), logging.LevelInfo)
		for idx, image := range imagesToRedownload {
			// Redownload the image
			imageRedownloadResult := make(map[string]any)
			imageRedownloadResult["image"] = utils.GetFileDownloadName(mediaItem.Title, image.ImageFile)
			imageRedownloadResult["image_type"] = image.Type
			imageRedownloadResult["redownload_reason"] = image.Reason
			Err.Message = ""
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
					sendFileDownloadNotification(*mediaItem, dbSet, image)
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

func checkImageDates(
	image models.ImageFile,
	dbSet *models.DBPosterSetDetail,
	oldImageByKey map[string]models.ImageFile,
	imagesToRedownload *[]ImageFileWithReason,
	check *ImageCheckResult,
) {
	key := image.Type + "|" + image.ID
	matchingOldImage, found := oldImageByKey[key]

	if !found {
		check.Outcome = "redownload"
		check.Reason = fmt.Sprintf(
			"New image added to set since last download\nImage Updated New: %s\nLast Downloaded: %s",
			image.Modified.Format("2006-01-02 15:04:05"),
			dbSet.LastDownloaded.Format("2006-01-02 15:04:05"),
		)
		check.Details = map[string]any{
			"image_modified_new": image.Modified.Format("2006-01-02 15:04:05"),
			"last_downloaded":    dbSet.LastDownloaded.Format("2006-01-02 15:04:05"),
		}
		*imagesToRedownload = append(*imagesToRedownload, ImageFileWithReason{
			ImageFile:   image,
			ReasonTitle: "New Image in Set",
			Reason:      check.Reason,
		})
		return
	}

	check.Details = map[string]any{
		"image_modified_new": image.Modified.Format("2006-01-02 15:04:05"),
		"image_modified_old": matchingOldImage.Modified.Format("2006-01-02 15:04:05"),
		"last_downloaded":    dbSet.LastDownloaded.Format("2006-01-02 15:04:05"),
	}

	if !image.Modified.Equal(matchingOldImage.Modified) && image.Modified.After(matchingOldImage.Modified) {
		check.Outcome = "redownload"
		check.Reason = fmt.Sprintf(
			"Image modified date is newer than previous image\nImage Updated New: %s\nImage Updated Old: %s",
			image.Modified.Format("2006-01-02 15:04:05"),
			matchingOldImage.Modified.Format("2006-01-02 15:04:05"),
		)
		*imagesToRedownload = append(*imagesToRedownload, ImageFileWithReason{
			ImageFile:   image,
			ReasonTitle: "Image Updated",
			Reason:      check.Reason,
		})
		return
	} else if image.Modified.After(dbSet.LastDownloaded) {
		check.Outcome = "redownload"
		check.Reason = fmt.Sprintf(
			"Image modified date is newer than last downloaded date\nImage Updated New: %s\nLast Downloaded: %s",
			image.Modified.Format("2006-01-02 15:04:05"),
			dbSet.LastDownloaded.Format("2006-01-02 15:04:05"),
		)

		*imagesToRedownload = append(*imagesToRedownload, ImageFileWithReason{
			ImageFile:   image,
			ReasonTitle: "Image Updated Since Last Download",
			Reason:      check.Reason,
		})
		return
	}

	check.Outcome = "skipped"
	check.Reason = "No changes detected"
}

func getShowInfoChangeReason(changes ShowChangeDetails) string {
	reason := "Change detected in show info:"
	if changes.RatingKeyChanged {
		reason += fmt.Sprintf("\nRating Key changed from '%s' to '%s'", changes.OldRatingKey, changes.NewRatingKey)
	}
	if changes.PathChanged {
		reason += fmt.Sprintf("\nPath changed from '%s' to '%s'", changes.OldPath, changes.NewPath)
	}
	return reason
}

func getEpisodeInfoChangeReason(changes EpisodeChangeDetails) string {
	lines := []string{"Change detected in episode info:"}

	if changes.RatingKeyChanged {
		lines = append(lines, fmt.Sprintf(
			"Rating Key changed:\n- old: %s\n- new: %s",
			changes.OldRatingKey, changes.NewRatingKey,
		))
	}

	if changes.PathChanged {
		lines = append(lines, formatChangedPath(changes.OldPath, changes.NewPath))
	}

	if changes.SizeChanged {
		lines = append(lines, fmt.Sprintf(
			"Size changed:\n- old: %d\n- new: %d",
			changes.OldSize, changes.NewSize,
		))
	}

	if changes.DurationChanged {
		lines = append(lines, fmt.Sprintf(
			"Duration changed:\n- old: %d\n- new: %d",
			changes.OldDuration, changes.NewDuration,
		))
	}

	return strings.Join(lines, "\n")
}

func formatChangedPath(oldPath, newPath string) string {
	if oldPath == newPath {
		return ""
	}

	return fmt.Sprintf("Path changed:\n- old: %s\n- new: %s", oldPath, newPath)
}

type ShowChangeDetails struct {
	RatingKeyChanged bool                   `json:"rating_key_changed"`
	OldRatingKey     string                 `json:"old_rating_key,omitempty"`
	NewRatingKey     string                 `json:"new_rating_key,omitempty"`
	PathChanged      bool                   `json:"path_changed"`
	OldPath          string                 `json:"old_path,omitempty"`
	NewPath          string                 `json:"new_path,omitempty"`
	SeasonsAdded     bool                   `json:"seasons_added"`
	EpisodesAdded    bool                   `json:"episodes_added"`
	EpisodesChanged  bool                   `json:"episodes_changed"`
	AddedSeasons     []int                  `json:"added_seasons,omitempty"`
	AddedEpisodes    []EpisodeRef           `json:"added_episodes,omitempty"`
	ChangedEpisodes  []EpisodeChangeDetails `json:"changed_episodes,omitempty"`
}

type EpisodeRef struct {
	SeasonNumber  int `json:"season_number"`
	EpisodeNumber int `json:"episode_number"`
}

type EpisodeChangeDetails struct {
	EpisodeRef
	RatingKeyChanged bool   `json:"rating_key_changed"`
	OldRatingKey     string `json:"old_rating_key,omitempty"`
	NewRatingKey     string `json:"new_rating_key,omitempty"`
	PathChanged      bool   `json:"path_changed"`
	OldPath          string `json:"old_path,omitempty"`
	NewPath          string `json:"new_path,omitempty"`
	SizeChanged      bool   `json:"size_changed"`
	OldSize          int64  `json:"old_size,omitempty"`
	NewSize          int64  `json:"new_size,omitempty"`
	DurationChanged  bool   `json:"duration_changed"`
	OldDuration      int64  `json:"old_duration,omitempty"`
	NewDuration      int64  `json:"new_duration,omitempty"`
}

type ImageCheckResult struct {
	Type    string         `json:"type"`
	Outcome string         `json:"outcome"`
	Reason  string         `json:"reason,omitempty"`
	Details map[string]any `json:"details,omitempty"`
}
