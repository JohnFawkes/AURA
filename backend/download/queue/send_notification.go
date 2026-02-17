package downloadqueue

import (
	"aura/config"
	"aura/logging"
	"aura/mediux"
	"aura/models"
	"aura/notification"
	"context"
	"fmt"
	"strings"
	"time"
)

func SendNotification(
	fileIssues FileIssues,
	itemTitle string,
	posterSet models.DBPosterSetDetail,
	tmdbPoster string,
	tmdbBackdrop string,
) {
	if len(config.Current.Notifications.Providers) == 0 || !config.Current.Notifications.Enabled {
		return
	}

	var result Status
	if len(fileIssues.Errors) > 0 {
		result = LAST_STATUS_ERROR
	} else if len(fileIssues.Warnings) > 0 {
		result = LAST_STATUS_WARNING
	} else {
		result = LAST_STATUS_SUCCESS
	}

	if posterSet.ID == "" {
		posterSet.ID = "Unknown Set ID"
	}

	imageURL := getImageURLFromPosterSet(posterSet, tmdbPoster, tmdbBackdrop)

	notificationTitle := ""
	messageBody := ""

	switch result {
	case LAST_STATUS_SUCCESS:
		notificationTitle = "Download Queue - Success"
		messageBody = fmt.Sprintf("%s (Set: %s)%s", itemTitle, posterSet.ID)
	case LAST_STATUS_WARNING:
		notificationTitle = "Download Queue - Warning"
		messageBody = fmt.Sprintf("%s (Set: %s)%s\n\nWarnings:\n%s", itemTitle, posterSet.ID, strings.Join(fileIssues.Warnings, "\n"))
	case LAST_STATUS_ERROR:
		notificationTitle = "Download Queue - Error"
		messageBody = fmt.Sprintf("%s (Set: %s)%s\n\nErrors:\n%s", itemTitle, posterSet.ID, strings.Join(fileIssues.Errors, "\n"))
		if len(fileIssues.Warnings) > 0 {
			messageBody += fmt.Sprintf("\n\nWarnings:\n%s", strings.Join(fileIssues.Warnings, "\n"))
		}
	}
	// Update the Global LatestInfo
	LatestInfo.Time = time.Now()
	LatestInfo.Status = result
	LatestInfo.Message = fmt.Sprintf("%s (Set: %s)", itemTitle, posterSet.ID)
	LatestInfo.Errors = fileIssues.Errors
	LatestInfo.Warnings = fileIssues.Warnings

	ctx, ld := logging.CreateLoggingContext(context.Background(), "Notification - Send Download Queue Update")
	logAction := ld.AddAction("Sending Download Queue Notification", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)
	defer ld.Log()
	defer logAction.Complete()

	// Send notification using all configured providers
	for _, provider := range config.Current.Notifications.Providers {
		if provider.Enabled {
			switch provider.Provider {
			case "Discord":
				notification.SendDiscordMessage(
					ctx,
					provider.Discord,
					messageBody,
					imageURL,
					notificationTitle,
				)
			case "Pushover":
				notification.SendPushoverMessage(
					ctx,
					provider.Pushover,
					messageBody,
					imageURL,
					notificationTitle,
				)
			case "Gotify":
				notification.SendGotifyMessage(
					ctx,
					provider.Gotify,
					messageBody,
					imageURL,
					notificationTitle,
				)
			case "Webhook":
				notification.SendWebhookMessage(
					ctx,
					provider.Webhook,
					messageBody,
					imageURL,
					notificationTitle,
				)
			}
		}
	}
}

func getImageURLFromPosterSet(posterSet models.DBPosterSetDetail, tmdbPoster, tmdbBackdrop string) string {
	item_tmdb_id := ""
	posterURL := ""
	backdropURL := ""
	seasonURL := ""
	titlecardURL := ""
	tmdbPosterURL := tmdbPoster
	tmdbBackdropURL := tmdbBackdrop

	for _, img := range posterSet.Images {
		switch img.Type {
		case "poster":
			posterURL = mediux.GetImageURLFromSrc(img.Src)
			if item_tmdb_id == "" {
				item_tmdb_id = img.ItemTMDB_ID
			}
		case "backdrop":
			backdropURL = mediux.GetImageURLFromSrc(img.Src)
			if item_tmdb_id == "" {
				item_tmdb_id = img.ItemTMDB_ID
			}
		case "seasonPoster":
			if seasonURL == "" {
				seasonURL = mediux.GetImageURLFromSrc(img.Src)
				if item_tmdb_id == "" {
					item_tmdb_id = img.ItemTMDB_ID
				}
			}
		case "titlecard":
			if titlecardURL == "" {
				titlecardURL = mediux.GetImageURLFromSrc(img.Src)
				if item_tmdb_id == "" {
					item_tmdb_id = img.ItemTMDB_ID
				}
			}
		}
	}

	if item_tmdb_id != "" && tmdbPoster == "" && tmdbBackdrop == "" {
		itemType := posterSet.Type
		if posterSet.Type == "collection" {
			itemType = "movie"
		}

		itemInfo, Err := mediux.GetBaseItemInfoByTMDB_ID(item_tmdb_id, itemType)
		if Err.Message != "" {
			tmdbPosterURL = itemInfo.TMDB_PosterPath
			tmdbBackdropURL = itemInfo.TMDB_BackdropPath
		}
	}

	// Single-image fallback order:
	// poster -> tmdb poster -> backdrop -> tmdb backdrop -> season -> titlecard
	if posterURL != "" {
		return posterURL
	}
	if tmdbPosterURL != "" {
		return tmdbPosterURL
	}
	if backdropURL != "" {
		return backdropURL
	}
	if tmdbBackdropURL != "" {
		return tmdbBackdropURL
	}
	if seasonURL != "" {
		return seasonURL
	}
	if titlecardURL != "" {
		return titlecardURL
	}
	return ""
}
