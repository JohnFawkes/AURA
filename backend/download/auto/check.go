package autodownload

import (
	"aura/database"
	"aura/logging"
	"aura/mediaserver"
	"aura/models"
	"aura/utils"
	"context"
	"fmt"
)

type AutoDownloadResult struct {
	Item           string                  `json:"item"`
	Sets           []AutoDownloadSetResult `json:"sets"`
	OverallResult  string                  `json:"overall_result"`
	OverallMessage string                  `json:"overall_message"`
}

type AutoDownloadSetResult struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	UserCreated string `json:"user_created"`
	Result      string `json:"result"`
	Reason      string `json:"reason"`
}

type ImageFileWithReason struct {
	models.ImageFile
	ReasonTitle string
	Reason      string
}

func CheckAllItems(ctx context.Context) (Err logging.LogErrorInfo) {
	ctx, getAllItemAction := logging.AddSubActionToContext(ctx, " Getting all saved sets for AutoDownload Check", logging.LevelInfo)
	out, Err := database.GetAllSavedSets(ctx, models.DBFilter{ItemsPerPage: -1})
	if Err.Message != "" {
		getAllItemAction.Complete()
		return *getAllItemAction.Error
	}
	getAllItemAction.Complete()

	mediaserver.GetAllLibrarySectionsAndItems(ctx, true)

	errorCount := 0
	warningCount := 0
	successCount := 0
	skippedCount := 0
	for _, item := range out.Items {
		itemCtx, ld := logging.CreateLoggingContext(context.Background(), "AutoDownload - Check For Updates")
		itemAction := ld.AddAction(fmt.Sprintf("Checking Item %s", utils.MediaItemInfo(item.MediaItem)), logging.LevelInfo)
		itemCtx = logging.WithCurrentAction(itemCtx, itemAction)
		result := CheckItem(itemCtx, item)
		switch result.OverallResult {
		case "error":
			errorCount++
		case "warning":
			warningCount++
		case "success":
			successCount++
		case "skipped":
			skippedCount++
		}
		itemAction.AppendResult("outcomes", result)
		ld.Log()
	}

	logging.LOGGER.Info().Timestamp().Int("error_count", errorCount).
		Int("warning_count", warningCount).
		Int("success_count", successCount).
		Int("skipped_count", skippedCount).
		Msg("Completed AutoDownload Check for all items")
	return logging.LogErrorInfo{}
}

func CheckItem(ctx context.Context, dbItem models.DBSavedItem) (result AutoDownloadResult) {
	switch dbItem.MediaItem.Type {
	case "movie":
		result = handleMovie(ctx, dbItem)
	case "show":
		result = handleShow(ctx, dbItem)
	default:
		result = AutoDownloadResult{}
		result.Item = utils.MediaItemInfo(dbItem.MediaItem)
		result.OverallResult = "error"
		result.OverallMessage = "Unknown media type"
	}
	return result
}

func seasonExists(mediaItem models.MediaItem, seasonNumber int) bool {
	if mediaItem.Series == nil {
		return false
	}
	for _, season := range mediaItem.Series.Seasons {
		if season.SeasonNumber == seasonNumber {
			return true
		}
	}
	return false
}

func episodeExists(mediaItem models.MediaItem, seasonNumber int, episodeNumber int) bool {
	if mediaItem.Series == nil {
		return false
	}
	for _, season := range mediaItem.Series.Seasons {
		if season.SeasonNumber == seasonNumber {
			for _, episode := range season.Episodes {
				if episode.EpisodeNumber == episodeNumber {
					return true
				}
			}
		}
	}
	return false
}

// Only mark size as changed if the size differs by at least 256KB to avoid minor Media Server metadata/container drift.
// If Old and New sizes are both under 256KB, consider that no change
// If one of the sizes is under 256KB and the other is over, consider that a change regardless of the difference, since that could indicate a change from a stub/small file to a real file or vice versa.
// If both sizes are over 256KB, then check if the difference is over 256KB to consider it a change.
func sizeReallyChanged(newSize int64, oldSize int64) bool {
	const sizeThresholdBytes int64 = 256 * 1024 // 256 KB threshold
	sizeDifference := newSize - oldSize
	if sizeDifference < 0 {
		sizeDifference = -sizeDifference
	}

	if newSize < sizeThresholdBytes && oldSize < sizeThresholdBytes {
		return false
	} else if (newSize < sizeThresholdBytes && oldSize >= sizeThresholdBytes) || (newSize >= sizeThresholdBytes && oldSize < sizeThresholdBytes) {
		return true
	} else if sizeDifference > sizeThresholdBytes {
		return true
	}
	return false
}

func durationReallyChanged(newDuration int64, oldDuration int64) bool {
	const durationThresholdSeconds int64 = 10
	durationDifference := newDuration - oldDuration
	if durationDifference < 0 {
		durationDifference = -durationDifference
	}
	if newDuration < durationThresholdSeconds || oldDuration < durationThresholdSeconds || durationDifference > durationThresholdSeconds {
		return true
	}
	return false
}
