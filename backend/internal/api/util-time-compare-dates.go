package api

import (
	"time"
)

/*
Time_IsLastDownloadedBeforeLatestPosterSetDate compares the lastDownloaded string to the latestPosterSetDate time.Time

Returns: bool

If lastDownloaded is before latestPosterSetDate, returns true (indicating the item has been updated since last download)
If lastDownloaded is after or equal to latestPosterSetDate, returns false (indicating the item has not been updated since last download)
*/
func Time_IsLastDownloadedBeforeLatestPosterSetDate(lastDownloaded string, latestPosterSetDate time.Time) bool {
	lastDownloadedTime, err := time.Parse(time.RFC3339, lastDownloaded)
	if err != nil {
		// If parsing fails, we can't compare the dates, so return false
		return false
	}
	return lastDownloadedTime.Before(latestPosterSetDate)
}
