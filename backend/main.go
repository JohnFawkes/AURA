package main

import (
	"aura/internal/config"
	"aura/internal/database"
	"aura/internal/download"
	"aura/internal/logging"
	"aura/internal/notifications"
	"aura/internal/routes"
	mediaserver "aura/internal/server"
	mediaserver_shared "aura/internal/server/shared"
	"aura/internal/utils"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync/atomic"

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

	// Print the banner with application details on startup
	utils.PrintBanner(APP_VERSION, Author, License, APP_PORT)

	// Load (if exists) + validate
	config.LoadYamlConfig()
	if config.ConfigLoaded {
		config.ValidateConfig(config.Global)
	}

	// Set onboarding completion hook (called by /onboarding/apply success)
	routes.OnboardingComplete = func() {
		logging.LOG.Info("Onboarding Complete: Running Preflight")
		preFlightAfterOnboardErr := finishPreflight()
		if preFlightAfterOnboardErr.Message != "" {
			logging.LOG.Error("Preflight failed after onboarding: " + preFlightAfterOnboardErr.Message)
			return
		}
		startRuntimeServices()
		// Swap to full router (now config is valid)
		activeHandler.Store(routes.NewRouter())
		logging.LOG.Info("Onboarding complete. Main routes active.")
	}

	// Build initial router (either onboarding-only or full)
	initialRouter := routes.NewRouter()
	activeHandler.Store(initialRouter)

	// If already valid at startup, run preflight + runtime services
	if config.ConfigLoaded && config.ConfigValid {
		preFlightErr := finishPreflight()
		if preFlightErr.Message != "" {
			// Preflight failed: mark invalid and FALL BACK to onboarding routes
			config.ConfigValid = false
			logging.LOG.Error("Preflight failed on startup: " + preFlightErr.Message)
			// Swap router so only onboarding endpoints are exposed
			activeHandler.Store(routes.NewRouter())
		} else {
			startRuntimeServices()
			// Ensure full router (in case initial was onboarding)
			activeHandler.Store(routes.NewRouter())
		}
	}

	// Start HTTP server using dispatch (so we can atomically swap routers)
	logging.LOG.Info(fmt.Sprintf("Starting HTTP server on :%d", APP_PORT))
	if err := http.ListenAndServe(fmt.Sprintf(":%d", APP_PORT), http.HandlerFunc(dispatch)); err != nil {
		logging.LOG.Error(fmt.Sprintf("Error starting server: %s", err.Error()))
	}
}

// dispatch forwards to the currently active router.
func dispatch(w http.ResponseWriter, r *http.Request) {
	h := activeHandler.Load().(http.Handler)
	h.ServeHTTP(w, r)
}

// Called once config becomes valid (either at boot or after onboarding)
func startRuntimeServices() {

	if ok := database.InitDB(); !ok {
		logging.LOG.Error("Database initialization failed; application not fully started.")
		// Kill Application
		os.Exit(1)
	}

	// Schedule auto-download if enabled
	cronInstance = cron.New()
	if config.ConfigLoaded && config.ConfigValid && config.Global.AutoDownload.Enabled {
		_, err := cronInstance.AddFunc(config.Global.AutoDownload.Cron, func() {
			if config.Global.AutoDownload.Enabled {
				download.CheckForUpdatesToPosters()
			}
		})
		if err != nil {
			logging.LOG.Error(fmt.Sprintf("Failed to schedule AutoDownload cron: %v", err))
		} else {
			logging.LOG.Info(fmt.Sprintf("AutoDownload set for: %s", config.Global.AutoDownload.Cron))
			cronInstance.Start()
		}
	}

	// Send notification (only if not dev & notifications enabled)
	if !strings.Contains(APP_VERSION, "dev") &&
		config.Global.Notifications.Enabled {
		if Err := notifications.SendAppStartNotification(); Err.Message != "" {
			logging.LOG.ErrorWithLog(Err)
		}
	}

	go func() {
		logging.LOG.Info("Fetching all items from the media server")
		mediaserver.GetAllSectionsAndItems()
	}()
}

// Perform token/userID preflight
func finishPreflight() logging.StandardError {
	Err := logging.NewStandardError()

	validateMediuxTokenErr := utils.ValidateMediuxToken(config.Global.Mediux.Token)
	if validateMediuxTokenErr.Message != "" {
		config.ConfigMediuxValid = false
		logging.LOG.ErrorWithLog(validateMediuxTokenErr)
	}

	initUserIDErr := mediaserver_shared.InitUserID()
	if initUserIDErr.Message != "" {
		config.ConfigMediaServerValid = false
		logging.LOG.ErrorWithLog(initUserIDErr)
	}

	// If any errors in preflight, mark config as invalid
	if !config.ConfigMediuxValid || !config.ConfigMediaServerValid {
		logging.LOG.Warn("Configuration invalid after preflight checks")
		Err.Message = validateMediuxTokenErr.Message + " " + initUserIDErr.Message
	}

	return Err
}
