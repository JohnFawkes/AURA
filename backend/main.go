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

	Err := logging.StandardError{}

	// Load (if exists) + validate
	config.LoadYamlConfig()
	if config.ConfigLoaded {
		config.ValidateConfig(config.Global)
	}

	// Set onboarding completion hook (called by /onboarding/apply success)
	routes.OnboardingComplete = func() {
		logging.LOG.Info("Onboarding Complete: Running Preflight")
		e := logging.StandardError{}
		if !finishPreflight(&e) {
			logging.LOG.Error("Preflight failed after onboarding: " + e.Message)
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
		if finishPreflight(&Err) {
			startRuntimeServices()
		} else {
			logging.LOG.Error("Preflight failed on startup: " + Err.Message)
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
	// Init DB
	if ok := database.InitDB(); !ok {
		logging.LOG.Error("Database initialization failed; application not fully started.")
		return
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
func finishPreflight(Err *logging.StandardError) bool {
	*Err = utils.ValidateMediuxToken(config.Global.Mediux.Token)
	if Err.Message != "" {
		logging.LOG.ErrorWithLog(*Err)
		return false
	}
	*Err = mediaserver_shared.InitUserID()
	if Err.Message != "" {
		logging.LOG.ErrorWithLog(*Err)
		return false
	}
	return true
}
