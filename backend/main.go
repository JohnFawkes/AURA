package main

import (
	"fmt"
	"net/http"
	"os"
	"poster-setter/internal/config"
	"poster-setter/internal/database"
	"poster-setter/internal/download"
	"poster-setter/internal/logging"
	"poster-setter/internal/routes"
	"poster-setter/internal/utils"
	"strconv"

	"github.com/robfig/cron/v3"
)

var (
	Author  = "xmoosex"
	License = "MIT"
	Version = "v0.1.0"
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

	APP_PORT := os.Getenv("APP_PORT")
	if APP_PORT == "" {
		APP_PORT = "8888"
	}
	APP_PORT_INT, err := strconv.Atoi(APP_PORT)
	if err != nil {
		fmt.Printf("Error converting app_port to integer: %s\n", err.Error())
		return
	}

	utils.PrintBanner(
		Version,
		Author,
		License,
		APP_PORT_INT,
		config.Global.Logging.Level,
	)

	logging.SetLogLevel(config.Global.Logging.Level)

	init := database.InitDB()
	if !init {
		fmt.Println("Database initialization failed. Exiting...")
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
	logging.LOG.Info(fmt.Sprintf("AutoDownload set for: %s", config.Global.AutoDownload.Cron))
	c.Start()

	download.CheckForUpdatesToPosters()

	go func() {
		// Start the API server
		if err := http.ListenAndServe(fmt.Sprintf(":%d", APP_PORT_INT), r); err != nil {
			fmt.Printf("Error starting server: %s\n", err.Error())
		}
	}()

	select {}

}
