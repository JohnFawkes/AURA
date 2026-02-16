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

	// Run Bootstrap
	bootStrapSuccess := runBootstrap()

	// Setup the OnboardingComplete callback to swap routers
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
		// Swap to the full router
		activeHandler.Store(routing.NewRouter())
		logging.LOGGER.Info().Timestamp().Msg("Onboarding complete. Main routes active.")
	}

	// If the config is already loaded and valid, finish preflight
	if bootStrapSuccess {
		preflightSuccess := runPreFlight()
		if !preflightSuccess {
			// Preflight failed, force onboarding
			config.Valid = false
			activeHandler.Store(routing.NewRouter())
		} else {
			warmupSuccess := runWarmup()
			if !warmupSuccess {
				// Kill Application
				logging.LOGGER.Fatal().Timestamp().Msg("Warmup failed. Exiting application.")
				os.Exit(1)
			}
			config.AppFullyLoaded = true
			// Swap to the full router
			activeHandler.Store(routing.NewRouter())
		}
	} else {
		// Load Onboarding Router
		config.AppFullyLoaded = true
		activeHandler.Store(routing.NewRouter())
	}

	// Start the API
	startAPI()
}
