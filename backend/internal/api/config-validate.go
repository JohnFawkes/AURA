package api

import (
	"aura/internal/logging"
	"fmt"
	"slices"
	"strings"

	"github.com/alexedwards/argon2id"
	"github.com/robfig/cron/v3"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func (config *Config) ValidateConfig() {

	logging.LOG.Debug("Validating configuration file...")

	// Check if config is nil
	if config == nil {
		logging.LOG.Error("\tConfig is nil")
		return
	}

	// Validate Auth configuration
	isAuthValid, _ := Config_ValidateAuth(config.Auth)

	// Validate Logging configuration
	isLoggingValid, _, loggingConfig := Config_ValidateLogging(config.Logging)
	logging.SetLogLevel(loggingConfig.Level)

	// Check MediaServer configuration
	isMediaServerValid, _, mediaServerConfig := Config_ValidateMediaServer(config.MediaServer)
	config.MediaServer = mediaServerConfig

	// Validate Mediux configuration
	isMediuxValid, _, mediuxConfig := Config_ValidateMediux(config.Mediux)
	config.Mediux = mediuxConfig

	// Validate AutoDownload configuration
	isAutoDownloadValid, _, autodownloadConfig := Config_ValidateAutoDownload(config.AutoDownload)
	config.AutoDownload = autodownloadConfig

	// Validate Notifications configuration
	isNotificationsValid, _, notificationsConfig := Config_ValidateNotifications(config.Notifications)
	config.Notifications = notificationsConfig

	// Validate Sonarr/Radarr configuration
	isSonarrRadarrValid, _, sonarrRadarrConfig := Config_ValidateSonarrRadarr(config.SonarrRadarr, mediaServerConfig)
	config.SonarrRadarr = sonarrRadarrConfig

	if isAuthValid {
		Global_Config.Auth = config.Auth
	}

	if isLoggingValid {
		Global_Config.Logging = config.Logging
	}

	if isMediaServerValid {
		Global_Config.MediaServer = config.MediaServer
	}

	if isAutoDownloadValid {
		Global_Config.AutoDownload = config.AutoDownload
	}

	if isMediuxValid {
		Global_Config.Mediux = config.Mediux
	}

	if isNotificationsValid {
		Global_Config.Notifications = config.Notifications
	}

	if isSonarrRadarrValid {
		Global_Config.SonarrRadarr = config.SonarrRadarr
	}

	if !isAuthValid || !isLoggingValid || !isMediaServerValid || !isAutoDownloadValid || !isMediuxValid || !isNotificationsValid || !isSonarrRadarrValid {
		logging.LOG.Error("Invalid configuration file. See errors above.")
		Global_Config_Valid = false
		return
	}

	Global_Config_Valid = true
}

func Config_ValidateAuth(Auth Config_Auth) (bool, string) {
	isValid := true
	var errorMsg string

	if Auth.Enabled {
		if Auth.Password == "" {
			errorMsg = "Auth.Password is not set"
			logging.LOG.Error(errorMsg)
			isValid = false
		} else {
			// Check if Password is a valid Argon2id hash
			_, _, _, err := argon2id.DecodeHash(Auth.Password)
			if err != nil {
				errorMsg = "Auth.Password is not a valid Argon2id hash"
				logging.LOG.Error(errorMsg)
				isValid = false
			}
		}
	}

	return isValid, errorMsg
}

func Config_ValidateLogging(Logging Config_Logging) (bool, string, Config_Logging) {
	isValid := true
	var errorMsg string

	if Logging.Level == "" {
		errorMsg = "Logging.Level is not set, using default level: INFO"
		logging.LOG.Warn(errorMsg)
		Logging.Level = "INFO"
	}

	if Logging.Level != "TRACE" && Logging.Level != "DEBUG" && Logging.Level != "INFO" && Logging.Level != "WARN" && Logging.Level != "ERROR" {
		errorMsg = fmt.Sprintf("\tLogging.Level: '%s'. Must be one of: TRACE, DEBUG, INFO, WARN, ERROR", Logging.Level)
		logging.LOG.Warn(errorMsg)
		Logging.Level = "INFO"
	}

	return isValid, errorMsg, Logging

}

func Config_ValidateMediaServer(MediaServer Config_MediaServer) (bool, []string, Config_MediaServer) {
	isValid := true
	var errorMsgs []string
	var errorMsg string

	if MediaServer.Type == "" {
		errorMsg = "MediaServer.Type is not set"
		logging.LOG.Warn(errorMsg)
		errorMsgs = append(errorMsgs, errorMsg)
		isValid = false
	}

	if MediaServer.URL == "" {
		errorMsg = "MediaServer.URL is not set"
		logging.LOG.Warn(errorMsg)
		errorMsgs = append(errorMsgs, errorMsg)
		isValid = false
	} else if !strings.HasPrefix(MediaServer.URL, "http") {
		errorMsg = fmt.Sprintf("MediaServer.URL: '%s' must start with http:// or https:// ", MediaServer.URL)
		logging.LOG.Warn(errorMsg)
		errorMsgs = append(errorMsgs, errorMsg)
		isValid = false
	}

	if MediaServer.Token == "" {
		errorMsg = "MediaServer.Token is not set"
		logging.LOG.Warn(errorMsg)
		errorMsgs = append(errorMsgs, errorMsg)
		isValid = false
	}

	if len(MediaServer.Libraries) == 0 {
		errorMsg = "MediaServer.Libraries are not set"
		logging.LOG.Warn(errorMsg)
		errorMsgs = append(errorMsgs, errorMsg)
		isValid = false
	}

	if !isValid {
		logging.LOG.Error("\tMediaServer configuration is invalid. Fix the errors above.")
		return false, errorMsgs, MediaServer
	}

	// Set the MediaServer Type to Title Case
	MediaServer.Type = cases.Title(language.English).String(MediaServer.Type)

	// If the MediaServer type is not Plex, Emby, or Jellyfin, return an error
	if MediaServer.Type != "Plex" && MediaServer.Type != "Emby" && MediaServer.Type != "Jellyfin" {
		errorMsg = fmt.Sprintf("\tMediaServer.Type: '%s'. Must be one of: Plex, Emby, Jellyfin", MediaServer.Type)
		logging.LOG.Error(errorMsg)
		errorMsgs = append(errorMsgs, errorMsg)
		return false, errorMsgs, MediaServer
	}

	// If the MediaServer type is Plex, set the SeasonNamingConvention to 2 if not set
	if MediaServer.Type == "Plex" && MediaServer.SeasonNamingConvention == "" {
		logging.LOG.Warn("\tMediaServer.SeasonNamingConvention is not set, using default value: 2")
		MediaServer.SeasonNamingConvention = "2"
	}
	// If the SeasonNamingConvention is not 1 or 2, return an error
	if MediaServer.Type == "Plex" && MediaServer.SeasonNamingConvention != "1" && MediaServer.SeasonNamingConvention != "2" {
		errorMsg = fmt.Sprintf("\tBad MediaServer.SeasonNamingConvention: '%s'. Must be one of: 1, 2", MediaServer.SeasonNamingConvention)
		logging.LOG.Error(errorMsg)
		errorMsgs = append(errorMsgs, errorMsg)
		return false, errorMsgs, MediaServer
	}

	// Trim the trailing slash from the URL
	MediaServer.URL = strings.TrimSuffix(MediaServer.URL, "/")

	return true, errorMsgs, MediaServer
}

func Config_ValidateMediux(Mediux Config_Mediux) (bool, []string, Config_Mediux) {
	isValid := true
	var errorMsgs []string
	var errorMsg string

	if Mediux.Token == "" {
		errorMsg = "\tMediux.Token is not set"
		logging.LOG.Warn(errorMsg)
		errorMsgs = append(errorMsgs, errorMsg)
		isValid = false
	}

	if Mediux.DownloadQuality == "" {
		logging.LOG.Warn("\tMediux.DownloadQuality is not set, using default value: optimized")
		Mediux.DownloadQuality = "optimized"
	}

	if Mediux.DownloadQuality != "original" && Mediux.DownloadQuality != "optimized" {
		errorMsg = fmt.Sprintf("\tBad Mediux.DownloadQuality: '%s'. Must be one of: original, optimized", Mediux.DownloadQuality)
		logging.LOG.Error(errorMsg)
		errorMsgs = append(errorMsgs, errorMsg)
		isValid = false
	}

	return isValid, errorMsgs, Mediux
}

func Config_ValidateAutoDownload(Autodownload Config_AutoDownload) (bool, []string, Config_AutoDownload) {
	isValid := true
	var errorMsgs []string

	if !Autodownload.Enabled {
		return isValid, errorMsgs, Autodownload
	}

	if Autodownload.Cron == "" {
		logging.LOG.Warn("\tAutoDownload.Cron is not set, using default value: 0 0 * * *")
		Autodownload.Cron = "0 0 * * *" // Default to daily at midnight
	}

	if !Config_ValidateCron(Autodownload.Cron) {
		errorMsg := fmt.Sprintf("\tBad AutoDownload.Cron: '%s'. Must be a valid cron expression. Use something like https://crontab.guru to help you.", Autodownload.Cron)
		logging.LOG.Error(errorMsg)
		errorMsgs = append(errorMsgs, errorMsg)
		isValid = false
	}

	return isValid, errorMsgs, Autodownload
}

func Config_ValidateCron(cronExpression string) bool {
	_, err := cron.ParseStandard(cronExpression)
	return err == nil
}

func Config_ValidateNotifications(Notifications Config_Notifications) (bool, []string, Config_Notifications) {
	isValid := true
	var errorMsgs []string
	var errorMsg string

	// If the notifications are not enabled, skip validation
	if !Notifications.Enabled {
		return isValid, errorMsgs, Notifications
	}

	// If notifications are enabled, validate each provider
	for i, provider := range Notifications.Providers {

		// Set the provider name to Title Case
		provider.Provider = cases.Title(language.English).String(provider.Provider)

		// If the provider name is not set, return an error
		if provider.Provider == "" {
			errorMsg = fmt.Sprintf("\tNotifications[%d].Provider is not set", i)
			logging.LOG.Warn(errorMsg)
			errorMsgs = append(errorMsgs, errorMsg)
			return false, errorMsgs, Notifications
		}

		// If the provider is not enabled, log a warning and continue to the next provider
		if !provider.Enabled {
			logging.LOG.Warn(fmt.Sprintf("\tNotifications for %s are disabled", provider.Provider))
			continue
		}

		validProviders := []string{"Discord", "Pushover", "Gotify"}

		// If the provider is not in the list of valid providers, return an error
		if !stringSliceContains(validProviders, provider.Provider) {
			errorMsg = fmt.Sprintf("\tBad Notifications[%d].Provider: '%s'. Must be one of: %v", i, provider.Provider, validProviders)
			logging.LOG.Error(errorMsg)
			errorMsgs = append(errorMsgs, errorMsg)
			return false, errorMsgs, Notifications
		}

		switch provider.Provider {
		case "Discord":
			if provider.Discord.Webhook == "" {
				errorMsg = fmt.Sprintf("\tNotifications[%d].Webhook URL is not set", i)
				logging.LOG.Warn(errorMsg)
				errorMsgs = append(errorMsgs, errorMsg)
				return false, errorMsgs, Notifications
			}

		case "Pushover":
			if provider.Pushover.UserKey == "" {
				errorMsg = fmt.Sprintf("\tNotifications[%d].UserKey is not set", i)
				logging.LOG.Warn(errorMsg)
				errorMsgs = append(errorMsgs, errorMsg)
				return false, errorMsgs, Notifications
			}
			if provider.Pushover.Token == "" {
				errorMsg = fmt.Sprintf("\tNotifications[%d].Token is not set", i)
				logging.LOG.Warn(errorMsg)
				errorMsgs = append(errorMsgs, errorMsg)
				return false, errorMsgs, Notifications
			}

		case "Gotify":
			if provider.Gotify.URL == "" {
				errorMsg = fmt.Sprintf("\tNotifications[%d].URL is not set", i)
				logging.LOG.Warn(errorMsg)
				errorMsgs = append(errorMsgs, errorMsg)
				return false, errorMsgs, Notifications
			}
			if provider.Gotify.Token == "" {
				errorMsg = fmt.Sprintf("\tNotifications[%d].Token is not set", i)
				logging.LOG.Warn(errorMsg)
				errorMsgs = append(errorMsgs, errorMsg)
				return false, errorMsgs, Notifications
			}
		}

		Notifications.Providers[i] = provider

	}

	return isValid, errorMsgs, Notifications
}

func Config_ValidateSonarrRadarr(apps Config_SonarrRadarr_Apps, mediaServerConfig Config_MediaServer) (bool, []string, Config_SonarrRadarr_Apps) {
	isValid := true
	var errorMsgs []string
	var errorMsg string

	if len(apps.Applications) == 0 {
		return isValid, errorMsgs, apps
	}

	for i, app := range apps.Applications {

		// Set the app type to Title Case
		app.Type = cases.Title(language.English).String(app.Type)

		// If the app type is not set, return an error
		if app.Type == "" {
			errorMsg = fmt.Sprintf("\tSonarrRadarr[%d].Type is not set", i)
			logging.LOG.Warn(errorMsg)
			errorMsgs = append(errorMsgs, errorMsg)
			isValid = false
		} else if app.Type != "Sonarr" && app.Type != "Radarr" {
			// If the app type is not Sonarr or Radarr, return an error
			errorMsg = fmt.Sprintf("\tBad SonarrRadarr[%d].Type: '%s'. Must be one of: Sonarr, Radarr", i, app.Type)
			logging.LOG.Error(errorMsg)
			errorMsgs = append(errorMsgs, errorMsg)
			isValid = false
		}

		libraryNames := make([]string, len(mediaServerConfig.Libraries))
		for idx, lib := range mediaServerConfig.Libraries {
			libraryNames[idx] = lib.Name // Replace 'Name' with the actual field name containing the library string
		}

		if app.Library == "" {
			errorMsg = fmt.Sprintf("\tSonarrRadarr[%d].Library is not set", i)
			logging.LOG.Warn(errorMsg)
			errorMsgs = append(errorMsgs, errorMsg)
			isValid = false
		} else if !stringSliceContains(libraryNames, app.Library) {
			// If the library is not in the list of MediaServer libraries, return an error
			errorMsg = fmt.Sprintf("\tBad SonarrRadarr[%d].Library: '%s'. Must be one of: %v", i, app.Library, libraryNames)
			logging.LOG.Error(errorMsg)
			errorMsgs = append(errorMsgs, errorMsg)
			isValid = false
		}

		if app.URL == "" {
			errorMsg = fmt.Sprintf("\tSonarrRadarr[%d].URL is not set", i)
			logging.LOG.Warn(errorMsg)
			errorMsgs = append(errorMsgs, errorMsg)
			isValid = false
		} else if !strings.HasPrefix(app.URL, "http") {
			// If the URL does not start with http or https, return an error
			errorMsg = fmt.Sprintf("\tSonarrRadarr[%d].URL: '%s' must start with http:// or https:// ", i, app.URL)
			logging.LOG.Warn(errorMsg)
			errorMsgs = append(errorMsgs, errorMsg)
			isValid = false
		}

		if app.APIKey == "" {
			errorMsg = fmt.Sprintf("\tSonarrRadarr[%d].APIKey is not set", i)
			logging.LOG.Warn(errorMsg)
			errorMsgs = append(errorMsgs, errorMsg)
			isValid = false
		}

		// Test Connection to the App
		valid, Err := SR_CallTestConnection(app)
		if Err.Message != "" || !valid {
			logging.LOG.Warn(Err.Message)
			errorMsgs = append(errorMsgs, Err.Message)
			isValid = false
		}

		// Update the app in the list
		apps.Applications[i] = app
	}

	if !isValid {
		logging.LOG.Error("\tSonarrRadarr configuration is invalid. Fix the errors above.")
		return false, errorMsgs, apps
	}

	return isValid, errorMsgs, apps
}

// stringSliceContains checks if a string is present in a slice of strings.
func stringSliceContains(slice []string, item string) bool {
	return slices.Contains(slice, item)
}
