package api

import (
	"aura/internal/logging"
	"context"
	"time"
)

/*
Time_IsLastDownloadedBeforeLatestPosterSetDate compares the lastDownloaded string to the latestPosterSetDate time.Time

Returns: bool

If lastDownloaded is before latestPosterSetDate, returns true (indicating the item has been updated since last download)
If lastDownloaded is after or equal to latestPosterSetDate, returns false (indicating the item has not been updated since last download)
*/
func Time_IsLastDownloadedBeforeLatestPosterSetDate(ctx context.Context, lastDownloaded string, latestPosterSetDate time.Time) bool {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Comparing Time", logging.LevelTrace)
	defer logAction.Complete()
	lastDownloadedTime, err := time.Parse(time.RFC3339, lastDownloaded)
	if err != nil {
		logAction.AppendWarning("message", "Failed to parse lastDownloaded time")
		logAction.AppendResult("lastDownloaded", lastDownloaded)
		logAction.AppendResult("error", err.Error())
		// If parsing fails, we can't compare the dates, so return false
		return false
	}
	return lastDownloadedTime.Before(latestPosterSetDate)
}
