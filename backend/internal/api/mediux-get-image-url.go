package api

import (
	"aura/internal/logging"
	"context"
	"fmt"
	"net/url"
	"path"
	"strings"
	"time"
)

func Mediux_GetImageURL(ctx context.Context, assetID, dateTimeString string, imageQuality MediuxImageQuality) (string, logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, fmt.Sprintf("Constructing MediUX Image URL for Asset ID '%s'", assetID), logging.LevelTrace)
	defer logAction.Complete()

	if assetID == "" {
		logAction.SetError("Asset ID is required to construct MediUX URL",
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
	if imageQuality != MediuxImageQualityOriginal && imageQuality != MediuxImageQualityThumb && imageQuality != MediuxImageQualityOptimized {
		logAction.SetError("Invalid quality parameter",
			"Quality must be either 'original', 'thumb', or 'optimized'.",
			map[string]any{
				"imageQuality": imageQuality,
			})
		return "", *logAction.Error
	}

	// Format the date to YYYYMMDDHHMMSS
	dateTimeFormatted := dateTime.Format("20060102150405")

	qualityParam := ""
	switch imageQuality {
	case MediuxImageQualityThumb:
		qualityParam = "thumb"
	case MediuxImageQualityOptimized:
		qualityParam = "jpg"
	case MediuxImageQualityOriginal:
		// Leave qualityParam as empty for original
	}

	// Override qualityParam based on global config if it is set to optimized
	if imageQuality != MediuxImageQualityThumb && Global_Config.Mediux.DownloadQuality == "optimized" {
		qualityParam = "jpg"
	}

	var MediuxURL string
	if strings.HasPrefix(assetID, "---") {
		// Special handling for assets starting with "---"
		// This is for manual imports
		MediuxURL = "https://api.mediux.pro"
		assetID = strings.TrimPrefix(assetID, "---")
	} else {
		MediuxURL = MediuxBaseURL
	}

	// Construct the MediUX URL
	u, err := url.Parse(MediuxURL)
	if err != nil {
		logAction.SetError("Failed to parse MediUX base URL", err.Error(), nil)
		return "", *logAction.Error
	}
	u.Path = path.Join(u.Path, "assets", assetID)
	var queryStr string
	if qualityParam != "" {
		queryStr = fmt.Sprintf("v=%s&key=%s", dateTimeFormatted, qualityParam)
	} else {
		queryStr = fmt.Sprintf("v=%s", dateTimeFormatted)
	}
	u.RawQuery = queryStr
	URL := u.String()

	// Append Values
	logAction.AppendResult("url", URL)
	logAction.AppendResult("imageQuality", imageQuality)

	// Return the constructed URL
	return URL, logging.LogErrorInfo{}
}

func Mediux_GetImageURLFromSrc(src string) string {
	if src == "" {
		return ""
	}

	var MediuxURL string
	if strings.HasPrefix(src, "---") {
		// Special handling for assets starting with "---"
		// This is for manual imports
		MediuxURL = "https://api.mediux.pro"
		src = strings.TrimPrefix(src, "---")
	} else {
		MediuxURL = MediuxBaseURL
	}

	// Construct the MediUX URL
	u, err := url.Parse(MediuxURL)
	if err != nil {
		return ""
	}
	u.Path = path.Join(u.Path, "assets", src)
	return u.String() + "&key=thumb"
}
