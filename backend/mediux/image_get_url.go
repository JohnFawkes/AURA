package mediux

import (
	"aura/config"
	"aura/logging"
	"aura/utils"
	"context"
	"fmt"
	"net/url"
	"path"
	"strings"
)

func ConstructImageUrl(ctx context.Context, assetID, dateTimeString string, imageQuality ImageQuality) (imageURL string, Err logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, fmt.Sprintf("MediUX: Constructing Image URL for Asset ID '%s'", assetID), logging.LevelTrace)
	defer logAction.Complete()

	if assetID == "" {
		logAction.SetError("Asset ID is required to construct MediUX URL",
			"Please provide a valid Asset ID.",
			nil)
		return "", *logAction.Error
	}

	// Check quality is set to "original" or "thumb"
	if imageQuality != ImageQualityOriginal && imageQuality != ImageQualityThumb && imageQuality != ImageQualityOptimized {
		logAction.SetError("Invalid quality parameter",
			"Quality must be either 'original', 'thumb', or 'optimized'.",
			map[string]any{
				"imageQuality": imageQuality,
			})
		return "", *logAction.Error
	}

	// Format the date to YYYYMMDDHHMMSS
	dateTime := utils.ConvertDateStringToTime(dateTimeString)
	dateTimeFormatted := dateTime.Format("20060102150405")

	qualityParam := ""
	switch imageQuality {
	case ImageQualityThumb:
		qualityParam = "thumb"
	case ImageQualityOptimized:
		qualityParam = "jpg"
	case ImageQualityOriginal:
		// Leave qualityParam as empty for original
	}

	// Override qualityParam based on global config if it is set to optimized
	if imageQuality != ImageQualityThumb && config.Current.Mediux.DownloadQuality == "optimized" {
		qualityParam = "jpg"
	}

	var MediuxURL string
	if strings.HasPrefix(assetID, "---") {
		// Special handling for assets starting with "---"
		// This is for manual imports
		MediuxURL = MediuxPublicURL
		assetID = strings.TrimPrefix(assetID, "---")
		qualityParam = ""
	} else {
		MediuxURL = MediuxApiURL
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

func GetImageURLFromSrc(src string) string {
	if src == "" {
		return ""
	}

	var MediuxURL string
	if strings.HasPrefix(src, "---") {
		// Special handling for assets starting with "---"
		// This is for manual imports
		MediuxURL = MediuxPublicURL
		src = strings.TrimPrefix(src, "---")
	} else {
		MediuxURL = MediuxApiURL
	}

	// Construct the MediUX URL
	u, err := url.Parse(MediuxURL)
	if err != nil {
		return ""
	}
	u.Path = path.Join(u.Path, "assets", src)
	return u.String() + "&key=thumb"
}
