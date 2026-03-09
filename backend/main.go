// @title Aura API
// @version 1.0
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
package main

import (
	"aura/config"
	"aura/logging"
	"aura/routing"
	"os"
	"strings"
	"sync/atomic"
)

var (
	APP_NAME    = "aura"
	APP_VERSION = "dev"
	AUTHOR      = "xmoosex"
	LICENSE     = "MIT"
	APP_PORT    = 8888
)

var activeHandler atomic.Value

func init() {
	if strings.HasSuffix(APP_VERSION, "dev") {
		logging.SetDevMode(true)
	}
}

func main() {
	// Serve immediately with onboarding/public routes first.
	config.AppFullyLoaded = false
	config.AppVersion = APP_VERSION
	config.AppLoadingStep = "Initializing Application"
	activeHandler.Store(routing.NewRouter())

	// Start API now (non-blocking for init pipeline).
	go startAPI()

	// Run startup pipeline in background.
	go func() {
		bootStrapSuccess := runBootstrap()

		// Keep callback for onboarding finalization path.
		routing.OnboardingComplete = func() {
			preflightSuccess := runPreFlight()
			if !preflightSuccess {
				logging.LOGGER.Error().Timestamp().Msg("Preflight failed during OnboardingComplete, not swapping routers")
				return
			}
			warmupSuccess := runWarmup()
			if !warmupSuccess {
				logging.LOGGER.Fatal().Timestamp().Msg("Warmup failed during OnboardingComplete. Exiting application.")
				return
			}
			config.AppFullyLoaded = true
			activeHandler.Store(routing.NewRouter())
			logging.LOGGER.Info().Timestamp().Msg("Onboarding complete. Main routes active.")
		}

		if bootStrapSuccess {
			preflightSuccess := runPreFlight()
			if !preflightSuccess {
				config.Valid = false
				activeHandler.Store(routing.NewRouter()) // stays onboarding
				return
			}

			warmupSuccess := runWarmup()
			if !warmupSuccess {
				logging.LOGGER.Fatal().Timestamp().Msg("Warmup failed. Exiting application.")
				os.Exit(1)
			}

			config.AppFullyLoaded = true
			config.AppLoadingStep = "App Fully Loaded"
			activeHandler.Store(routing.NewRouter()) // swap to full routes
			return
		}

		// Config not loaded/valid: onboarding mode remains active.
		activeHandler.Store(routing.NewRouter())
	}()

	// Keep process alive while startAPI runs.
	select {}
}
