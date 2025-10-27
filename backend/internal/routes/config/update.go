package routes_config

import (
	"aura/internal/api"
	"aura/internal/logging"

	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"time"
)

// UpdateConfig handles requests to update the server configuration.
//
// Method: POST
//
// Endpoint: /api/config
//
// It accepts a JSON payload representing the new configuration, validates it, and applies the changes if valid.
//
// If the configuration is successfully updated, it responds with the sanitized configuration.
//
// If there are validation errors, it responds with an error message detailing the issues.
func UpdateConfig(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	logging.LOG.Trace(r.URL.Path)
	Err := logging.NewStandardError()

	var newConfig api.Config

	// Get the new config from the request
	if err := json.NewDecoder(r.Body).Decode(&newConfig); err != nil {
		Err.Message = "Failed to decode request body"
		Err.HelpText = "Ensure the request body is valid JSON"
		Err.Details = map[string]any{
			"error": err.Error(),
		}
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		return
	}

	// Set Config Valid to true
	oldConfigValidValue := api.Global_Config_Valid
	api.Global_Config_Valid = true

	authChanged, authValid, authErrorMessages := checkConfigDifferences_Auth(api.Global_Config.Auth, newConfig.Auth)
	loggingChanged, loggingValid, loggingErrorMessages, loggingConfig := checkConfigDifferences_Logging(api.Global_Config.Logging, newConfig.Logging)
	mediaServerChanged, mediaServerValid, mediaServerErrorMessages, mediaServerConfig := checkConfigDifferences_MediaServer(api.Global_Config.MediaServer, newConfig.MediaServer)
	mediuxChanged, mediuxValid, mediuxErrorMessages, mediuxConfig := checkConfigDifferences_Mediux(api.Global_Config.Mediux, newConfig.Mediux)
	autodownloadChanged, autodownloadValid, autodownloadErrorMessages, autodownloadConfig := checkConfigDifferences_Autodownload(api.Global_Config.AutoDownload, newConfig.AutoDownload)
	imagesChanged := checkConfigDifferences_Images(api.Global_Config.Images, newConfig.Images)
	tmdbChanged := checkConfigDifferences_TMDB(api.Global_Config.TMDB, newConfig.TMDB)
	labelsAndTagsChanged := checkConfigDifferences_LabelsAndTags(api.Global_Config.LabelsAndTags, newConfig.LabelsAndTags)
	notificationsChanged, notificationsValid, notificationsErrorMessages, notificationsConfig := checkConfigDifferences_Notifications(api.Global_Config.Notifications, newConfig.Notifications)
	srChanged, srValid, srErrorMessages, newSonarrRadarr := checkConfigDifferences_SonarrRadarr(api.Global_Config.SonarrRadarr, newConfig.SonarrRadarr, mediaServerConfig)

	if authChanged {
		logging.LOG.Info("Auth configuration changes detected")
	}
	if !authValid {
		Err.Message = "Auth configuration is invalid"
		Err.HelpText = "Please correct the Auth configuration and try again."
		Err.Details = map[string]any{
			"error": authErrorMessages,
		}
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		api.Global_Config_Valid = oldConfigValidValue
		return
	}

	if loggingChanged {
		logging.LOG.Info("Logging configuration changes detected")
	}
	if !loggingValid {
		Err.Message = "Logging configuration is invalid"
		Err.Details = map[string]any{
			"error": loggingErrorMessages,
		}
		Err.HelpText = "Please correct the Logging configuration and try again."
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		api.Global_Config_Valid = oldConfigValidValue
		return
	}

	if mediaServerChanged {
		logging.LOG.Info("MediaServer configuration changes detected")
	}
	if !mediaServerValid {
		Err.Message = "MediaServer configuration is invalid"
		Err.Details = map[string]any{
			"error": mediaServerErrorMessages,
		}
		Err.HelpText = "Please correct the MediaServer configuration and try again."
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		api.Global_Config_Valid = oldConfigValidValue
		return
	}

	if mediuxChanged {
		logging.LOG.Info("Mediux configuration changes detected")
	}
	if !mediuxValid {
		Err.Message = "Mediux configuration is invalid"
		Err.Details = map[string]any{
			"error": mediuxErrorMessages,
		}
		Err.HelpText = "Please correct the Mediux configuration and try again."
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		api.Global_Config_Valid = oldConfigValidValue
		return
	}

	if autodownloadChanged {
		logging.LOG.Info("AutoDownload configuration changes detected")
	}
	if !autodownloadValid {
		Err.Message = "AutoDownload configuration is invalid"
		Err.Details = map[string]any{
			"error": autodownloadErrorMessages,
		}
		Err.HelpText = "Please correct the AutoDownload configuration and try again."
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		api.Global_Config_Valid = oldConfigValidValue
		return
	}

	if imagesChanged {
		logging.LOG.Info("Images configuration changes detected")
	}

	if tmdbChanged {
		logging.LOG.Info("TMDB configuration changes detected")
	}

	if labelsAndTagsChanged {
		logging.LOG.Info("LabelsAndTags configuration changes detected")
	}

	if notificationsChanged {
		logging.LOG.Info("Notifications configuration changes detected")
	}
	if !notificationsValid {
		Err.Message = "Notifications configuration is invalid"
		Err.Details = map[string]any{
			"error": notificationsErrorMessages,
		}
		Err.HelpText = "Please correct the Notifications configuration and try again."
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		api.Global_Config_Valid = oldConfigValidValue
		return
	}

	if srChanged {
		logging.LOG.Info("Sonarr/Radarr configuration changes detected")
	}
	if !srValid {
		Err.Message = "Sonarr/Radarr configuration is invalid"
		Err.Details = map[string]any{
			"error": srErrorMessages,
		}
		Err.HelpText = "Please correct the Sonarr/Radarr configuration and try again."
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		api.Global_Config_Valid = oldConfigValidValue
		return
	}

	if api.Global_Config_Loaded && api.Global_Config_Valid {
		if !authChanged && !loggingChanged && !mediaServerChanged &&
			!mediuxChanged && !autodownloadChanged && !imagesChanged &&
			!tmdbChanged && !labelsAndTagsChanged && !notificationsChanged && !srChanged {

			// No changes detected
			api.Util_Response_SendJson(w, http.StatusOK, api.JSONResponse{
				Status:  "warn",
				Elapsed: api.Util_ElapsedTime(startTime),
				Data:    "No configuration changes detected",
			})
			api.Global_Config_Valid = oldConfigValidValue
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
	newConfig.SonarrRadarr = newSonarrRadarr

	// Save the new configuration
	newConfig.Config_UpdateConfig()

	api.Util_Response_SendJson(w, http.StatusOK, api.JSONResponse{
		Status:  "success",
		Elapsed: api.Util_ElapsedTime(startTime),
		Data:    api.Global_Config.Sanitize(),
	})
}

// checkConfigDifferences_Auth compares old and new Auth configurations.
func checkConfigDifferences_Auth(oldAuth, newAuth api.Config_Auth) (bool, bool, string) {
	changed := false
	newValid := false
	if api.Global_Config_Loaded && api.Global_Config_Valid {
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
	newValid, errorMessages := api.Config_ValidateAuth(newAuth)
	return changed, newValid, errorMessages
}

// checkConfigDifferences_Logging compares old and new Logging configurations.
func checkConfigDifferences_Logging(oldLogging, newLogging api.Config_Logging) (bool, bool, string, api.Config_Logging) {
	changed := false
	newValid := false
	if api.Global_Config_Loaded && api.Global_Config_Valid {
		if !reflect.DeepEqual(oldLogging, newLogging) {
			if oldLogging.Level != newLogging.Level {
				logging.LOG.Info(fmt.Sprintf("Logging.Level changed: '%v' -> '%v'", oldLogging.Level, newLogging.Level))
				changed = true
			}
		}
	}
	newValid, errorMessages, newLogging := api.Config_ValidateLogging(newLogging)
	return changed, newValid, errorMessages, newLogging
}

// checkConfigDifferences_MediaServer compares old and new MediaServer configurations.
func checkConfigDifferences_MediaServer(oldMediaServer, newMediaServer api.Config_MediaServer) (bool, bool, []string, api.Config_MediaServer) {
	changed := false
	newValid := false
	if api.Global_Config_Loaded && api.Global_Config_Valid {
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
	newValid, errorMessages, newMediaServer := api.Config_ValidateMediaServer(newMediaServer)
	return changed, newValid, errorMessages, newMediaServer
}

// checkConfigDifferences_Mediux compares old and new Mediux configurations.
func checkConfigDifferences_Mediux(oldMediux, newMediux api.Config_Mediux) (bool, bool, []string, api.Config_Mediux) {
	changed := false
	newValid := false
	if api.Global_Config_Loaded && api.Global_Config_Valid {
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
	newValid, errorMessages, newMediux := api.Config_ValidateMediux(newMediux)
	return changed, newValid, errorMessages, newMediux
}

// checkConfigDifferences_Autodownload compares old and new AutoDownload configurations.
func checkConfigDifferences_Autodownload(oldAutodownload, newAutodownload api.Config_AutoDownload) (bool, bool, []string, api.Config_AutoDownload) {
	changed := false
	newValid := false
	if api.Global_Config_Loaded && api.Global_Config_Valid {
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
	newValid, errorMessages, newAutodownload := api.Config_ValidateAutoDownload(newAutodownload)
	return changed, newValid, errorMessages, newAutodownload
}

// checkConfigDifferences_Images compares old and new Images configurations.
func checkConfigDifferences_Images(oldImages, newImages api.Config_Images) bool {
	changed := false
	if api.Global_Config_Loaded && api.Global_Config_Valid {
		if !reflect.DeepEqual(oldImages, newImages) {
			if oldImages.CacheImages.Enabled != newImages.CacheImages.Enabled {
				logging.LOG.Info(fmt.Sprintf("Images.CacheImages changed: '%v' -> '%v'", oldImages.CacheImages.Enabled, newImages.CacheImages.Enabled))
				changed = true
			}

			if oldImages.SaveImagesLocally.Enabled != newImages.SaveImagesLocally.Enabled {
				logging.LOG.Info(fmt.Sprintf("Images.SaveImagesLocally changed: '%v' -> '%v'", oldImages.SaveImagesLocally.Enabled, newImages.SaveImagesLocally.Enabled))
				changed = true
			}

			if oldImages.SaveImagesLocally.Path != newImages.SaveImagesLocally.Path {
				logging.LOG.Info(fmt.Sprintf("Images.SaveImagesLocally.Path changed: '%v' -> '%v'", oldImages.SaveImagesLocally.Path, newImages.SaveImagesLocally.Path))
				changed = true
			}
		}
	}
	return changed
}

// checkConfigDifferences_TMDB compares old and new TMDB configurations.
func checkConfigDifferences_TMDB(oldTMDB, newTMDB api.Config_TMDB) bool {
	changed := false
	if api.Global_Config_Loaded && api.Global_Config_Valid {
		if !reflect.DeepEqual(oldTMDB, newTMDB) {
			if oldTMDB.APIKey != newTMDB.APIKey {
				if !strings.HasPrefix(newTMDB.APIKey, "***") {
					logging.LOG.Info(fmt.Sprintf("TMDB.APIKey changed: '%v' -> '%v'", oldTMDB.APIKey, newTMDB.APIKey))
					changed = true
				} else {
					newTMDB.APIKey = oldTMDB.APIKey
				}
			}
		}
	}
	return changed
}

// checkConfigDifferences_LabelsAndTags compares old and new LabelsAndTags configurations.
func checkConfigDifferences_LabelsAndTags(oldLAT, newLAT api.Config_LabelsAndTags) bool {
	changed := false
	if api.Global_Config_Loaded && api.Global_Config_Valid {
		if !reflect.DeepEqual(oldLAT, newLAT) {

			// Applications diff
			oldMap := applicationMapLabelsAndTags(oldLAT.Applications)
			newMap := applicationMapLabelsAndTags(newLAT.Applications)

			// Added / removed apps
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
				logging.LOG.Info(fmt.Sprintf("LabelsAndTags.Applications added: %s", joinNonEmptyComma(added)))
				changed = true
			}
			if len(removed) > 0 {
				sort.Strings(removed)
				logging.LOG.Info(fmt.Sprintf("LabelsAndTags.Applications removed: %s", joinNonEmptyComma(removed)))
				changed = true
			}

			// Compare applications present in both
			for name, oldApp := range oldMap {
				newApp, ok := newMap[name]
				if !ok {
					continue
				}

				// Per-application enabled
				if oldApp.Enabled != newApp.Enabled {
					logging.LOG.Info(fmt.Sprintf(
						"LabelsAndTags.Application[%s].Enabled changed: '%v' -> '%v'",
						name, oldApp.Enabled, newApp.Enabled,
					))
					changed = true
				}

				// Add list diff
				var addAdded, addRemoved []string
				oldAddMap := make(map[string]bool)
				for _, v := range oldApp.Add {
					oldAddMap[v] = true
				}
				newAddMap := make(map[string]bool)
				for _, v := range newApp.Add {
					newAddMap[v] = true
				}
				for k := range oldAddMap {
					if _, ok := newAddMap[k]; !ok {
						addRemoved = append(addRemoved, k)
					}
				}
				for k := range newAddMap {
					if _, ok := oldAddMap[k]; !ok {
						addAdded = append(addAdded, k)
					}
				}
				if len(addAdded) > 0 {
					sort.Strings(addAdded)
					logging.LOG.Info(fmt.Sprintf("LabelsAndTags.Application[%s].Add added: %s", name, joinNonEmptyComma(addAdded)))
					changed = true
				}
				if len(addRemoved) > 0 {
					sort.Strings(addRemoved)
					logging.LOG.Info(fmt.Sprintf("LabelsAndTags.Application[%s].Add removed: %s", name, joinNonEmptyComma(addRemoved)))
					changed = true
				}

				// Remove list diff
				var removeAdded, removeRemoved []string
				oldRemoveMap := make(map[string]bool)
				for _, v := range oldApp.Remove {
					oldRemoveMap[v] = true
				}
				newRemoveMap := make(map[string]bool)
				for _, v := range newApp.Remove {
					newRemoveMap[v] = true
				}
				for k := range oldRemoveMap {
					if _, ok := newRemoveMap[k]; !ok {
						removeRemoved = append(removeRemoved, k)
					}
				}
				for k := range newRemoveMap {
					if _, ok := oldRemoveMap[k]; !ok {
						removeAdded = append(removeAdded, k)
					}
				}
				if len(removeAdded) > 0 {
					sort.Strings(removeAdded)
					logging.LOG.Info(fmt.Sprintf("LabelsAndTags.Application[%s].Remove added: %s", name, joinNonEmptyComma(removeAdded)))
					changed = true
				}
				if len(removeRemoved) > 0 {
					sort.Strings(removeRemoved)
					logging.LOG.Info(fmt.Sprintf("LabelsAndTags.Application[%s].Remove removed: %s", name, joinNonEmptyComma(removeRemoved)))
					changed = true
				}
			}
		}
	}
	return changed
}

// checkConfigDifferences_Notifications compares old and new Notifications configurations.
func checkConfigDifferences_Notifications(oldNotifications, newNotifications api.Config_Notifications) (bool, bool, []string, api.Config_Notifications) {
	changed := false
	newValid := false
	if api.Global_Config_Loaded && api.Global_Config_Valid {
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
			oldMap := providerMapNotifications(oldNotifications.Providers)
			newMap := providerMapNotifications(newNotifications.Providers)

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
						if !IsMaskedWebhook(newWebhook) {
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

				case "Gotify":
					var oldGotifyURL, newGotifyURL string
					if oldProv.Gotify != nil {
						oldGotifyURL = strings.TrimSpace(oldProv.Gotify.URL)
					}
					if newProv.Gotify != nil {
						newGotifyURL = strings.TrimSpace(newProv.Gotify.URL)
					}
					// URL is never masked; any difference is a real change
					if oldGotifyURL != newGotifyURL {
						logging.LOG.Info(fmt.Sprintf(
							"Notifications.Gotify.URL changed: '%s' -> '%s'", oldGotifyURL, newGotifyURL))
						changed = true
					}
					// Token may still arrive masked; keep existing mask logic
					if oldProv.Gotify != nil && newProv.Gotify != nil {
						if oldProv.Gotify.Token != newProv.Gotify.Token {
							if !IsMaskedWebhook(newProv.Gotify.Token) {
								logging.LOG.Info(fmt.Sprintf(
									"Notifications.Gotify.Token changed: '%s' -> '%s'",
									oldProv.Gotify.Token, newProv.Gotify.Token))
								changed = true
							} else {
								newProv.Gotify.Token = oldProv.Gotify.Token
							}
						}
					}
				default:
					// Unknown provider type: nothing more to compare
				}
			}
		}
	}
	newValid, errorMessages, newNotifications := api.Config_ValidateNotifications(newNotifications)
	return changed, newValid, errorMessages, newNotifications
}

// checkConfigDifferences_SonarrRadarr compares old and new Sonarr/Radarr configurations.
func checkConfigDifferences_SonarrRadarr(oldSonarrRadarr, newSonarrRadarr api.Config_SonarrRadarr_Apps, mediaServerConfig api.Config_MediaServer) (bool, bool, []string, api.Config_SonarrRadarr_Apps) {
	changed := false
	newValid := false
	if api.Global_Config_Loaded && api.Global_Config_Valid {
		if !reflect.DeepEqual(oldSonarrRadarr, newSonarrRadarr) {

			// Providers diff
			oldMap := applicationSonarrRadarr(oldSonarrRadarr.Applications)
			newMap := applicationSonarrRadarr(newSonarrRadarr.Applications)

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
				logging.LOG.Info(fmt.Sprintf("SonarrRadarr.Applications added: %s", joinNonEmptyComma(added)))
				changed = true
			}
			if len(removed) > 0 {
				sort.Strings(removed)
				logging.LOG.Info(fmt.Sprintf("SonarrRadarr.Applications removed: %s", joinNonEmptyComma(removed)))
				changed = true
			}

			// Compare providers present in both
			for name, oldProv := range oldMap {
				newProv, ok := newMap[name]
				if !ok {
					continue
				}

				// Per App - APIKey
				if oldProv.APIKey != newProv.APIKey {
					if !strings.HasPrefix(newProv.APIKey, "***") {
						logging.LOG.Info(fmt.Sprintf(
							"SonarrRadarr.Application[%s].APIKey changed: '%v' -> '%v'",
							name, oldProv.APIKey, newProv.APIKey,
						))
						changed = true
					} else {

						newProv.APIKey = oldProv.APIKey
					}
				}

				// Per App - URL
				if oldProv.URL != newProv.URL {
					logging.LOG.Info(fmt.Sprintf(
						"SonarrRadarr.Application[%s].URL changed: '%v' -> '%v'",
						name, oldProv.URL, newProv.URL,
					))
					changed = true
				}

				// Per App - Type
				if oldProv.Type != newProv.Type {
					logging.LOG.Info(fmt.Sprintf(
						"SonarrRadarr.Application[%s].Type changed: '%v' -> '%v'",
						name, oldProv.Type, newProv.Type,
					))
					changed = true
				}

				// Per App - Library
				if oldProv.Library != newProv.Library {
					logging.LOG.Info(fmt.Sprintf(
						"SonarrRadarr.Application[%s].Library changed: '%v' -> '%v'",
						name, oldProv.Library, newProv.Library,
					))
					changed = true
				}

			}
		}
	}

	// Restore API keys
	for i, app := range newSonarrRadarr.Applications {
		if strings.HasPrefix(app.APIKey, "***") {
			// Find matching old app by Library and Type
			for _, oldApp := range oldSonarrRadarr.Applications {
				if app.Library == oldApp.Library && app.Type == oldApp.Type {
					newSonarrRadarr.Applications[i].APIKey = oldApp.APIKey
					break
				}
			}
		}
	}

	newValid, errorMessages, newSonarrRadarr := api.Config_ValidateSonarrRadarr(newSonarrRadarr, mediaServerConfig)
	return changed, newValid, errorMessages, newSonarrRadarr
}

// libraryNames returns a comma-separated string of library names from the given slice.
func libraryNames(libs []api.Config_MediaServerLibrary) string {
	names := make([]string, 0, len(libs))
	for _, l := range libs {
		n := strings.TrimSpace(l.Name)
		if n != "" {
			names = append(names, n)
		}
	}
	return joinNonEmptyComma(names)
}

// joinNonEmptyComma joins non-empty trimmed strings with a comma, or returns "(none)" if all are empty.
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

// providerMapNotifications creates a map of notification providers keyed by provider name.
func providerMapNotifications(items []api.Config_Notification_Provider) map[string]api.Config_Notification_Provider {
	m := make(map[string]api.Config_Notification_Provider, len(items))
	for _, p := range items {
		if p.Provider == "" {
			continue
		}
		m[p.Provider] = p
	}
	return m
}

// applicationMapLabelsAndTags creates a map of LabelsAndTags providers keyed by application name.
func applicationMapLabelsAndTags(items []api.Config_LabelsAndTagsProvider) map[string]api.Config_LabelsAndTagsProvider {
	m := make(map[string]api.Config_LabelsAndTagsProvider, len(items))
	for _, p := range items {
		if p.Application == "" {
			continue
		}
		m[p.Application] = p
	}
	return m
}

func applicationSonarrRadarr(apps []api.Config_SonarrRadarrApp) map[string]api.Config_SonarrRadarrApp {
	m := make(map[string]api.Config_SonarrRadarrApp, len(apps))
	for _, a := range apps {
		if a.Library == "" {
			continue
		}
		m[a.Library] = a
	}
	return m
}

// IsMaskedWebhook checks if the given string matches the masked webhook pattern.
func IsMaskedWebhook(s string) bool {
	var reMasked3 = regexp.MustCompile(`\*{4}[^/]{3}/\*{3}[^/]{3}$`)
	return reMasked3.MatchString(strings.TrimSpace(s))
}
