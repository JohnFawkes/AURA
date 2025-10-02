package config

import (
	"aura/internal/logging"
	"aura/internal/modals"
	"aura/internal/utils/masking"
	"fmt"
	"os"
	"path"
	"slices"
	"strings"

	"github.com/alexedwards/argon2id"
	"github.com/robfig/cron/v3"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

// LoadYamlConfig loads the application configuration from a YAML file.
//
// Returns:
//   - An error if the configuration file is missing, unreadable, or invalid.
func LoadYamlConfig() {
	logging.LOG.Debug("Attempting to load configuration from YAML file...")
	// Use an environment variable to determine the config path
	// By default, it will use /config
	// This is useful for testing and local development
	// In Docker, the config path is set to /config
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "/config"
	}

	// Check for a config.yaml or config.yml file
	yamlFile := path.Join(configPath, "config.yaml")
	if _, err := os.Stat(yamlFile); os.IsNotExist(err) {
		yamlFile = path.Join(configPath, "config.yml")
		if _, err := os.Stat(yamlFile); os.IsNotExist(err) {
			logging.LOG.Error(fmt.Sprintf("Config file not found in '%s'", configPath))
			return
		}
	}

	// Read the YAML file
	data, err := os.ReadFile(yamlFile)
	if err != nil {
		logging.LOG.Error(fmt.Sprintf("failed to read config.yaml file: %s", err.Error()))
		return
	}

	// Parse the YAML file into a Config struct
	var config modals.Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		logging.LOG.Error(fmt.Sprintf("failed to parse config.yaml file: %s", err.Error()))
		return
	}

	Global = &config
	ConfigLoaded = true
}

func ValidateConfig(config *modals.Config) {

	logging.LOG.Debug("Validating configuration file...")

	// Check if config is nil
	if config == nil {
		logging.LOG.Error("\tConfig is nil")
		return
	}

	// Validate Auth configuration
	isAuthValid, _ := ValidateAuthConfig(config.Auth)

	// Validate Logging configuration
	isLoggingValid, _, loggingConfig := ValidateLoggingConfig(config.Logging)
	logging.SetLogLevel(loggingConfig.Level)

	// Check MediaServer configuration
	isMediaServerValid, _, mediaServerConfig := ValidateMediaServerConfig(config.MediaServer)
	config.MediaServer = mediaServerConfig

	// Validate Mediux configuration
	isMediuxValid, _, mediuxConfig := ValidateMediuxConfig(config.Mediux)
	config.Mediux = mediuxConfig

	// Validate AutoDownload configuration
	isAutoDownloadValid, _, autodownloadConfig := ValidateAutoDownloadConfig(config.AutoDownload)
	config.AutoDownload = autodownloadConfig

	// Validate Notifications configuration
	isNotificationsValid, _, notificationsConfig := ValidateNotificationsConfig(config.Notifications)
	config.Notifications = notificationsConfig

	if isAuthValid {
		Global.Auth = config.Auth
	}

	if isLoggingValid {
		Global.Logging = config.Logging
	}

	if isMediaServerValid {
		Global.MediaServer = config.MediaServer
	}

	if isAutoDownloadValid {
		Global.AutoDownload = config.AutoDownload
	}

	if isMediuxValid {
		Global.Mediux = config.Mediux
	}

	if isNotificationsValid {
		Global.Notifications = config.Notifications
	}

	if !isAuthValid || !isLoggingValid || !isMediaServerValid || !isAutoDownloadValid || !isMediuxValid || !isNotificationsValid {
		logging.LOG.Error("Invalid configuration file. See errors above.")
		ConfigValid = false
		return
	}

	ConfigValid = true
}

func PrintConfig() {
	logging.LOG.NoTime("Current Configuration Settings\n")

	// Auth Mode
	logging.LOG.NoTime("\tAuth Mode:\n")
	logging.LOG.NoTime(fmt.Sprintf("\t\tEnabled: %t\n", Global.Auth.Enabled))
	if Global.Auth.Enabled {
		logging.LOG.NoTime(fmt.Sprintf("\t\tPassword: %s\n", "***"+Global.Auth.Password[len(Global.Auth.Password)-4:]))
	}

	// Logging Configuration
	logging.LOG.NoTime(fmt.Sprintf("\tLogging Level: %s\n", Global.Logging.Level))

	// Media Server Configuration
	logging.LOG.NoTime("\tMedia Server\n")
	logging.LOG.NoTime(fmt.Sprintf("\t\tType: %s\n", Global.MediaServer.Type))
	logging.LOG.NoTime(fmt.Sprintf("\t\tURL: %s\n", Global.MediaServer.URL))
	logging.LOG.NoTime(fmt.Sprintf("\t\tToken: %s\n", "***"+Global.MediaServer.Token[len(Global.MediaServer.Token)-4:]))
	logging.LOG.NoTime(fmt.Sprintf("\t\tUserID: %s\n", Global.MediaServer.UserID))
	logging.LOG.NoTime("\t\tLibraries:\n")
	for _, library := range Global.MediaServer.Libraries {
		logging.LOG.NoTime(fmt.Sprintf("\t\t\tName: %s\n", library.Name))
	}
	if Global.MediaServer.Type == "Plex" {
		logging.LOG.NoTime(fmt.Sprintf("\t\tSeason Naming Convention: %s\n", Global.MediaServer.SeasonNamingConvention))
	}

	// Mediux Configuration
	logging.LOG.NoTime("\tMediux\n")
	logging.LOG.NoTime(fmt.Sprintf("\t\tToken: %s\n", "***"+Global.Mediux.Token[len(Global.Mediux.Token)-4:]))
	logging.LOG.NoTime(fmt.Sprintf("\t\tDownload Quality: %s\n", Global.Mediux.DownloadQuality))

	// Auto Download Configuration
	logging.LOG.NoTime("\tAuto Download\n")
	logging.LOG.NoTime(fmt.Sprintf("\t\tEnabled: %t\n", Global.AutoDownload.Enabled))
	logging.LOG.NoTime(fmt.Sprintf("\t\tCron: %s\n", Global.AutoDownload.Cron))

	// Cache Images and Save Image Next To Content
	logging.LOG.NoTime("\tImages Options\n")
	logging.LOG.NoTime(fmt.Sprintf("\t\tCache Images: %t\n", Global.Images.CacheImages))
	logging.LOG.NoTime(fmt.Sprintf("\t\tSave Image Locally: %t\n", Global.Images.SaveImagesLocally.Enabled))
	if Global.Images.SaveImagesLocally.Enabled {
		if Global.Images.SaveImagesLocally.Path != "" {
			logging.LOG.NoTime(fmt.Sprintf("\t\tSave Image Locally Path: %s\n", Global.Images.SaveImagesLocally.Path))
		} else {
			logging.LOG.NoTime("\t\tSave Image Locally Path: Saving next to content\n")
		}
	}

	// TMDB Configuration
	if Global.TMDB.ApiKey != "" {
		logging.LOG.NoTime("\tTMDB\n")
		logging.LOG.NoTime(fmt.Sprintf("\t\tAPI Key: %s\n", "***"+Global.TMDB.ApiKey[len(Global.TMDB.ApiKey)-4:]))
	}

	// Labels and Tags Configuration
	logging.LOG.NoTime("\tLabels and Tags\n")
	for _, application := range Global.LabelsAndTags.Applications {
		logging.LOG.NoTime(fmt.Sprintf("\t\tApplication: %s\n", application.Application))
		logging.LOG.NoTime(fmt.Sprintf("\t\t\tEnabled: %t\n", application.Enabled))
		if application.Enabled {
			logging.LOG.NoTime(fmt.Sprintf("\t\t\tAdd: %v\n", application.Add))
			logging.LOG.NoTime(fmt.Sprintf("\t\t\tRemove: %v\n", application.Remove))
		}
	}

	// Notification Configuration
	logging.LOG.NoTime("\tNotifications\n")
	logging.LOG.NoTime(fmt.Sprintf("\t\tEnabled: %t\n", Global.Notifications.Enabled))
	for _, notification := range Global.Notifications.Providers {
		logging.LOG.NoTime(fmt.Sprintf("\t\tProvider: %s\n", notification.Provider))
		switch notification.Provider {
		case "Discord":
			logging.LOG.NoTime(fmt.Sprintf("\t\t\tEnabled: %t\n", notification.Enabled))
			logging.LOG.NoTime(fmt.Sprintf("\t\t\tWebhook: %s\n", masking.MaskWebhookURL(notification.Discord.Webhook)))
		case "Pushover":
			logging.LOG.NoTime(fmt.Sprintf("\t\t\tEnabled: %t\n", notification.Enabled))
			logging.LOG.NoTime(fmt.Sprintf("\t\t\tToken: %s\n", "***"+notification.Pushover.Token[len(notification.Pushover.Token)-4:]))
			logging.LOG.NoTime(fmt.Sprintf("\t\t\tUserKey: %s\n", "***"+notification.Pushover.UserKey[len(notification.Pushover.UserKey)-4:]))
		case "Gotify":
			logging.LOG.NoTime(fmt.Sprintf("\t\t\tEnabled: %t\n", notification.Enabled))
			logging.LOG.NoTime(fmt.Sprintf("\t\t\tURL: %s\n", notification.Gotify.URL))
			logging.LOG.NoTime(fmt.Sprintf("\t\t\tToken: %s\n", "***"+notification.Gotify.Token[len(notification.Gotify.Token)-4:]))
		}
	}

}

// contains checks if a string is present in a slice of strings.
func contains(slice []string, item string) bool {
	return slices.Contains(slice, item)
}

func ValidateAuthConfig(Auth modals.Config_Auth) (bool, string) {
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

func ValidateLoggingConfig(Logging modals.Config_Logging) (bool, string, modals.Config_Logging) {
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

func ValidateMediaServerConfig(MediaServer modals.Config_MediaServer) (bool, []string, modals.Config_MediaServer) {
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

func ValidateMediuxConfig(Mediux modals.Config_Mediux) (bool, []string, modals.Config_Mediux) {
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

func ValidateAutoDownloadConfig(Autodownload modals.Config_AutoDownload) (bool, []string, modals.Config_AutoDownload) {
	isValid := true
	var errorMsgs []string

	if !Autodownload.Enabled {
		return isValid, errorMsgs, Autodownload
	}

	if Autodownload.Cron == "" {
		logging.LOG.Warn("\tAutoDownload.Cron is not set, using default value: 0 0 * * *")
		Autodownload.Cron = "0 0 * * *" // Default to daily at midnight
	}

	if !validateCronExpression(Autodownload.Cron) {
		errorMsg := fmt.Sprintf("\tBad AutoDownload.Cron: '%s'. Must be a valid cron expression. Use something like https://crontab.guru to help you.", Autodownload.Cron)
		logging.LOG.Error(errorMsg)
		errorMsgs = append(errorMsgs, errorMsg)
		isValid = false
	}

	return isValid, errorMsgs, Autodownload
}

func validateCronExpression(cronExpression string) bool {
	_, err := cron.ParseStandard(cronExpression)
	return err == nil
}

func ValidateNotificationsConfig(Notifications modals.Config_Notifications) (bool, []string, modals.Config_Notifications) {
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
		if !contains(validProviders, provider.Provider) {
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
