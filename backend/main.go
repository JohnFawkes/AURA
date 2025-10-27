package main

import (
	"aura/internal/api"
	"aura/internal/logging"
	"aura/internal/routes"
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

	// // Print the banner with application details on startup
	api.Util_Banner_Print(APP_VERSION, Author, License, APP_PORT)

	// Set the umask if specified in environment
	SetUMask()

	// Load the config
	api.Config_LoadYamlConfig()

	// If config loaded, validate it
	if api.Global_Config_Loaded {
		api.Global_Config.ValidateConfig()
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
	if api.Global_Config_Loaded && api.Global_Config_Valid {
		api.Global_Config.PrintDetails()
		preFlightErr := finishPreflight()
		if preFlightErr.Message != "" {
			// Preflight failed: mark invalid and FALL BACK to onboarding routes
			api.Global_Config_Valid = false
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

	logging.LOG.Info("Fetching all items from the media server")
	api.GetAllSectionsAndItems()

	// Log the number of items in each section
	sections := api.Global_Cache_LibraryStore.GetAllSectionsSortedByTitle()
	for _, section := range sections {
		sectionTitle := section.Title
		logging.LOG.Info(fmt.Sprintf("Section '%s' contains %d items", fmt.Sprint(sectionTitle), len(section.MediaItems)))

		// Find the number of items without a TMDB ID
		countNoTMDB := 0
		for _, item := range section.MediaItems {
			if item.TMDB_ID == "" {
				countNoTMDB++
			}
		}
		if countNoTMDB > 0 {
			logging.LOG.Warn(fmt.Sprintf("Section '%s' contains %d items without a TMDB ID", fmt.Sprint(sectionTitle), countNoTMDB))
		}
	}

	if ok := api.DB_Init(); !ok {
		logging.LOG.Error("Database initialization failed; application not fully started.")
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
			logging.LOG.Error(fmt.Sprintf("Failed to schedule AutoDownload cron: %v", err))
		} else {
			logging.LOG.Info(fmt.Sprintf("AutoDownload set for: %s", api.Global_Config.AutoDownload.Cron))
			cronInstance.Start()
		}
	}

	// Send notification (only if not dev & notifications enabled)
	if !strings.Contains(APP_VERSION, "dev") &&
		api.Global_Config.Notifications.Enabled {
		if Err := api.SendAppStartNotification(); Err.Message != "" {
			logging.LOG.ErrorWithLog(Err)
		}
	}
}

// Perform token/userID preflight
func finishPreflight() logging.StandardError {
	Err := logging.NewStandardError()

	validateMediuxTokenErr := api.Mediux_ValidateToken(api.Global_Config.Mediux.Token)
	if validateMediuxTokenErr.Message != "" {
		api.Global_Config_MediuxValid = false
		logging.LOG.ErrorWithLog(validateMediuxTokenErr)
	}

	initUserIDErr := api.MediaServer_Init(api.Global_Config.MediaServer)
	if initUserIDErr.Message != "" {
		api.Global_Config_MediaServerValid = false
		logging.LOG.ErrorWithLog(initUserIDErr)
	}

	// If any errors in preflight, mark config as invalid
	if !api.Global_Config_MediuxValid || !api.Global_Config_MediaServerValid {
		logging.LOG.Warn("Configuration invalid after preflight checks")
		Err.Message = validateMediuxTokenErr.Message + " " + initUserIDErr.Message
	}

	return Err
}

func SetUMask() {
	if umaskStr := os.Getenv("UMASK"); umaskStr != "" {
		if umask, err := strconv.ParseInt(umaskStr, 8, 0); err == nil {
			syscall.Umask(int(umask))
			logging.LOG.Info(fmt.Sprintf("Set umask to %03o", umask))
		}
	}
}
