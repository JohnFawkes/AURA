package main

import (
	"aura/internal/api"
	"aura/internal/logging"
	"aura/internal/routes"
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"

	"github.com/robfig/cron/v3"
)

var (
	APP_VERSION = "dev"
	Author      = "xmoosex"
	License     = "MIT"
	APP_PORT    = 8888
)

// activeHandler dynamically switches from onboarding router to main app router.
var activeHandler atomic.Value

// cron instance started only after app is fully configured
var cronInstance *cron.Cron

func main() {

	// Print application details
	PrintDetails()

	// Set umask if specified
	SetUMask()

	// Load the config
	api.Config_LoadYamlConfig(context.Background())

	// Print the config details (sanitized)
	api.Global_Config.PrintDetails()

	// If config loaded, validate it
	if api.Global_Config_Loaded {
		api.Global_Config.ValidateConfig()
	}

	// Set the OnboardingComplete callback to swap routers
	routes.OnboardingComplete = func() {
		valid := finishPreflight()
		if !valid {
			return
		}
		startRuntimeServices()
		// Swap to the main router
		activeHandler.Store(routes.NewRouter())
		logging.LOGGER.Info().Timestamp().Msg("Onboarding complete. Main routes active.")
	}

	// Build initial router (either onboarding-only or full)
	initialRouter := routes.NewRouter()
	activeHandler.Store(initialRouter)

	// If the config is already loaded and valid, finish preflight
	if api.Global_Config_Loaded && api.Global_Config_Valid {
		valid := finishPreflight()
		if !valid {
			// Preflight failed, force onboarding
			api.Global_Config_Valid = false
			activeHandler.Store(routes.NewRouter())
		} else {
			startRuntimeServices()
			// Ensure the active handler is set to the main router
			activeHandler.Store(routes.NewRouter())
		}
	}

	logging.LOGGER.Info().Timestamp().Int("port", APP_PORT).Bool("full_routes", api.Global_Config_Loaded && api.Global_Config_Valid).Msg("Starting HTTP Server")
	if err := http.ListenAndServe(fmt.Sprintf(":%d", APP_PORT), http.HandlerFunc(dispatch)); err != nil {
		logging.LOGGER.Fatal().Err(err).Msg("Failed to start server")
	}
}

// Finish Preflight
// Validate MediaServer Connection (Get User ID for Emby/Jellyfin)
// Validate Mediux Token
func finishPreflight() bool {
	ctx, ld := logging.CreateLoggingContext(context.Background(), "Setting Up - Finish Preflight")
	defer ld.Log()

	action := ld.AddAction("Preflight Checks", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, action)
	defer action.Complete()

	// Validate Media Server Connection
	_, logErr := api.CallInitializeMediaServerConnection(ctx, api.Global_Config.MediaServer)
	if logErr.Message != "" {
		api.Global_Config_MediaServerValid = false
	}

	// Validate Mediux Token
	logErr = api.Mediux_ValidateToken(ctx, api.Global_Config.Mediux.Token)
	if logErr.Message != "" {
		api.Global_Config_MediuxValid = false
	}

	// Check if both Media Server and Mediux configs are valid
	if !api.Global_Config_MediuxValid || !api.Global_Config_MediaServerValid {
		return false
	}

	return true
}

// Start runtime services like:
// - Preload Sections and Items
// - Initialize Database
// - Start Cron Jobs
// - Send Startup Notification
func startRuntimeServices() {

	// Preload all sections and items
	api.GetAllSectionsAndItems()

	// Initialize the database
	if ok := api.DB_Init(); !ok {
		logging.LOGGER.Error().Timestamp().Msg("Database initialization failed, terminating application.")
		// Kill Application
		os.Exit(1)
	}

	// Schedule auto-download if enabled
	cronInstance = cron.New()
	if api.Global_Config_Loaded && api.Global_Config_Valid && api.Global_Config.AutoDownload.Enabled {
		_, err := cronInstance.AddFunc(api.Global_Config.AutoDownload.Cron, func() {
			if api.Global_Config.AutoDownload.Enabled {
				api.AutoDownload_CheckForUpdatesToPosters()
			}
		})
		if err != nil {
			logging.LOGGER.Error().Timestamp().Err(err).Msg("Failed to schedule AutoDownload cron job")
		} else {
			logging.LOGGER.Info().Timestamp().Msg(fmt.Sprintf("AutoDownload set for: %s", api.Global_Config.AutoDownload.Cron))
			cronInstance.Start()
		}
	} else {
		logging.LOGGER.Warn().Timestamp().Msg("AutoDownload is disabled")
	}

	// Send notification (only if not dev & notifications enabled)
	if !strings.Contains(APP_VERSION, "dev") &&
		api.Global_Config.Notifications.Enabled {
		api.SendAppStartNotification()
	}

}

func PrintDetails() {
	// Print application details
	logging.LOGGER.Info().
		Timestamp().
		Str("version", APP_VERSION).
		Str("author", Author).
		Str("license", License).
		Int("port", APP_PORT).
		Msg("Starting Aura")
}

func SetUMask() {
	_, ld := logging.CreateLoggingContext(context.Background(), "Setting Up - Set UMASK")
	logAction := ld.AddAction("Setting UMASK", logging.LevelInfo)
	defer func() {
		if logAction.Result != nil {
			ld.Log()
		}
	}()
	defer logAction.Complete()

	if umaskStr := os.Getenv("UMASK"); umaskStr != "" {
		if umask, err := strconv.ParseInt(umaskStr, 8, 0); err == nil {
			syscall.Umask(int(umask))
			logAction.AppendResult("umask_set", umaskStr)
		}
	}
}

// dispatch forwards to the currently active router.
func dispatch(w http.ResponseWriter, r *http.Request) {
	h := activeHandler.Load().(http.Handler)
	h.ServeHTTP(w, r)
}
