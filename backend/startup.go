package main

import (
	"aura/cache"
	"aura/config"
	"aura/database"
	"aura/database/migration"
	downloadqueue "aura/download/queue"
	"aura/jobs"
	"aura/logging"
	"aura/mediaserver"
	"aura/mediux"
	"aura/notification"
	"aura/utils"
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

func runBootstrap() (success bool) {
	ctx, ld := logging.CreateLoggingContext(context.Background(), "Bootstrap")
	defer ld.Log()

	logAction := ld.AddAction("Application Startup", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)
	defer logAction.Complete()

	success = false

	// Print App Info
	utils.PrintAppStartUpDetails(APP_VERSION, AUTHOR, LICENSE, APP_PORT, APP_NAME)
	config.Version = APP_VERSION

	// Set Umask for file permissions (if needed)
	utils.SetUMask(ctx)

	// Load the config file
	config.LoadYAML(ctx)
	logAction.Complete()

	// Print the config details (sanitized)
	config.Current.PrintDetails()

	// If the config is loaded, validate it
	if config.Loaded {
		config.Current.Validate(ctx)
	}

	if config.Loaded && config.Valid {
		success = true
	}

	return success
}

func runPreFlight() (success bool) {
	ctx, ld := logging.CreateLoggingContext(context.Background(), "Preflight")
	defer ld.Log()

	action := ld.AddAction("Checking Services", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, action)
	defer action.Complete()

	success = false
	config.AppFullyLoaded = false

	// Validate Media Server Connection
	connectionOk, serverName, serverVersion, msErr := mediaserver.TestConnection(ctx, &config.Current.MediaServer)
	if msErr.Message != "" || !connectionOk || serverVersion == "" || serverName == "" {
		config.MediaServerValid = false
		return success
	}
	if config.Current.MediaServer.Type == "Jellyfin" || config.Current.MediaServer.Type == "Emby" {
		// Get Admin User for Emby/Jellyfin
		ejUserID, initErr := mediaserver.GetAdminUser(ctx, &config.Current.MediaServer)
		if initErr.Message != "" {
			config.MediaServerValid = false
			return success
		} else if ejUserID == "" {
			config.MediaServerValid = false
			logging.LOGGER.Error().Timestamp().Msg("Failed to retrieve admin user ID from Emby/Jellyfin server")
			return success
		}
		config.Current.MediaServer.UserID = ejUserID
	}
	config.MediaServerName = serverName
	logging.LOGGER.Trace().Timestamp().Str("media_server_name", serverName).
		Str("media_server_version", serverVersion).
		Msg("Media Server connection validated successfully")
	config.MediaServerValid = true

	// Validate MediUX Token
	mediuxTokenValid, mediuxErr := mediux.ValidateToken(ctx, config.Current.Mediux.ApiToken)
	if mediuxErr.Message != "" || !mediuxTokenValid {
		config.MediuxValid = false
		return success
	}

	if config.MediaServerValid || config.MediuxValid {
		success = true
	}

	return success
}

func runWarmup() (success bool) {
	ctx, ld := logging.CreateLoggingContext(context.Background(), "Warmup")

	action := ld.AddAction("Initializing Application", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, action)

	success = false

	// Cache: Add all MediUX users
	mediux.PreloadMediuxUsers(ctx)
	logging.LOGGER.Info().Timestamp().Int("mediux_users", len(cache.MediuxUsers.GetMediuxUsers())).Msg("Preloaded MediUX users into cache")

	// Cache: Get a list of all items in MediUX that has a set
	mediux.PreLoadMediuxItemsWithSets(ctx)
	logging.LOGGER.Info().Timestamp().Int("mediux_items_with_sets", len(cache.MediuxItems.GetMediuxItems())).Msg("Preloaded MediUX items with sets into cache")

	// Database: Initialize
	newDB, dbInitErr := database.Init(ctx)
	if dbInitErr.Message != "" {
		return false
	}
	logging.LOGGER.Info().Timestamp().Bool("new_database", newDB).Msg("Database initialized")

	// Database-Migration: If not a new DB, run migrations
	if !newDB {
		migrationsCompleted, _ := migration.RunMigrations()
		logging.LOGGER.Info().Timestamp().Msgf("%d database migrations performed", migrationsCompleted)
	}

	// Cache: Add all media server sections and items
	_ = mediaserver.GetAllLibrarySectionsAndItems(ctx, false)
	logging.LOGGER.Info().Timestamp().Int("sections", cache.LibraryStore.GetSectionsCount()).Int("items", cache.LibraryStore.GetItemsCount()).Msg("Preloaded Media Server sections and items into cache")
	logging.LOGGER.Info().Timestamp().Int("collection_items", len(cache.CollectionsStore.GetAllCollections())).
		Msg("Preloaded Media Server data into cache")

	// Database: Vacuum
	vacuumErr := database.Vacuum(ctx)
	if vacuumErr.Message != "" {
		logging.LOGGER.Error().Timestamp().Msgf("Database VACUUM failed: %s", vacuumErr.Message)
		return false
	}

	action.Complete()
	ld.Log()

	// Cronjob: Auto Download Processing
	jobs.StartAutoDownloadJob()

	// Cronjob: Download Queue Processing
	err := jobs.StartDownloadQueueJob()
	if err != nil {
		logging.LOGGER.Error().Timestamp().Err(err).Msg("Failed to schedule Download Queue Processing cron job")
		downloadqueue.LatestInfo.Time = time.Now()
		downloadqueue.LatestInfo.Status = downloadqueue.LAST_STATUS_ERROR
		downloadqueue.LatestInfo.Message = "Failed to schedule Download Queue Processing"
		downloadqueue.LatestInfo.Errors = []string{err.Error()}
		downloadqueue.LatestInfo.Warnings = []string{}
	} else {
		downloadqueue.LatestInfo.Time = time.Now()
		downloadqueue.LatestInfo.Status = downloadqueue.LAST_STATUS_IDLE
	}

	// Cronjob: Refresh Media Items and Collections
	err = jobs.StartRefreshMediaItemsAndCollectionsJob()
	if err != nil {
		logging.LOGGER.Error().Timestamp().Err(err).Msg("Failed to schedule Refresh Media Items and Collections cron job")
	}

	// Cronjob: Refresh Mediux Users
	err = jobs.StartRefreshMediuxUsersJob()
	if err != nil {
		logging.LOGGER.Error().Timestamp().Err(err).Msg("Failed to schedule Refresh Mediux Users cron job")
	}

	// Cronjob: Check MediUX Site Link Availability
	mediux.CheckSiteLinkAvailability()
	err = jobs.StartCheckMediuxSiteLinkJob()
	if err != nil {
		logging.LOGGER.Error().Timestamp().Err(err).Msg("Failed to schedule Check MediUX Site Link Availability cron job")
	}

	// Cronjob: Start Check for Rating Key Changes Job
	err = jobs.StartCheckForRatingKeyChangesJob()
	if err != nil {
		logging.LOGGER.Error().Timestamp().Err(err).Msg("Failed to schedule Check for Rating Key Changes cron job")
	}

	// Cron: Start Jobs Scheduler
	jobs.StartJobs()

	// Initialize MediUX WebSocket Listener
	go mediux.StartWebSocketClient()

	success = true
	return success
}

func startAPI() {
	// Send App Start Notification
	// Send notification (only if not dev & notifications enabled)
	if !strings.Contains(APP_VERSION, "dev") &&
		config.Current.Notifications.Enabled {
		notification.SendAppStartNotification(APP_PORT, APP_NAME, APP_VERSION)
	} else {
		logging.LOGGER.Warn().Timestamp().Bool("notifications_enabled", config.Current.Notifications.Enabled).Bool("dev_version", strings.Contains(APP_VERSION, "dev")).Msg("App start notification not sent")
	}

	jobs.RunAutoDownloadJobNow()
	jobs.RunCheckForRatingKeyChangesJobNow()

	// Start HTTP Server
	logging.LOGGER.Info().Timestamp().Int("port", APP_PORT).
		Bool("full_routes", config.Loaded && config.Valid).
		Str("log_level", logging.LOGGER.GetLevel().String()).
		Msg("Starting HTTP Server")
	if err := http.ListenAndServe(fmt.Sprintf(":%d", APP_PORT), http.HandlerFunc(dispatch)); err != nil {
		logging.LOGGER.Fatal().Err(err).Msg("Failed to start server")
	}
}

// dispatch forwards to the currently active router.
func dispatch(w http.ResponseWriter, r *http.Request) {
	v := activeHandler.Load()
	h, ok := v.(http.Handler)
	if !ok || h == nil {
		// router not initialized yet (or stored value is wrong type)
		logging.LOGGER.Error().
			Timestamp().
			Str("path", r.URL.Path).
			Msg("activeHandler not initialized")
		http.Error(w, "Service starting up; router not initialized", http.StatusServiceUnavailable)
		return
	}

	h.ServeHTTP(w, r)
}
