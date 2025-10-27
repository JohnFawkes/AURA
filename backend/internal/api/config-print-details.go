package api

import (
	"aura/internal/logging"
	"aura/internal/masking"
	"fmt"
)

func (config *Config) PrintDetails() {
	logging.LOG.NoTime("Current Configuration Settings\n")

	// Auth Mode
	logging.LOG.NoTime("\tAuth Mode:\n")
	logging.LOG.NoTime(fmt.Sprintf("\t\tEnabled: %t\n", config.Auth.Enabled))
	if config.Auth.Enabled {
		logging.LOG.NoTime(fmt.Sprintf("\t\tPassword: %s\n", masking.Masking_Token(config.Auth.Password)))
	}

	// Logging Configuration
	logging.LOG.NoTime(fmt.Sprintf("\tLogging Level: %s\n", config.Logging.Level))

	// Media Server Configuration
	logging.LOG.NoTime("\tMedia Server\n")
	logging.LOG.NoTime(fmt.Sprintf("\t\tType: %s\n", config.MediaServer.Type))
	logging.LOG.NoTime(fmt.Sprintf("\t\tURL: %s\n", config.MediaServer.URL))
	logging.LOG.NoTime(fmt.Sprintf("\t\tToken: %s\n", masking.Masking_Token(config.MediaServer.Token)))
	logging.LOG.NoTime(fmt.Sprintf("\t\tUserID: %s\n", config.MediaServer.UserID))
	logging.LOG.NoTime("\t\tLibraries:\n")
	for _, library := range config.MediaServer.Libraries {
		logging.LOG.NoTime(fmt.Sprintf("\t\t\tName: %s\n", library.Name))
	}
	if config.MediaServer.Type == "Plex" {
		logging.LOG.NoTime(fmt.Sprintf("\t\tSeason Naming Convention: %s\n", config.MediaServer.SeasonNamingConvention))
	}

	// Mediux Configuration
	logging.LOG.NoTime("\tMediux\n")
	logging.LOG.NoTime(fmt.Sprintf("\t\tToken: %s\n", masking.Masking_Token(config.Mediux.Token)))
	logging.LOG.NoTime(fmt.Sprintf("\t\tDownload Quality: %s\n", config.Mediux.DownloadQuality))

	// Auto Download Configuration
	logging.LOG.NoTime("\tAuto Download\n")
	logging.LOG.NoTime(fmt.Sprintf("\t\tEnabled: %t\n", config.AutoDownload.Enabled))
	logging.LOG.NoTime(fmt.Sprintf("\t\tCron: %s\n", config.AutoDownload.Cron))

	// Cache Images and Save Image Next To Content
	logging.LOG.NoTime("\tImages Options\n")
	logging.LOG.NoTime(fmt.Sprintf("\t\tCache Images: %t\n", config.Images.CacheImages))
	logging.LOG.NoTime(fmt.Sprintf("\t\tSave Image Locally: %t\n", config.Images.SaveImagesLocally.Enabled))
	if config.Images.SaveImagesLocally.Enabled {
		if config.Images.SaveImagesLocally.Path != "" {
			logging.LOG.NoTime(fmt.Sprintf("\t\tSave Image Locally Path: %s\n", config.Images.SaveImagesLocally.Path))
		} else {
			logging.LOG.NoTime("\t\tSave Image Locally Path: Saving next to content\n")
		}
	}

	// TMDB Configuration
	if config.TMDB.APIKey != "" {
		logging.LOG.NoTime("\tTMDB\n")
		logging.LOG.NoTime(fmt.Sprintf("\t\tAPI Key: %s\n", masking.Masking_Token(config.TMDB.APIKey)))
	}

	// Labels and Tags Configuration
	if len(config.LabelsAndTags.Applications) > 0 {
		logging.LOG.NoTime("\tLabels and Tags\n")
		for _, application := range config.LabelsAndTags.Applications {
			logging.LOG.NoTime(fmt.Sprintf("\t\tApplication: %s\n", application.Application))
			logging.LOG.NoTime(fmt.Sprintf("\t\t\tEnabled: %t\n", application.Enabled))
			if application.Enabled {
				logging.LOG.NoTime(fmt.Sprintf("\t\t\tAdd: %v\n", application.Add))
				logging.LOG.NoTime(fmt.Sprintf("\t\t\tRemove: %v\n", application.Remove))
			}
		}
	}

	// Notification Configuration
	logging.LOG.NoTime("\tNotifications\n")
	logging.LOG.NoTime(fmt.Sprintf("\t\tEnabled: %t\n", config.Notifications.Enabled))
	for _, notification := range config.Notifications.Providers {
		logging.LOG.NoTime(fmt.Sprintf("\t\tProvider: %s\n", notification.Provider))
		switch notification.Provider {
		case "Discord":
			logging.LOG.NoTime(fmt.Sprintf("\t\t\tEnabled: %t\n", notification.Enabled))
			logging.LOG.NoTime(fmt.Sprintf("\t\t\tWebhook: %s\n", masking.Masking_WebhookURL(notification.Discord.Webhook)))
		case "Pushover":
			logging.LOG.NoTime(fmt.Sprintf("\t\t\tEnabled: %t\n", notification.Enabled))
			logging.LOG.NoTime(fmt.Sprintf("\t\t\tToken: %s\n", masking.Masking_Token(notification.Pushover.Token)))
			logging.LOG.NoTime(fmt.Sprintf("\t\t\tUserKey: %s\n", masking.Masking_Token(notification.Pushover.UserKey)))
		case "Gotify":
			logging.LOG.NoTime(fmt.Sprintf("\t\t\tEnabled: %t\n", notification.Enabled))
			logging.LOG.NoTime(fmt.Sprintf("\t\t\tURL: %s\n", notification.Gotify.URL))
			logging.LOG.NoTime(fmt.Sprintf("\t\t\tToken: %s\n", masking.Masking_Token(notification.Gotify.Token)))
		}
	}

	// Sonarr / Radarr Configuration
	if len(config.SonarrRadarr.Applications) > 0 {
		logging.LOG.NoTime("\tSonarr / Radarr\n")
		for _, app := range config.SonarrRadarr.Applications {
			logging.LOG.NoTime(fmt.Sprintf("\t\tType: %s (%s)\n", app.Type, app.Library))
			logging.LOG.NoTime(fmt.Sprintf("\t\t\tURL: %s\n", app.URL))
			logging.LOG.NoTime(fmt.Sprintf("\t\t\tAPI Key: %s\n", masking.Masking_Token(app.APIKey)))
		}
	}

}
