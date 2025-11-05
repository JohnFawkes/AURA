package api

import (
	"aura/internal/logging"
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/alexedwards/argon2id"
	"github.com/robfig/cron/v3"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func (config *Config) ValidateConfig() {
	ctx, ld := logging.CreateLoggingContext(context.Background(), "Config - Validate")
	defer ld.Log()

	// Top-level action
	action := ld.AddAction("Validating All Config Sections", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, action)
	defer action.Complete()

	// Sub-action: Auth Config
	isAuthValid := Config_ValidateAuth(ctx, &config.Auth)

	// Sub-action: Logging Config
	isLoggingValid := Config_ValidateLogging(ctx, &config.Logging)

	// // Sub-action: MediaServer Config
	isMediaServerValid := Config_ValidateMediaServer(ctx, &config.MediaServer)

	// Sub-action: Mediux Config
	isMediuxValid := Config_ValidateMediux(ctx, &config.Mediux)

	// Sub-action: AutoDownload Config
	isAutoDownloadValid := Config_ValidateAutoDownload(ctx, &config.AutoDownload)

	// Sub-action: Images Config
	isImagesValid := Config_ValidateImages(ctx, &config.Images, config.MediaServer)

	// Sub-action: Notifications Config
	isNotificationsValid := Config_ValidateNotifications(ctx, &config.Notifications)

	// Sub-action: SonarrRadarr Config
	isSonarrRadarrValid := Config_ValidateSonarrRadarr(ctx, &config.SonarrRadarr, config.MediaServer)

	// If any validation failed, set status to error
	if !isAuthValid || !isLoggingValid || !isMediaServerValid ||
		!isMediuxValid || !isAutoDownloadValid ||
		!isImagesValid || !isNotificationsValid || !isSonarrRadarrValid {
		ld.Status = logging.StatusError
		Global_Config_Valid = false
	} else {
		logging.SetLogLevel(config.Logging.Level)
		ld.Status = logging.StatusSuccess
		Global_Config_Valid = true
	}
}

func Config_ValidateAuth(ctx context.Context, Auth *Config_Auth) bool {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Validating Auth Config", logging.LevelTrace)
	defer logAction.Complete()

	isValid := true

	if Auth.Enabled {
		if Auth.Password == "" {
			logAction.SetError("Auth.Password is not set", "Password must be set when auth is enabled", nil)
			isValid = false
		} else {
			_, _, _, err := argon2id.DecodeHash(Auth.Password)
			if err != nil {
				logAction.SetError("Auth.Password is not a valid Argon2id hash", err.Error(), nil)

				isValid = false
			}
		}
	}

	return isValid
}

func Config_ValidateLogging(ctx context.Context, Logging *Config_Logging) bool {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Validating Logging Config", logging.LevelTrace)
	defer logAction.Complete()

	isValid := true

	// Check if Logging.Level is set
	if Logging.Level == "" {
		logAction.SetError("Logging.Level is not set", "Logging level must be specified", nil)
		isValid = false
	}

	// Set Logging.Level to uppercase for comparison
	Logging.Level = strings.ToUpper(Logging.Level)
	if Logging.Level != "TRACE" && Logging.Level != "DEBUG" && Logging.Level != "INFO" && Logging.Level != "WARN" && Logging.Level != "ERROR" {
		logAction.SetError("Logging.Level is not a valid level", "Valid levels are TRACE, DEBUG, INFO, WARN, ERROR", nil)
		isValid = false
	}

	return isValid
}

func Config_ValidateMediaServer(ctx context.Context, MediaServer *Config_MediaServer) bool {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Validating MediaServer Config", logging.LevelTrace)
	defer logAction.Complete()

	isValid := true

	// Title case the MediaServer.Type
	MediaServer.Type = cases.Title(language.English).String(MediaServer.Type)

	// Check if MediaServer.Type is set
	if MediaServer.Type == "" {
		logAction.SetError("MediaServer.Type is not set", "Media server type must be specified", nil)
		isValid = false
	} else if MediaServer.Type != "Plex" && MediaServer.Type != "Emby" && MediaServer.Type != "Jellyfin" {
		logAction.SetError(fmt.Sprintf("MediaServer.Type: '%s' is not a valid type", MediaServer.Type), "Valid types are Plex, Emby, Jellyfin", nil)
		isValid = false
	}

	// Check if MediaServer.URL is set
	if MediaServer.URL == "" {
		logAction.SetError("MediaServer.URL is not set", "Media server URL must be specified", nil)
		isValid = false
	} else if !strings.HasPrefix(MediaServer.URL, "http://") && !strings.HasPrefix(MediaServer.URL, "https://") {
		logAction.SetError(fmt.Sprintf("MediaServer.URL: '%s' must start with http:// or https:// ", MediaServer.URL), "", nil)
		isValid = false
	}

	// Check if MediaServer.Token is set
	if MediaServer.Token == "" {
		logAction.SetError("MediaServer.Token is not set", "Media server token must be specified", nil)
		isValid = false
	}

	// Check if MediaServer.Libraries is set
	if len(MediaServer.Libraries) == 0 {
		logAction.SetError("MediaServer.Libraries is not set", "At least one media server library must be specified", nil)
		isValid = false
	}

	// If we reach here, return if not valid
	if !isValid {
		return isValid
	}

	// Now that we know the Media Server config is valid, clean up fields and set defaults
	// Trim the trailing slash from the URL
	MediaServer.URL = strings.TrimSuffix(MediaServer.URL, "/")

	return isValid
}

func Config_ValidateMediux(ctx context.Context, Mediux *Config_Mediux) bool {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Validating Mediux Config", logging.LevelTrace)
	defer logAction.Complete()

	isValid := true

	// Check if Mediux.Token is set
	if Mediux.Token == "" {
		logAction.SetError("Mediux.Token is not set", "Mediux token must be specified", nil)
		isValid = false
	}

	// Check if Mediux.DownloadQuality is set
	if Mediux.DownloadQuality == "" {
		Mediux.DownloadQuality = "optimized"
		logAction.AppendWarning("message", "Mediux.DownloadQuality not set, defaulting to 'optimized'")
	} else if Mediux.DownloadQuality != "original" && Mediux.DownloadQuality != "optimized" {
		Mediux.DownloadQuality = "optimized"
		logAction.AppendWarning("message", "Mediux.DownloadQuality invalid, defaulting to 'optimized'")
	}

	return isValid
}

func Config_ValidateAutoDownload(ctx context.Context, AutoDownload *Config_AutoDownload) bool {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Validating AutoDownload Config", logging.LevelTrace)
	defer logAction.Complete()

	isValid := true

	// Check if AutoDownload is enabled
	if !AutoDownload.Enabled {
		return isValid
	}

	// Check if AutoDownload.Cron is set
	if AutoDownload.Cron == "" {
		AutoDownload.Cron = "0 0 * * *"
		logAction.AppendWarning("message", "AutoDownload.Cron not set, defaulting to '0 0 * * *' (every day at midnight)")
	}

	// Validate the cron expression
	if !Config_ValidateCron(AutoDownload.Cron) {
		logAction.SetError(fmt.Sprintf("AutoDownload.Cron: '%s' is not a valid cron expression", AutoDownload.Cron), "Please provide a valid cron expression", nil)
		isValid = false
	}

	return isValid
}

func Config_ValidateCron(cronExpression string) bool {
	_, err := cron.ParseStandard(cronExpression)
	return err == nil
}

func Config_ValidateImages(ctx context.Context, Images *Config_Images, msConfig Config_MediaServer) bool {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Validating Images Config", logging.LevelTrace)
	defer logAction.Complete()

	isValid := true

	// If Images.SaveImagesLocally.Enabled is true, validate the SeasonNamingConvention and EpisodeNamingConvention
	if Images.SaveImagesLocally.Enabled {
		if msConfig.Type != "Plex" {
			return isValid
		}

		validSeasonNamingConventions := []string{"1", "2"}
		validEpisodeNamingConventions := []string{"match", "static"}

		if !stringSliceContains(validSeasonNamingConventions, Images.SaveImagesLocally.SeasonNamingConvention) {
			if msConfig.Type == "Plex" && Images.SaveImagesLocally.SeasonNamingConvention == "" {
				Images.SaveImagesLocally.SeasonNamingConvention = "2"
				logAction.AppendWarning("message", "Images.SaveImagesLocally.SeasonNamingConvention not set, defaulting to '2'")
			} else if msConfig.Type == "Plex" && Images.SaveImagesLocally.SeasonNamingConvention != "1" && Images.SaveImagesLocally.SeasonNamingConvention != "2" {
				Images.SaveImagesLocally.SeasonNamingConvention = "2"
				logAction.AppendWarning("message", "Images.SaveImagesLocally.SeasonNamingConvention invalid, defaulting to '2'")
			}
		}

		Images.SaveImagesLocally.EpisodeNamingConvention = strings.ToLower(Images.SaveImagesLocally.EpisodeNamingConvention)
		if !stringSliceContains(validEpisodeNamingConventions, Images.SaveImagesLocally.EpisodeNamingConvention) {
			if msConfig.Type == "Plex" && Images.SaveImagesLocally.EpisodeNamingConvention == "" {
				Images.SaveImagesLocally.EpisodeNamingConvention = "match"
				logAction.AppendWarning("message", "Images.SaveImagesLocally.EpisodeNamingConvention not set, defaulting to 'match'")
			} else if msConfig.Type == "Plex" && Images.SaveImagesLocally.EpisodeNamingConvention != "match" && Images.SaveImagesLocally.EpisodeNamingConvention != "static" {
				Images.SaveImagesLocally.EpisodeNamingConvention = "match"
				logAction.AppendWarning("message", "Images.SaveImagesLocally.EpisodeNamingConvention invalid, defaulting to 'match'")
			}
		}

	}
	return isValid
}

func Config_ValidateNotifications(ctx context.Context, Notifications *Config_Notifications) bool {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Validating Notifications Config", logging.LevelTrace)
	defer logAction.Complete()

	isValid := true

	// If the notifications are not enabled, skip validation
	if !Notifications.Enabled || len(Notifications.Providers) == 0 {
		return isValid
	}

	// If notifications are enabled, validate each provider
	for i, provider := range Notifications.Providers {

		// Set the provider name to Title Case
		provider.Provider = cases.Title(language.English).String(provider.Provider)

		// If the provider name is not set, return an error
		if provider.Provider == "" {
			logAction.SetError(fmt.Sprintf("\tNotifications[%d].Provider is not set", i), "Provider name must be specified", nil)
			isValid = false
			return isValid
		}

		// If the provider is not enabled, log a warning and continue to the next provider
		if !provider.Enabled {
			logAction.AppendWarning("message", fmt.Sprintf("Notifications[%d].Provider '%s' is disabled, skipping validation", i, provider.Provider))
			continue
		}

		validProviders := []string{"Discord", "Pushover", "Gotify", "Webhook"}

		// If the provider is not in the list of valid providers, return an error
		if !stringSliceContains(validProviders, provider.Provider) {
			logAction.SetError(fmt.Sprintf("\tBad Notifications[%d].Provider: '%s'. Must be one of: %v", i, provider.Provider, validProviders), "Please provide a valid provider", nil)
			isValid = false
		}

		switch provider.Provider {
		case "Discord":
			if provider.Discord.Webhook == "" {
				logAction.SetError(fmt.Sprintf("\tNotifications[%d].Webhook is not set", i), "Discord webhook must be specified", nil)
				isValid = false
			}

		case "Pushover":
			if provider.Pushover.UserKey == "" {
				logAction.SetError(fmt.Sprintf("\tNotifications[%d].UserKey is not set", i), "Pushover UserKey must be specified", nil)
				isValid = false
			}
			if provider.Pushover.Token == "" {
				logAction.SetError(fmt.Sprintf("\tNotifications[%d].Token is not set", i), "Pushover Token must be specified", nil)
				isValid = false
			}

		case "Gotify":
			if provider.Gotify.URL == "" {
				logAction.SetError(fmt.Sprintf("\tNotifications[%d].URL is not set", i), "Gotify URL must be specified", nil)
				isValid = false
			}
			if provider.Gotify.Token == "" {
				logAction.SetError(fmt.Sprintf("\tNotifications[%d].Token is not set", i), "Gotify Token must be specified", nil)
				isValid = false
			}
		case "Webhook":
			if provider.Webhook.URL == "" {
				logAction.SetError(fmt.Sprintf("\tNotifications[%d].URL is not set", i), "Webhook URL must be specified", nil)
				isValid = false
			}
		}

		Notifications.Providers[i] = provider
	}

	return isValid
}

func Config_ValidateSonarrRadarr(ctx context.Context, apps *Config_SonarrRadarr_Apps, mediaServerConfig Config_MediaServer) bool {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Validating SonarrRadarr Config", logging.LevelTrace)
	defer logAction.Complete()

	isValid := true

	if len(apps.Applications) == 0 {
		return isValid
	}

	for i, app := range apps.Applications {

		// Set the app type to Title Case
		app.Type = cases.Title(language.English).String(app.Type)

		// If the app type is not set, return an error
		if app.Type == "" {
			logAction.SetError(fmt.Sprintf("\tSonarrRadarr[%d].Type is not set", i), "App type must be specified", nil)
			isValid = false
		} else if app.Type != "Sonarr" && app.Type != "Radarr" {
			// If the app type is not Sonarr or Radarr, return an error
			logAction.SetError(fmt.Sprintf("\tBad SonarrRadarr[%d].Type: '%s'. Must be one of: Sonarr, Radarr", i, app.Type), "Invalid app type", nil)
			isValid = false
		}

		libraryNames := make([]string, len(mediaServerConfig.Libraries))
		for idx, lib := range mediaServerConfig.Libraries {
			libraryNames[idx] = lib.Name // Replace 'Name' with the actual field name containing the library string
		}

		if app.Library == "" {
			logAction.SetError(fmt.Sprintf("\tSonarrRadarr[%d].Library is not set", i), "Library must be specified", nil)
			isValid = false
		} else if !stringSliceContains(libraryNames, app.Library) {
			// If the library is not in the list of MediaServer libraries, return an error
			logAction.SetError(fmt.Sprintf("\tBad SonarrRadarr[%d].Library: '%s'. Must be one of: %v", i, app.Library, libraryNames), "Invalid library", nil)
			isValid = false
		}

		if app.URL == "" {
			logAction.SetError(fmt.Sprintf("\tSonarrRadarr[%d].URL is not set", i), "App URL must be specified", nil)
			isValid = false
		} else if !strings.HasPrefix(app.URL, "http") {
			// If the URL does not start with http or https, return an error
			logAction.SetError(fmt.Sprintf("\tSonarrRadarr[%d].URL: '%s' must start with http:// or https:// ", i, app.URL), "", nil)
			isValid = false
		}

		if app.APIKey == "" {
			logAction.SetError(fmt.Sprintf("\tSonarrRadarr[%d].APIKey is not set", i), "App APIKey must be specified", nil)
			isValid = false
		}

		// Test Connection to the App
		valid, Err := SR_CallTestConnection(ctx, app)
		if Err.Message != "" || !valid {
			isValid = false
		}

		// Update the app in the list
		apps.Applications[i] = app
	}

	return isValid
}

// stringSliceContains checks if a string is present in a slice of strings.
func stringSliceContains(slice []string, item string) bool {
	return slices.Contains(slice, item)
}
