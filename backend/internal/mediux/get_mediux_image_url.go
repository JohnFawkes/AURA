package mediux

import (
	"aura/internal/config"
	"aura/internal/logging"
	"fmt"
	"time"
)

func GetMediuxImageURL(assetID, dateTimeString, quality string) (string, logging.StandardError) {
	Err := logging.NewStandardError()

	if assetID == "" {
		Err.Message = "Missing asset ID"
		Err.HelpText = "Ensure the asset ID is provided."
		Err.Details = "Asset ID cannot be empty."
		return "", Err
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
					logging.LOG.Warn(fmt.Sprintf("Failed to parse dateTime '%s': %v. Using current time as fallback.", dateTimeString, err))
					dateTime = time.Now()
				}
			}
		}
	}

	// Check quality is set to "original" or "thumb"
	if quality != "original" && quality != "thumb" && quality != "optimized" {
		Err.Message = "Invalid quality parameter"
		Err.HelpText = "Quality must be either 'original', 'thumb', or 'optimized'."
		Err.Details = map[string]any{
			"quality": quality,
		}
		return "", Err
	}

	// Format the date to YYYYMMDDHHMMSS
	dateTimeFormatted := dateTime.Format("20060102150405")

	qualityParam := ""
	if quality == "thumb" {
		qualityParam = "&key=thumb"
	} else if config.Global.Mediux.DownloadQuality == "optimized" || quality == "optimized" {
		// If quality is "optimized" or configured as such, use optimized quality
		qualityParam = "&key=jpg"
	}

	// Construct the Mediux URL
	mediuxURL := fmt.Sprintf("%s/%s?v=%s%s", "https://images.mediux.io/assets", assetID, dateTimeFormatted, qualityParam)
	logging.LOG.Trace(fmt.Sprintf("Constructed Mediux URL: %s", mediuxURL))

	return mediuxURL, logging.StandardError{}
}
