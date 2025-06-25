package main

import (
	"aura/internal/config"
	"aura/internal/database"
	"aura/internal/download"
	"aura/internal/logging"
	"aura/internal/notifications"
	"aura/internal/routes"
	mediaserver_shared "aura/internal/server/shared"
	"aura/internal/utils"
	"fmt"
	"net/http"
	"strings"

	"github.com/robfig/cron/v3"
)

var (
	APP_VERSION = "dev"
	Author      = "xmoosex"
	License     = "MIT"
	APP_PORT    = 8888
)

func main() {

	// Print the banner with application details on startup
	utils.PrintBanner(
		APP_VERSION,
		Author,
		License,
		APP_PORT,
	)

	// Load the configuration file
	// If the config file is not found, exit the program
	Err := config.LoadYamlConfig()
	if Err.Message != "" {
		logging.LOG.ErrorWithLog(Err)
		return
	}

	// Check if the config file is valid
	valid := config.ValidateConfig()
	if !valid {
		return
	}

	// Print the configuration settings
	config.PrintConfig()

	// Validate Mediux Token
	Err = utils.ValidateMediUXToken(config.Global.Mediux.Token)
	if Err.Message != "" {
		logging.LOG.ErrorWithLog(Err)
		return
	}

	// Initialize the database
	init := database.InitDB()
	if !init {
		return
	}

	// For Jellyfin/Emby, we need to get the User ID for the Admin user
	Err = mediaserver_shared.InitUserID()
	if Err.Message != "" {
		logging.LOG.ErrorWithLog(Err)
		return
	}

	// Create a new router
	r := routes.NewRouter()

	// Create a new cron instance
	c := cron.New()

	// Add a cron job for auto-downloading posters
	c.AddFunc(config.Global.AutoDownload.Cron, func() {
		// Call the auto download function if enabled
		if config.Global.AutoDownload.Enabled {
			download.CheckForUpdatesToPosters()
		}
	})

	// Start the cron tasks
	if config.Global.AutoDownload.Enabled {
		logging.LOG.Info(fmt.Sprintf("AutoDownload set for: %s", config.Global.AutoDownload.Cron))
		c.Start()
	}

	// Send a notification to Discord when the application starts if not in development mode
	if !strings.Contains(APP_VERSION, "dev") {
		notifications.SendDiscordAppStartNotification()
	}

	go func() {
		// Start the API server
		if err := http.ListenAndServe(fmt.Sprintf(":%d", APP_PORT), r); err != nil {
			logging.LOG.Error(fmt.Sprintf("Error starting server: %s", err.Error()))
		}
	}()

	select {}

}
