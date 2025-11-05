package routes_config

import (
	"aura/internal/api"
	"aura/internal/logging"
	"context"
	"fmt"
	"maps"
	"net/http"
	"reflect"
	"regexp"
	"sort"
	"strings"
)

func UpdateConfig(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Update Config", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	// Decode the incoming JSON request body into newConfig
	var newConfig api.Config
	Err := api.DecodeRequestBodyJSON(ctx, r.Body, &newConfig, "Config")
	if Err.Message != "" {
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}

	authChanged, authValid := checkConfigDifferences_Auth(ctx, api.Global_Config.Auth, &newConfig.Auth)
	loggingChanged, loggingValid := checkConfigDifferences_Logging(ctx, api.Global_Config.Logging, &newConfig.Logging)
	mediaServerChanged, mediaServerValid := checkConfigDifferences_MediaServer(ctx, api.Global_Config.MediaServer, &newConfig.MediaServer)
	mediuxChanged, mediuxValid := checkConfigDifferences_Mediux(ctx, api.Global_Config.Mediux, &newConfig.Mediux)
	autoDownloadChanged, autoDownloadValid := checkConfigDifferences_Autodownload(ctx, api.Global_Config.AutoDownload, &newConfig.AutoDownload)
	imagesChanged, imagesValid := checkConfigDifferences_Images(ctx, api.Global_Config.Images, &newConfig.Images, newConfig.MediaServer)
	tmdbChanged, tmdbValid := checkConfigDifferences_TMDB(ctx, api.Global_Config.TMDB, &newConfig.TMDB)
	labelsAndTagsChanged, labelsAndTagsValid := checkConfigDifferences_LabelsAndTags(ctx, api.Global_Config.LabelsAndTags, &newConfig.LabelsAndTags)
	notificationsChanged, notificationsValid := checkConfigDifferences_Notifications(ctx, api.Global_Config.Notifications, &newConfig.Notifications)
	sonarrRadarrChanged, sonarrRadarrValid := checkConfigDifferences_SonarrRadarr(ctx, api.Global_Config.SonarrRadarr, &newConfig.SonarrRadarr, newConfig.MediaServer)

	if !authValid {
		api.Util_Response_SendJSON(w, ld, "Invalid Auth configuration")
		return
	}

	if !loggingValid {
		api.Util_Response_SendJSON(w, ld, "Invalid Logging configuration")
		return
	}

	if !mediaServerValid {
		api.Util_Response_SendJSON(w, ld, "Invalid MediaServer configuration")
		return
	}

	if !mediuxValid {
		api.Util_Response_SendJSON(w, ld, "Invalid Mediux configuration")
		return
	}

	if !autoDownloadValid {
		api.Util_Response_SendJSON(w, ld, "Invalid AutoDownload configuration")
		return
	}

	if !imagesValid {
		api.Util_Response_SendJSON(w, ld, "Invalid Images configuration")
		return
	}

	if !tmdbValid {
		api.Util_Response_SendJSON(w, ld, "Invalid TMDB configuration")
		return
	}

	if !labelsAndTagsValid {
		api.Util_Response_SendJSON(w, ld, "Invalid LabelsAndTags configuration")
		return
	}

	if !notificationsValid {
		api.Util_Response_SendJSON(w, ld, "Invalid Notifications configuration")
		return
	}

	if !sonarrRadarrValid {
		api.Util_Response_SendJSON(w, ld, "Invalid SonarrRadarr configuration")
		return
	}

	if !authChanged && !loggingChanged && !mediaServerChanged && !mediuxChanged &&
		!autoDownloadChanged && !imagesChanged && !tmdbChanged && !labelsAndTagsChanged &&
		!notificationsChanged && !sonarrRadarrChanged {
		ld.Status = logging.StatusWarn
		logging.LOGGER.Warn().Timestamp().Msg("No changes detected in configuration")
		api.Util_Response_SendJSON(w, ld, "No changes detected in configuration")
		return
	}

	// Save the new configuration
	Err = newConfig.Update(ctx)
	if Err.Message != "" {
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}

	api.Util_Response_SendJSON(w, ld, newConfig.Sanitize(ctx))
}

// checkConfigDifferences_Auth compares old and new Auth configurations.
func checkConfigDifferences_Auth(ctx context.Context, oldAuth api.Config_Auth, newAuth *api.Config_Auth) (changed, newValid bool) {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Check Config Differences: Auth", logging.LevelTrace)
	defer logAction.Complete()
	changed = false
	newValid = false
	if !reflect.DeepEqual(oldAuth, newAuth) {
		if oldAuth.Enabled != newAuth.Enabled {
			logAction.AppendResult("Auth.Enabled changed", fmt.Sprintf("from '%v' to '%v'", oldAuth.Enabled, newAuth.Enabled))
			logging.LOGGER.Info().
				Timestamp().
				Bool("old_enabled", oldAuth.Enabled).
				Bool("new_enabled", newAuth.Enabled).
				Msg("Auth.Enabled changed")
			changed = true
		}

		if oldAuth.Password != newAuth.Password {
			logAction.AppendResult("Auth.Password changed", fmt.Sprintf("from '%s' to '%s'", oldAuth.Password, newAuth.Password))
			logging.LOGGER.Info().
				Timestamp().
				Str("old_password", fmt.Sprintf("%s", oldAuth.Password)).
				Str("new_password", fmt.Sprintf("%s", newAuth.Password)).
				Msg("Auth.Password changed")
			changed = true
		}
	}

	newValid = api.Config_ValidateAuth(ctx, newAuth)
	return changed, newValid
}

// checkConfigDifferences_Logging compares old and new Logging configurations.
func checkConfigDifferences_Logging(ctx context.Context, oldLogging api.Config_Logging, newLogging *api.Config_Logging) (changed, newValid bool) {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Check Config Differences: Logging", logging.LevelTrace)
	defer logAction.Complete()
	changed = false
	newValid = false
	if !reflect.DeepEqual(oldLogging, newLogging) {
		if oldLogging.Level != newLogging.Level {
			logAction.AppendResult("Logging.Level changed", fmt.Sprintf("from '%s' to '%s'", oldLogging.Level, newLogging.Level))
			logging.LOGGER.Info().
				Timestamp().
				Str("old_level", oldLogging.Level).
				Str("new_level", newLogging.Level).
				Msg("Logging.Level changed")
			changed = true
		}
	}

	newValid = api.Config_ValidateLogging(ctx, newLogging)
	return changed, newValid
}

// checkConfigDifferences_MediaServer compares old and new MediaServer configurations.
func checkConfigDifferences_MediaServer(ctx context.Context, oldMediaServer api.Config_MediaServer, newMediaServer *api.Config_MediaServer) (changed, newValid bool) {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Check Config Differences: MediaServer", logging.LevelTrace)
	defer logAction.Complete()
	changed = false
	newValid = false
	if !reflect.DeepEqual(oldMediaServer, newMediaServer) {
		if oldMediaServer.Type != newMediaServer.Type {
			logAction.AppendResult("MediaServer.Type changed", fmt.Sprintf("from '%s' to '%s'", oldMediaServer.Type, newMediaServer.Type))
			logging.LOGGER.Info().
				Timestamp().
				Str("old_type", oldMediaServer.Type).
				Str("new_type", newMediaServer.Type).
				Msg("MediaServer.Type changed")
			changed = true
		}

		if oldMediaServer.URL != newMediaServer.URL {
			logAction.AppendResult("MediaServer.URL changed", fmt.Sprintf("from '%s' to '%s'", oldMediaServer.URL, newMediaServer.URL))
			logging.LOGGER.Info().
				Timestamp().
				Str("old_url", oldMediaServer.URL).
				Str("new_url", newMediaServer.URL).
				Msg("MediaServer.URL changed")
			changed = true
		}

		if oldMediaServer.Token != newMediaServer.Token {
			if !strings.HasPrefix(newMediaServer.Token, "***") {
				logAction.AppendResult("MediaServer.Token changed", fmt.Sprintf("from '%s' to '%s'", oldMediaServer.Token, newMediaServer.Token))
				logging.LOGGER.Info().
					Timestamp().
					Str("old_token", fmt.Sprintf("%s", oldMediaServer.Token)).
					Str("new_token", fmt.Sprintf("%s", newMediaServer.Token)).
					Msg("MediaServer.Token changed")
				changed = true
			} else {
				newMediaServer.Token = oldMediaServer.Token
			}
		}

		if !reflect.DeepEqual(oldMediaServer.Libraries, newMediaServer.Libraries) {
			oldNames := libraryNames(oldMediaServer.Libraries)
			newNames := libraryNames(newMediaServer.Libraries)
			logAction.AppendResult("MediaServer.Libraries changed", fmt.Sprintf("from '%s' to '%s'", oldNames, newNames))
			logging.LOGGER.Info().
				Timestamp().
				Str("old_libraries", oldNames).
				Str("new_libraries", newNames).
				Msg("MediaServer.Libraries changed")
			changed = true
		}

		if oldMediaServer.UserID != newMediaServer.UserID {
			logAction.AppendResult("MediaServer.UserID changed", fmt.Sprintf("from '%v' to '%v'", oldMediaServer.UserID, newMediaServer.UserID))
			logging.LOGGER.Info().
				Timestamp().
				Str("old_user_id", fmt.Sprintf("%v", oldMediaServer.UserID)).
				Str("new_user_id", fmt.Sprintf("%v", newMediaServer.UserID)).
				Msg("MediaServer.UserID changed")
			changed = true
		}
	}
	newValid = api.Config_ValidateMediaServer(ctx, newMediaServer)
	return changed, newValid
}

// checkConfigDifferences_Mediux compares old and new Mediux configurations.
func checkConfigDifferences_Mediux(ctx context.Context, oldMediux api.Config_Mediux, newMediux *api.Config_Mediux) (changed, newValid bool) {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Check Config Differences: Mediux", logging.LevelTrace)
	defer logAction.Complete()
	changed = false
	newValid = false
	if !reflect.DeepEqual(oldMediux, newMediux) {
		if oldMediux.Token != newMediux.Token {
			if !strings.HasPrefix(newMediux.Token, "***") {
				logAction.AppendResult("Mediux.Token changed", fmt.Sprintf("from '%s' to '%s'", oldMediux.Token, newMediux.Token))
				logging.LOGGER.Info().
					Timestamp().
					Str("old_token", fmt.Sprintf("%v", oldMediux.Token)).
					Str("new_token", fmt.Sprintf("%v", newMediux.Token)).
					Msg("Mediux.Token changed")
				changed = true
			} else {
				newMediux.Token = oldMediux.Token
			}
		}

		if oldMediux.DownloadQuality != newMediux.DownloadQuality {
			logAction.AppendResult("Mediux.DownloadQuality changed", fmt.Sprintf("from '%v' to '%v'", oldMediux.DownloadQuality, newMediux.DownloadQuality))
			logging.LOGGER.Info().
				Timestamp().
				Str("old_download_quality", fmt.Sprintf("%v", oldMediux.DownloadQuality)).
				Str("new_download_quality", fmt.Sprintf("%v", newMediux.DownloadQuality)).
				Msg("Mediux.DownloadQuality changed")
			changed = true
		}
	}
	newValid = api.Config_ValidateMediux(ctx, newMediux)
	return changed, newValid
}

// checkConfigDifferences_Autodownload compares old and new AutoDownload configurations.
func checkConfigDifferences_Autodownload(ctx context.Context, oldAutoDownload api.Config_AutoDownload, newAutoDownload *api.Config_AutoDownload) (changed, newValid bool) {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Check Config Differences: Autodownload", logging.LevelTrace)
	defer logAction.Complete()
	changed = false
	newValid = false
	if !reflect.DeepEqual(oldAutoDownload, newAutoDownload) {
		if oldAutoDownload.Enabled != newAutoDownload.Enabled {
			logAction.AppendResult("Autodownload.Enabled changed", fmt.Sprintf("from '%v' to '%v'", oldAutoDownload.Enabled, newAutoDownload.Enabled))
			logging.LOGGER.Info().
				Timestamp().
				Bool("old_enabled", oldAutoDownload.Enabled).
				Bool("new_enabled", newAutoDownload.Enabled).
				Msg("Autodownload.Enabled changed")
			changed = true
		}

		if oldAutoDownload.Cron != newAutoDownload.Cron {
			logAction.AppendResult("Autodownload.Cron changed", fmt.Sprintf("from '%s' to '%s'", oldAutoDownload.Cron, newAutoDownload.Cron))
			logging.LOGGER.Info().
				Timestamp().
				Str("old_cron", oldAutoDownload.Cron).
				Str("new_cron", newAutoDownload.Cron).
				Msg("Autodownload.Cron changed")
			changed = true
		}
	}
	newValid = api.Config_ValidateAutoDownload(ctx, newAutoDownload)
	return changed, newValid
}

// checkConfigDifferences_Images compares old and new Images configurations.
func checkConfigDifferences_Images(ctx context.Context, oldImages api.Config_Images, newImages *api.Config_Images, msConfig api.Config_MediaServer) (changed, newValid bool) {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Check Config Differences: Images", logging.LevelTrace)
	defer logAction.Complete()
	changed = false
	newValid = true
	if !reflect.DeepEqual(oldImages, newImages) {
		if oldImages.CacheImages.Enabled != newImages.CacheImages.Enabled {
			logAction.AppendResult("Images.CacheImages.Enabled changed", fmt.Sprintf("from '%v' to '%v'", oldImages.CacheImages.Enabled, newImages.CacheImages.Enabled))
			logging.LOGGER.Info().
				Timestamp().
				Bool("old_enabled", oldImages.CacheImages.Enabled).
				Bool("new_enabled", newImages.CacheImages.Enabled).
				Msg("Images.CacheImages.Enabled changed")
			changed = true
		}

		if oldImages.SaveImagesLocally.Enabled != newImages.SaveImagesLocally.Enabled {
			logAction.AppendResult("Images.SaveImagesLocally.Enabled changed", fmt.Sprintf("from '%v' to '%v'", oldImages.SaveImagesLocally.Enabled, newImages.SaveImagesLocally.Enabled))
			logging.LOGGER.Info().
				Timestamp().
				Bool("old_enabled", oldImages.SaveImagesLocally.Enabled).
				Bool("new_enabled", newImages.SaveImagesLocally.Enabled).
				Msg("Images.SaveImagesLocally.Enabled changed")
			changed = true
		}

		if oldImages.SaveImagesLocally.Path != newImages.SaveImagesLocally.Path {
			logAction.AppendResult("Images.SaveImagesLocally.Path changed", fmt.Sprintf("from '%s' to '%s'", oldImages.SaveImagesLocally.Path, newImages.SaveImagesLocally.Path))
			logging.LOGGER.Info().
				Timestamp().
				Str("old_path", oldImages.SaveImagesLocally.Path).
				Str("new_path", newImages.SaveImagesLocally.Path).
				Msg("Images.SaveImagesLocally.Path changed")
			changed = true
		}

		if oldImages.SaveImagesLocally.SeasonNamingConvention != newImages.SaveImagesLocally.SeasonNamingConvention {
			logAction.AppendResult("Images.SaveImagesLocally.SeasonNamingConvention changed", fmt.Sprintf("from '%s' to '%s'", oldImages.SaveImagesLocally.SeasonNamingConvention, newImages.SaveImagesLocally.SeasonNamingConvention))
			logging.LOGGER.Info().
				Timestamp().
				Str("old_season_naming_convention", oldImages.SaveImagesLocally.SeasonNamingConvention).
				Str("new_season_naming_convention", newImages.SaveImagesLocally.SeasonNamingConvention).
				Msg("Images.SaveImagesLocally.SeasonNamingConvention changed")
			changed = true
		}

		if oldImages.SaveImagesLocally.EpisodeNamingConvention != newImages.SaveImagesLocally.EpisodeNamingConvention {
			logAction.AppendResult("Images.SaveImagesLocally.EpisodeNamingConvention changed", fmt.Sprintf("from '%s' to '%s'", oldImages.SaveImagesLocally.EpisodeNamingConvention, newImages.SaveImagesLocally.EpisodeNamingConvention))
			logging.LOGGER.Info().
				Timestamp().
				Str("old_episode_naming_convention", oldImages.SaveImagesLocally.EpisodeNamingConvention).
				Str("new_episode_naming_convention", newImages.SaveImagesLocally.EpisodeNamingConvention).
				Msg("Images.SaveImagesLocally.EpisodeNamingConvention changed")
			changed = true
		}
	}
	newValid = api.Config_ValidateImages(ctx, newImages, msConfig)
	return changed, newValid
}

// checkConfigDifferences_TMDB compares old and new TMDB configurations.
func checkConfigDifferences_TMDB(ctx context.Context, oldTMDB api.Config_TMDB, newTMDB *api.Config_TMDB) (changed, newValid bool) {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Check Config Differences: TMDB", logging.LevelTrace)
	defer logAction.Complete()
	changed = false
	newValid = true
	if !reflect.DeepEqual(oldTMDB, newTMDB) {
		if oldTMDB.APIKey != newTMDB.APIKey {
			if !strings.HasPrefix(newTMDB.APIKey, "***") {
				logAction.AppendResult("TMDB.APIKey changed", fmt.Sprintf("from '%s' to '%s'", oldTMDB.APIKey, newTMDB.APIKey))
				logging.LOGGER.Info().
					Timestamp().
					Str("old_api_key", fmt.Sprintf("%v", oldTMDB.APIKey)).
					Str("new_api_key", fmt.Sprintf("%v", newTMDB.APIKey)).
					Msg("TMDB.APIKey changed")
				changed = true
			} else {
				newTMDB.APIKey = oldTMDB.APIKey
			}
		}
	}
	return changed, newValid
}

// checkConfigDifferences_LabelsAndTags compares old and new LabelsAndTags configurations.
func checkConfigDifferences_LabelsAndTags(ctx context.Context, oldLAT api.Config_LabelsAndTags, newLAT *api.Config_LabelsAndTags) (changed, newValid bool) {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Check Config Differences: LabelsAndTags", logging.LevelTrace)
	defer logAction.Complete()
	changed = false
	newValid = true
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
			logAction.AppendResult("LabelsAndTags.Applications added", fmt.Sprintf("%s", joinNonEmptyComma(added)))
			logging.LOGGER.Info().
				Timestamp().
				Str("added_apps", fmt.Sprintf("%s", joinNonEmptyComma(added))).
				Msg("LabelsAndTags.Applications added")
			changed = true
		}
		if len(removed) > 0 {
			sort.Strings(removed)
			logAction.AppendResult("LabelsAndTags.Applications removed", fmt.Sprintf("%s", joinNonEmptyComma(removed)))
			logging.LOGGER.Info().
				Timestamp().
				Str("removed_apps", fmt.Sprintf("%s", joinNonEmptyComma(removed))).
				Msg("LabelsAndTags.Applications removed")
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
				logAction.AppendResult("LabelsAndTags.Application.Enabled changed", fmt.Sprintf("from '%v' to '%v'", oldApp.Enabled, newApp.Enabled))
				logging.LOGGER.Info().
					Timestamp().
					Str("application", name).
					Bool("old_enabled", oldApp.Enabled).
					Bool("new_enabled", newApp.Enabled).
					Msg("LabelsAndTags.Application.Enabled changed")
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
				logAction.AppendResult("LabelsAndTags.Application.Add added", fmt.Sprintf("%s", joinNonEmptyComma(addAdded)))
				logging.LOGGER.Info().
					Timestamp().
					Str("application", name).
					Str("added_labels", fmt.Sprintf("%s", joinNonEmptyComma(addAdded))).
					Msg("LabelsAndTags.Application.Add added")
				changed = true
			}
			if len(addRemoved) > 0 {
				sort.Strings(addRemoved)
				logAction.AppendResult("LabelsAndTags.Application.Add removed", fmt.Sprintf("%s", joinNonEmptyComma(addRemoved)))
				logging.LOGGER.Info().
					Timestamp().
					Str("application", name).
					Str("removed_labels", fmt.Sprintf("%s", joinNonEmptyComma(addRemoved))).
					Msg("LabelsAndTags.Application.Add removed")
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
				logAction.AppendResult("LabelsAndTags.Application.Remove added", fmt.Sprintf("%s", joinNonEmptyComma(removeAdded)))
				logging.LOGGER.Info().
					Timestamp().
					Str("application", name).
					Str("added_labels", fmt.Sprintf("%s", joinNonEmptyComma(removeAdded))).
					Msg("LabelsAndTags.Application.Remove added")
				changed = true
			}
			if len(removeRemoved) > 0 {
				sort.Strings(removeRemoved)
				logAction.AppendResult("LabelsAndTags.Application.Remove removed", fmt.Sprintf("%s", joinNonEmptyComma(removeRemoved)))
				logging.LOGGER.Info().
					Timestamp().
					Str("application", name).
					Str("removed_labels", fmt.Sprintf("%s", joinNonEmptyComma(removeRemoved))).
					Msg("LabelsAndTags.Application.Remove removed")
				changed = true
			}
		}
	}
	return changed, newValid
}

// checkConfigDifferences_Notifications compares old and new Notifications configurations.
func checkConfigDifferences_Notifications(ctx context.Context, oldNotifications api.Config_Notifications, newNotifications *api.Config_Notifications) (changed, newValid bool) {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Check Config Differences: Notifications", logging.LevelTrace)
	defer logAction.Complete()
	changed = false
	newValid = false
	if !reflect.DeepEqual(oldNotifications, newNotifications) {
		// Global toggle
		if oldNotifications.Enabled != newNotifications.Enabled {
			logAction.AppendResult("Notifications.Enabled changed", fmt.Sprintf("from '%v' to '%v'", oldNotifications.Enabled, newNotifications.Enabled))
			logging.LOGGER.Info().
				Timestamp().
				Bool("old_enabled", oldNotifications.Enabled).
				Bool("new_enabled", newNotifications.Enabled).
				Msg("Notifications.Enabled changed")
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
			logAction.AppendResult("Notifications.Providers added", fmt.Sprintf("%s", joinNonEmptyComma(added)))
			logging.LOGGER.Info().
				Timestamp().
				Str("added_providers", fmt.Sprintf("%s", joinNonEmptyComma(added))).
				Msg("Notifications.Providers added")
			changed = true
		}
		if len(removed) > 0 {
			sort.Strings(removed)
			logAction.AppendResult("Notifications.Providers removed", fmt.Sprintf("%s", joinNonEmptyComma(removed)))
			logging.LOGGER.Info().
				Timestamp().
				Str("removed_providers", fmt.Sprintf("%s", joinNonEmptyComma(removed))).
				Msg("Notifications.Providers removed")
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
				logAction.AppendResult("Notifications.Provider.Enabled changed", fmt.Sprintf("from '%v' to '%v'", oldProv.Enabled, newProv.Enabled))
				logging.LOGGER.Info().
					Timestamp().
					Str("provider", name).
					Bool("old_enabled", oldProv.Enabled).
					Bool("new_enabled", newProv.Enabled).
					Msg("Notifications.Provider.Enabled changed")
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
						logAction.AppendResult("Notifications.Discord.Webhook changed", fmt.Sprintf("from '%v' to '%v'", oldWebhook, newWebhook))
						logging.LOGGER.Info().
							Timestamp().
							Str("old_webhook", oldWebhook).
							Str("new_webhook", newWebhook).
							Msg("Notifications.Discord.Webhook changed")
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
						logAction.AppendResult("Notifications.Pushover.UserKey changed", fmt.Sprintf("from '%v' to '%v'", oldUserKey, newUserKey))
						logging.LOGGER.Info().
							Timestamp().
							Str("old_user_key", oldUserKey).
							Str("new_user_key", newUserKey).
							Msg("Notifications.Pushover.UserKey changed")
						changed = true
					} else {
						newProv.Pushover.UserKey = oldProv.Pushover.UserKey
					}

				}
				if oldToken != newToken {
					if !strings.HasPrefix(newToken, "***") {
						logAction.AppendResult("Notifications.Pushover.Token changed", fmt.Sprintf("from '%v' to '%v'", oldToken, newToken))
						logging.LOGGER.Info().
							Timestamp().
							Str("old_token", oldToken).
							Str("new_token", newToken).
							Msg("Notifications.Pushover.Token changed")
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
					logAction.AppendResult("Notifications.Gotify.URL changed", fmt.Sprintf("from '%v' to '%v'", oldGotifyURL, newGotifyURL))
					logging.LOGGER.Info().
						Timestamp().
						Str("old_gotify_url", oldGotifyURL).
						Str("new_gotify_url", newGotifyURL).
						Msg("Notifications.Gotify.URL changed")
					changed = true
				}
				// Token may still arrive masked; keep existing mask logic
				if oldProv.Gotify != nil && newProv.Gotify != nil {
					if oldProv.Gotify.Token != newProv.Gotify.Token {
						if !IsMaskedWebhook(newProv.Gotify.Token) {
							logAction.AppendResult("Notifications.Gotify.Token changed", fmt.Sprintf("from '%v' to '%v'", oldProv.Gotify.Token, newProv.Gotify.Token))
							logging.LOGGER.Info().
								Timestamp().
								Str("old_token", fmt.Sprintf("%s", oldProv.Gotify.Token)).
								Str("new_token", fmt.Sprintf("%s", newProv.Gotify.Token)).
								Msg("Notifications.Gotify.Token changed")
							changed = true
						} else {
							newProv.Gotify.Token = oldProv.Gotify.Token
						}
					}
				}

			case "Webhook":
				var oldURL, newURL string
				if oldProv.Webhook != nil {
					oldURL = strings.TrimSpace(oldProv.Webhook.URL)
				}
				if newProv.Webhook != nil {
					newURL = strings.TrimSpace(newProv.Webhook.URL)
				}
				if oldURL != newURL {
					logAction.AppendResult("Notifications.Webhook.URL changed", fmt.Sprintf("from '%v' to '%v'", oldURL, newURL))
					logging.LOGGER.Info().
						Timestamp().
						Str("old_webhook_url", oldURL).
						Str("new_webhook_url", newURL).
						Msg("Notifications.Webhook.URL changed")
					changed = true
				}

				// Custom Headers
				oldHeaders := make(map[string]string)
				newHeaders := make(map[string]string)
				if oldProv.Webhook != nil {
					maps.Copy(oldHeaders, oldProv.Webhook.Headers)
				}
				if newProv.Webhook != nil {
					maps.Copy(newHeaders, newProv.Webhook.Headers)
				}
				// Check for changes
				for k, oldV := range oldHeaders {
					newV, ok := newHeaders[k]
					if !ok || oldV != newV {
						logAction.AppendResult("Notifications.Webhook.Header changed", fmt.Sprintf("Header '%s' changed from '%s' to '%s'", k, oldV, newV))
						logging.LOGGER.Info().
							Timestamp().
							Str("header_key", k).
							Str("old_value", oldV).
							Str("new_value", newV).
							Msg("Notifications.Webhook.Header changed")
						changed = true
					}
				}
				for k, newV := range newHeaders {
					if _, ok := oldHeaders[k]; !ok {
						logAction.AppendResult("Notifications.Webhook.Header added", fmt.Sprintf("Header '%s' added with value '%s'", k, newV))
						logging.LOGGER.Info().
							Timestamp().
							Str("header_key", k).
							Str("new_value", newV).
							Msg("Notifications.Webhook.Header added")
						changed = true
					}
				}
			default:
				// Unknown provider type: nothing more to compare
			}
		}
	}
	newValid = api.Config_ValidateNotifications(ctx, newNotifications)
	return changed, newValid
}

// checkConfigDifferences_SonarrRadarr compares old and new Sonarr/Radarr configurations.
func checkConfigDifferences_SonarrRadarr(ctx context.Context, oldSR api.Config_SonarrRadarr_Apps, newSR *api.Config_SonarrRadarr_Apps, msConfig api.Config_MediaServer) (changed, newValid bool) {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Check Config Differences: SonarrRadarr", logging.LevelTrace)
	defer logAction.Complete()
	changed = false
	newValid = false
	if !reflect.DeepEqual(oldSR, newSR) {

		// Providers diff
		oldMap := applicationSonarrRadarr(oldSR.Applications)
		newMap := applicationSonarrRadarr(newSR.Applications)

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
			logAction.AppendResult("SonarrRadarr.Applications added", fmt.Sprintf("%s", joinNonEmptyComma(added)))
			logging.LOGGER.Info().
				Timestamp().
				Str("added_applications", fmt.Sprintf("%s", joinNonEmptyComma(added))).
				Msg("SonarrRadarr.Applications added")
			changed = true
		}
		if len(removed) > 0 {
			sort.Strings(removed)
			logAction.AppendResult("SonarrRadarr.Applications removed", fmt.Sprintf("%s", joinNonEmptyComma(removed)))
			logging.LOGGER.Info().
				Timestamp().
				Str("removed_applications", fmt.Sprintf("%s", joinNonEmptyComma(removed))).
				Msg("SonarrRadarr.Applications removed")
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
					logAction.AppendResult("SonarrRadarr.Application.APIKey changed", fmt.Sprintf("from '%s' to '%s'", oldProv.APIKey, newProv.APIKey))
					logging.LOGGER.Info().
						Timestamp().
						Str("application", name).
						Str("old_api_key", oldProv.APIKey).
						Str("new_api_key", newProv.APIKey).
						Msg("SonarrRadarr.Application APIKey changed")
					changed = true
				} else {

					newProv.APIKey = oldProv.APIKey
				}
			}

			// Per App - URL
			if oldProv.URL != newProv.URL {
				logAction.AppendResult("SonarrRadarr.Application.URL changed", fmt.Sprintf("from '%s' to '%s'", oldProv.URL, newProv.URL))
				logging.LOGGER.Info().
					Timestamp().
					Str("application", name).
					Str("old_url", oldProv.URL).
					Str("new_url", newProv.URL).
					Msg("SonarrRadarr.Application URL changed")
				changed = true
			}

			// Per App - Type
			if oldProv.Type != newProv.Type {
				logAction.AppendResult("SonarrRadarr.Application.Type changed", fmt.Sprintf("from '%s' to '%s'", oldProv.Type, newProv.Type))
				logging.LOGGER.Info().
					Timestamp().
					Str("application", name).
					Str("old_type", oldProv.Type).
					Str("new_type", newProv.Type).
					Msg("SonarrRadarr.Application Type changed")
				changed = true
			}

			// Per App - Library
			if oldProv.Library != newProv.Library {
				logAction.AppendResult("SonarrRadarr.Application.Library changed", fmt.Sprintf("from '%s' to '%s'", oldProv.Library, newProv.Library))
				logging.LOGGER.Info().
					Timestamp().
					Str("application", name).
					Str("old_library", oldProv.Library).
					Str("new_library", newProv.Library).
					Msg("SonarrRadarr.Application Library changed")
				changed = true
			}

		}
	}

	// Restore API keys
	for i, app := range newSR.Applications {
		if strings.HasPrefix(app.APIKey, "***") {
			// Find matching old app by Library and Type
			for _, oldApp := range oldSR.Applications {
				if app.Library == oldApp.Library && app.Type == oldApp.Type {
					newSR.Applications[i].APIKey = oldApp.APIKey
					break
				}
			}
		}
	}

	newValid = api.Config_ValidateSonarrRadarr(ctx, newSR, msConfig)
	return changed, newValid
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
