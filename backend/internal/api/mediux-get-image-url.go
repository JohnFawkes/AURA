package api

import (
	"aura/internal/logging"
	"context"
	"fmt"
	"net/url"
	"path"
	"time"
)

func Mediux_GetImageURL(ctx context.Context, assetID, dateTimeString, quality string) (string, logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, fmt.Sprintf("Constructing Mediux Image URL for Asset ID '%s'", assetID), logging.LevelTrace)
	defer logAction.Complete()

	if assetID == "" {
		logAction.SetError("Asset ID is required to construct Mediux URL",
			"Please provide a valid Asset ID.",
			nil)
		return "", *logAction.Error
	}

	var dateTime time.Time
	if dateTimeString == "" || dateTimeString == "0" || dateTimeString == "undefined" {
		// Use today's date if the modified date is not provided
		dateTime = time.Now()
	} else {
		// Try multiple date formats
		var err error

		// First try RFC3339 format (ISO 8601)
		dateTime, err = time.Parse(time.RFC3339, dateTimeString)
		if err != nil {
			// If that fails, try the compact format (YYYYMMDDHHMMSS)
			dateTime, err = time.Parse("20060102150405", dateTimeString)
			if err != nil {
				// Try Go's time.String() format
				dateTime, err = time.Parse("2006-01-02 15:04:05 -0700 MST", dateTimeString)
				if err != nil {
					// If all formats fail, log the error but use current time as fallback
					logAction.AppendWarning("message", "Failed to parse dateTimeString, defaulting to current time")
					logAction.AppendWarning("dateTimeString", dateTimeString)
					logAction.AppendWarning("error", err.Error())
					dateTime = time.Now()
				}
			}
		}
	}

	// Check quality is set to "original" or "thumb"
	if quality != "original" && quality != "thumb" && quality != "optimized" {
		logAction.SetError("Invalid quality parameter",
			"Quality must be either 'original', 'thumb', or 'optimized'.",
			map[string]any{
				"quality": quality,
			})
		return "", *logAction.Error
	}

	// Format the date to YYYYMMDDHHMMSS
	dateTimeFormatted := dateTime.Format("20060102150405")

	qualityParam := ""
	if quality == "thumb" {
		qualityParam = "thumb"
	} else if Global_Config.Mediux.DownloadQuality == "optimized" || quality == "optimized" {
		// If quality is "optimized" or configured as such, use optimized quality
		qualityParam = "jpg"
	}

	// Construct the Mediux URL
	u, err := url.Parse(MediuxBaseURL)
	if err != nil {
		logAction.SetError("Failed to parse Mediux base URL", err.Error(), nil)
		return "", *logAction.Error
	}
	u.Path = path.Join(u.Path, "assets", assetID)
	query := u.Query()
	query.Set("v", dateTimeFormatted)
	if qualityParam != "" {
		query.Set("key", qualityParam)
	}
	u.RawQuery = query.Encode()
	URL := u.String()
	logAction.AppendResult("url", URL)

	return URL, logging.LogErrorInfo{}
}
