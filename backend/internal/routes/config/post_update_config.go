package route_config

import (
	"aura/internal/config"
	"aura/internal/logging"
	"aura/internal/modals"
	"aura/internal/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"time"
)

func UpdateConfig(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	logging.LOG.Trace(r.URL.Path)
	Err := logging.NewStandardError()

	var newConfig modals.Config

	// Get the new config from the request
	if err := json.NewDecoder(r.Body).Decode(&newConfig); err != nil {
		Err.Message = "Failed to decode request body"
		Err.HelpText = "Ensure the request body is valid JSON"
		Err.Details = fmt.Sprintf("Error: %v", err)
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}

	// Set Config Valid to true
	oldConfigValidValue := config.ConfigValid
	config.ConfigValid = true

	authChanged, authValid, authErrorMessages := CheckConfigDifferences_Auth(config.Global.Auth, newConfig.Auth)
	loggingChanged, loggingValid, loggingErrorMessages, loggingConfig := CheckConfigDifferences_Logging(config.Global.Logging, newConfig.Logging)
	mediaServerChanged, mediaServerValid, mediaServerErrorMessages, mediaServerConfig := CheckConfigDifferences_MediaServer(config.Global.MediaServer, newConfig.MediaServer)
	mediuxChanged, mediuxValid, mediuxErrorMessages, mediuxConfig := CheckConfigDifferences_Mediux(config.Global.Mediux, newConfig.Mediux)
	autodownloadChanged, autodownloadValid, autodownloadErrorMessages, autodownloadConfig := CheckConfigDifferences_Autodownload(config.Global.AutoDownload, newConfig.AutoDownload)
	imagesChanged := CheckConfigDifferences_Images(config.Global.Images, newConfig.Images)
	tmdbChanged := CheckConfigDifferences_TMDB(config.Global.TMDB, newConfig.TMDB)
	kometaChanged := CheckConfigDifferences_Kometa(config.Global.Kometa, newConfig.Kometa)
	notificationsChanged, notificationsValid, notificationsErrorMessages, notificationsConfig := CheckConfigDifferences_Notifications(config.Global.Notifications, newConfig.Notifications)

	if authChanged {
		logging.LOG.Info("Auth configuration changes detected")
	}
	if !authValid {
		Err.Message = "Auth configuration is invalid"
		Err.Details = authErrorMessages
		Err.HelpText = "Please correct the Auth configuration and try again."
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		config.ConfigValid = oldConfigValidValue
		return
	}

	if loggingChanged {
		logging.LOG.Info("Logging configuration changes detected")
	}
	if !loggingValid {
		Err.Message = "Logging configuration is invalid"
		Err.Details = loggingErrorMessages
		Err.HelpText = "Please correct the Logging configuration and try again."
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		config.ConfigValid = oldConfigValidValue
		return
	}

	if mediaServerChanged {
		logging.LOG.Info("MediaServer configuration changes detected")
	}
	if !mediaServerValid {
		Err.Message = "MediaServer configuration is invalid"
		Err.Details = mediaServerErrorMessages
		Err.HelpText = "Please correct the MediaServer configuration and try again."
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		config.ConfigValid = oldConfigValidValue
		return
	}

	if mediuxChanged {
		logging.LOG.Info("Mediux configuration changes detected")
	}
	if !mediuxValid {
		Err.Message = "Mediux configuration is invalid"
		Err.Details = mediuxErrorMessages
		Err.HelpText = "Please correct the Mediux configuration and try again."
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		config.ConfigValid = oldConfigValidValue
		return
	}

	if autodownloadChanged {
		logging.LOG.Info("AutoDownload configuration changes detected")
	}
	if !autodownloadValid {
		Err.Message = "AutoDownload configuration is invalid"
		Err.Details = autodownloadErrorMessages
		Err.HelpText = "Please correct the AutoDownload configuration and try again."
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		config.ConfigValid = oldConfigValidValue
		return
	}

	if imagesChanged {
		logging.LOG.Info("Images configuration changes detected")
	}

	if tmdbChanged {
		logging.LOG.Info("TMDB configuration changes detected")
	}

	if kometaChanged {
		logging.LOG.Info("Kometa configuration changes detected")
	}

	if notificationsChanged {
		logging.LOG.Info("Notifications configuration changes detected")
	}
	if !notificationsValid {
		Err.Message = "Notifications configuration is invalid"
		Err.Details = notificationsErrorMessages
		Err.HelpText = "Please correct the Notifications configuration and try again."
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		config.ConfigValid = oldConfigValidValue
		return
	}

	if config.ConfigLoaded && config.ConfigValid {
		if !authChanged && !loggingChanged && !mediaServerChanged && !mediuxChanged && !autodownloadChanged && !imagesChanged && !tmdbChanged && !kometaChanged && !notificationsChanged {
			// No changes detected
			utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
				Status:  "warn",
				Elapsed: utils.ElapsedTime(startTime),
				Data:    "No configuration changes detected",
			})
			config.ConfigValid = oldConfigValidValue
			return
		}
		logging.LOG.Info("Configuration changes detected and validated, saving...")
	} else {
		logging.LOG.Info("New configuration is valid, saving...")
	}

	// Apply validated sub-configs to the new config
	newConfig.Logging = loggingConfig
	newConfig.MediaServer = mediaServerConfig
	newConfig.Mediux = mediuxConfig
	newConfig.AutoDownload = autodownloadConfig
	newConfig.Notifications = notificationsConfig

	// Save the new configuration
	config.UpdateConfig(newConfig)

	utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
		Status:  "success",
		Elapsed: utils.ElapsedTime(startTime),
		Data:    utils.SanitizedCopy(),
	})
}

func CheckConfigDifferences_Auth(oldAuth, newAuth modals.Config_Auth) (bool, bool, string) {
	changed := false
	newValid := false
	if config.ConfigLoaded && config.ConfigValid {
		if !reflect.DeepEqual(oldAuth, newAuth) {
			if oldAuth.Enabled != newAuth.Enabled {
				logging.LOG.Info(fmt.Sprintf("Auth.Enabled changed: '%v' -> '%v'", oldAuth.Enabled, newAuth.Enabled))
				changed = true
			}

			if oldAuth.Password != newAuth.Password {
				logging.LOG.Info(fmt.Sprintf("Auth.Password changed: '%v' -> '%v'", oldAuth.Password, newAuth.Password))
				changed = true
			}
		}
	}
	newValid, errorMessages := config.ValidateAuthConfig(newAuth)
	return changed, newValid, errorMessages
}

func CheckConfigDifferences_Logging(oldLogging, newLogging modals.Config_Logging) (bool, bool, string, modals.Config_Logging) {
	changed := false
	newValid := false
	if config.ConfigLoaded && config.ConfigValid {
		if !reflect.DeepEqual(oldLogging, newLogging) {
			if oldLogging.Level != newLogging.Level {
				logging.LOG.Info(fmt.Sprintf("Logging.Level changed: '%v' -> '%v'", oldLogging.Level, newLogging.Level))
				changed = true
			}
		}
	}
	newValid, errorMessages, newLogging := config.ValidateLoggingConfig(newLogging)
	return changed, newValid, errorMessages, newLogging
}

func CheckConfigDifferences_MediaServer(oldMediaServer, newMediaServer modals.Config_MediaServer) (bool, bool, []string, modals.Config_MediaServer) {
	changed := false
	newValid := false
	if config.ConfigLoaded && config.ConfigValid {
		if !reflect.DeepEqual(oldMediaServer, newMediaServer) {
			if oldMediaServer.Type != newMediaServer.Type {
				logging.LOG.Info(fmt.Sprintf("MediaServer.Type changed: '%v' -> '%v'", oldMediaServer.Type, newMediaServer.Type))
				changed = true
			}

			if oldMediaServer.URL != newMediaServer.URL {
				logging.LOG.Info(fmt.Sprintf("MediaServer.URL changed: '%v' -> '%v'", oldMediaServer.URL, newMediaServer.URL))
				changed = true
			}

			if oldMediaServer.Token != newMediaServer.Token {
				if !strings.HasPrefix(newMediaServer.Token, "***") {
					logging.LOG.Info(fmt.Sprintf("MediaServer.Token changed: '%v' -> '%v'", oldMediaServer.Token, newMediaServer.Token))
					changed = true
				} else {
					newMediaServer.Token = oldMediaServer.Token
				}
			}

			if !reflect.DeepEqual(oldMediaServer.Libraries, newMediaServer.Libraries) {
				oldNames := libraryNames(oldMediaServer.Libraries)
				newNames := libraryNames(newMediaServer.Libraries)
				logging.LOG.Info(fmt.Sprintf("MediaServer.Libraries changed: '%s' -> '%s'", oldNames, newNames))
				changed = true
			}

			if oldMediaServer.SeasonNamingConvention != newMediaServer.SeasonNamingConvention {
				logging.LOG.Info(fmt.Sprintf("MediaServer.SeasonNamingConvention changed: '%v' -> '%v'", oldMediaServer.SeasonNamingConvention, newMediaServer.SeasonNamingConvention))
				changed = true
			}

			if oldMediaServer.UserID != newMediaServer.UserID {
				logging.LOG.Info(fmt.Sprintf("MediaServer.UserID changed: '%v' -> '%v'", oldMediaServer.UserID, newMediaServer.UserID))
				changed = true
			}
		}
	}
	newValid, errorMessages, newMediaServer := config.ValidateMediaServerConfig(newMediaServer)
	return changed, newValid, errorMessages, newMediaServer
}

func CheckConfigDifferences_Mediux(oldMediux, newMediux modals.Config_Mediux) (bool, bool, []string, modals.Config_Mediux) {
	changed := false
	newValid := false
	if config.ConfigLoaded && config.ConfigValid {
		if !reflect.DeepEqual(oldMediux, newMediux) {
			if oldMediux.Token != newMediux.Token {
				if !strings.HasPrefix(newMediux.Token, "***") {
					logging.LOG.Info(fmt.Sprintf("Mediux.Token changed: '%v' -> '%v'", oldMediux.Token, newMediux.Token))
					changed = true
				} else {
					newMediux.Token = oldMediux.Token
				}
			}

			if oldMediux.DownloadQuality != newMediux.DownloadQuality {
				logging.LOG.Info(fmt.Sprintf("Mediux.DownloadQuality changed: '%v' -> '%v'", oldMediux.DownloadQuality, newMediux.DownloadQuality))
				changed = true
			}
		}
	}
	newValid, errorMessages, newMediux := config.ValidateMediuxConfig(newMediux)
	return changed, newValid, errorMessages, newMediux
}

func CheckConfigDifferences_Autodownload(oldAutodownload, newAutodownload modals.Config_AutoDownload) (bool, bool, []string, modals.Config_AutoDownload) {
	changed := false
	newValid := false
	if config.ConfigLoaded && config.ConfigValid {
		if !reflect.DeepEqual(oldAutodownload, newAutodownload) {
			if oldAutodownload.Enabled != newAutodownload.Enabled {
				logging.LOG.Info(fmt.Sprintf("Autodownload.Enabled changed: '%v' -> '%v'", oldAutodownload.Enabled, newAutodownload.Enabled))
				changed = true
			}

			if oldAutodownload.Cron != newAutodownload.Cron {
				logging.LOG.Info(fmt.Sprintf("Autodownload.Cron changed: '%v' -> '%v'", oldAutodownload.Cron, newAutodownload.Cron))
				changed = true
			}
		}
	}
	newValid, errorMessages, newAutodownload := config.ValidateAutoDownloadConfig(newAutodownload)
	return changed, newValid, errorMessages, newAutodownload
}

func CheckConfigDifferences_Images(oldImages, newImages modals.Config_Images) bool {
	changed := false
	if config.ConfigLoaded && config.ConfigValid {
		if !reflect.DeepEqual(oldImages, newImages) {
			if oldImages.CacheImages.Enabled != newImages.CacheImages.Enabled {
				logging.LOG.Info(fmt.Sprintf("Images.CacheImages changed: '%v' -> '%v'", oldImages.CacheImages.Enabled, newImages.CacheImages.Enabled))
				changed = true
			}

			if oldImages.SaveImageNextToContent.Enabled != newImages.SaveImageNextToContent.Enabled {
				logging.LOG.Info(fmt.Sprintf("Images.SaveImageNextToContent changed: '%v' -> '%v'", oldImages.SaveImageNextToContent.Enabled, newImages.SaveImageNextToContent.Enabled))
				changed = true
			}
		}
	}
	return changed
}

func CheckConfigDifferences_TMDB(oldTMDB, newTMDB modals.Config_TMDB) bool {
	changed := false
	if config.ConfigLoaded && config.ConfigValid {
		if !reflect.DeepEqual(oldTMDB, newTMDB) {
			if oldTMDB.ApiKey != newTMDB.ApiKey {
				if !strings.HasPrefix(newTMDB.ApiKey, "***") {
					logging.LOG.Info(fmt.Sprintf("TMDB.ApiKey changed: '%v' -> '%v'", oldTMDB.ApiKey, newTMDB.ApiKey))
					changed = true
				} else {
					newTMDB.ApiKey = oldTMDB.ApiKey
				}
			}
		}
	}
	return changed
}

func CheckConfigDifferences_Kometa(oldKometa, newKometa modals.Config_Kometa) bool {
	changed := false
	if config.ConfigLoaded && config.ConfigValid {
		if !reflect.DeepEqual(oldKometa, newKometa) {
			if oldKometa.RemoveLabels != newKometa.RemoveLabels {
				logging.LOG.Info(fmt.Sprintf("Kometa.RemoveLabels changed: '%v' -> '%v'", oldKometa.RemoveLabels, newKometa.RemoveLabels))
				changed = true
			}

			if !reflect.DeepEqual(oldKometa.Labels, newKometa.Labels) {
				logging.LOG.Info(fmt.Sprintf(
					"Kometa.Labels changed: '%s' -> '%s'",
					joinNonEmptyComma(oldKometa.Labels),
					joinNonEmptyComma(newKometa.Labels),
				))
				changed = true
			}
		}
	}
	return changed
}

func CheckConfigDifferences_Notifications(oldNotifications, newNotifications modals.Config_Notifications) (bool, bool, []string, modals.Config_Notifications) {
	changed := false
	newValid := false
	if config.ConfigLoaded && config.ConfigValid {
		if !reflect.DeepEqual(oldNotifications, newNotifications) {
			// Global toggle
			if oldNotifications.Enabled != newNotifications.Enabled {
				logging.LOG.Info(fmt.Sprintf(
					"Notifications.Enabled changed: '%v' -> '%v'",
					oldNotifications.Enabled, newNotifications.Enabled,
				))
				changed = true
			}

			// Providers diff
			oldMap := providerMap(oldNotifications.Providers)
			newMap := providerMap(newNotifications.Providers)

			// Added / removed types
			var added, removed []string
			for k := range oldMap {
				if _, ok := newMap[k]; !ok {
					removed = append(removed, k)
				}
			}
			for k := range newMap {
				if _, ok := oldMap[k]; !ok {
					added = append(added, k)
				}
			}
			if len(added) > 0 {
				sort.Strings(added)
				logging.LOG.Info(fmt.Sprintf("Notifications.Providers added: %s", joinNonEmptyComma(added)))
				changed = true
			}
			if len(removed) > 0 {
				sort.Strings(removed)
				logging.LOG.Info(fmt.Sprintf("Notifications.Providers removed: %s", joinNonEmptyComma(removed)))
				changed = true
			}

			// Compare providers present in both
			for name, oldProv := range oldMap {
				newProv, ok := newMap[name]
				if !ok {
					continue
				}

				// Per-provider enabled
				if oldProv.Enabled != newProv.Enabled {
					logging.LOG.Info(fmt.Sprintf(
						"Notifications.Provider[%s].Enabled changed: '%v' -> '%v'",
						name, oldProv.Enabled, newProv.Enabled,
					))
					changed = true
				}

				switch name {
				case "Discord":
					var oldWebhook, newWebhook string
					if oldProv.Discord != nil {
						oldWebhook = strings.TrimSpace(oldProv.Discord.Webhook)
					}
					if newProv.Discord != nil {
						newWebhook = strings.TrimSpace(newProv.Discord.Webhook)
					}
					if oldWebhook != newWebhook {
						if !isMaskedWebhook(newWebhook) {
							logging.LOG.Info(fmt.Sprintf(
								"Notifications.Discord.Webhook changed: '%s' -> '%s'", oldWebhook, newWebhook))
							changed = true
						} else {
							newProv.Discord.Webhook = oldProv.Discord.Webhook
						}
					}

				case "Pushover":
					var oldToken, oldUserKey, newToken, newUserKey string
					if oldProv.Pushover != nil {
						oldToken = strings.TrimSpace(oldProv.Pushover.Token)
						oldUserKey = strings.TrimSpace(oldProv.Pushover.UserKey)
					}
					if newProv.Pushover != nil {
						newToken = strings.TrimSpace(newProv.Pushover.Token)
						newUserKey = strings.TrimSpace(newProv.Pushover.UserKey)
					}
					if oldUserKey != newUserKey {
						if !strings.HasPrefix(newUserKey, "***") {
							logging.LOG.Info(fmt.Sprintf(
								"Notifications.Pushover.UserKey changed: '%s' -> '%s'", oldUserKey, newUserKey))
							changed = true
						} else {
							newProv.Pushover.UserKey = oldProv.Pushover.UserKey
						}

					}
					if oldToken != newToken {
						if !strings.HasPrefix(newToken, "***") {
							logging.LOG.Info(fmt.Sprintf(
								"Notifications.Pushover.Token changed: '%s' -> '%s'", oldToken, newToken))
							changed = true
						} else {
							newProv.Pushover.Token = oldProv.Pushover.Token
						}

					}
				default:
					// Unknown provider type: nothing more to compare
				}
			}
		}
	}
	newValid, errorMessages, newNotifications := config.ValidateNotificationsConfig(newNotifications)
	return changed, newValid, errorMessages, newNotifications
}

func libraryNames(libs []modals.Config_MediaServerLibrary) string {
	names := make([]string, 0, len(libs))
	for _, l := range libs {
		n := strings.TrimSpace(l.Name)
		if n != "" {
			names = append(names, n)
		}
	}
	return joinNonEmptyComma(names)
}

func joinNonEmptyComma(items []string) string {
	out := make([]string, 0, len(items))
	for _, s := range items {
		if t := strings.TrimSpace(s); t != "" {
			out = append(out, t)
		}
	}
	if len(out) == 0 {
		return "(none)"
	}
	return strings.Join(out, ", ")
}

func providerMap(items []modals.Config_Notification_Providers) map[string]modals.Config_Notification_Providers {
	m := make(map[string]modals.Config_Notification_Providers, len(items))
	for _, p := range items {
		if p.Provider == "" {
			continue
		}
		m[p.Provider] = p
	}
	return m
}

var reMasked3 = regexp.MustCompile(`\*{4}[^/]{3}/\*{3}[^/]{3}$`)

func isMaskedWebhook(s string) bool {
	return reMasked3.MatchString(strings.TrimSpace(s))
}
