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

	"github.com/robfig/cron/v3"
)

var (
	Author      = "xmoosex"
	License     = "MIT"
	APP_VERSION = "dev"
	APP_PORT    = 8888
)

func main() {

	// Load the configuration file
	// If the config file is not found, exit the program
	_, err := config.LoadYamlConfig()
	if err != nil {
		// Exit the program if the config file is not found
		fmt.Printf("Error: %s\n", err.Error())
		return
	}

	utils.PrintBanner(
		APP_VERSION,
		Author,
		License,
		APP_PORT,
		config.Global.Logging.Level,
	)

	logging.SetLogLevel(config.Global.Logging.Level)

	init := database.InitDB()
	if !init {
		fmt.Println("Database initialization failed. Exiting...")
		return
	}

	logErr := mediaserver_shared.InitUserID()
	if logErr.Err != nil {
		logging.LOG.ErrorWithLog(logErr)
		return
	}

	// Create a new router
	r := routes.NewRouter()

	// Create a new cron instance
	c := cron.New()

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

	notifications.SendDiscordAppStartNotification()
	go func() {
		// Start the API server
		if err := http.ListenAndServe(fmt.Sprintf(":%d", APP_PORT), r); err != nil {
			logging.LOG.Error(fmt.Sprintf("Error starting server: %s", err.Error()))
		}
	}()

	select {}

}
