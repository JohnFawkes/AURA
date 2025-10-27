package routes

import (
	"aura/internal/api"
	"aura/internal/logging"
	routes_api "aura/internal/routes/api"
	routes_auth "aura/internal/routes/auth"
	routes_autodownload "aura/internal/routes/autodownload"
	routes_config "aura/internal/routes/config"
	routes_db "aura/internal/routes/db"
	routes_logging "aura/internal/routes/logging"
	routes_ms "aura/internal/routes/mediaserver"
	routes_mediux "aura/internal/routes/mediux"
	middleware "aura/internal/routes/middleware"
	routes_notification "aura/internal/routes/notification"
	routes_sonarr_radarr "aura/internal/routes/sonarr-radarr"
	routes_tempimages "aura/internal/routes/tempimages"
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
)

// OnboardingComplete can be set by main to perform:
// - validation preflight
// - DB init
// - cron start
// - router swap
var OnboardingComplete func()

// ensure finalize only runs once
var onboardingFinalizeOnce sync.Once

func NewRouter() *chi.Mux {
	// Create a new router
	r := chi.NewRouter()

	// Configure the router with middlewares
	middleware.Configure_Middlewares(r)

	// Add the routes to the router
	AddRoutes(r)

	// If the route is not found, return a JSON response
	r.NotFound(routes_api.NotFound)

	return r
}

// AddRoutes adds all the routes to the given router.
func AddRoutes(r *chi.Mux) {

	// Basic route to check if the server is running
	r.Get("/", routes_api.HealthCheck)

	// If config not yet valid, only expose onboarding endpoints.
	if !(api.Global_Config_Loaded && api.Global_Config_Valid) {
		logging.LOG.Warn("Configuration not valid - only onboarding routes available")
		addOnboardingRoutes(r)
		return
	} else {
		logging.LOG.Info("Configuration valid - all routes available")
	}

	routes_auth.SetTokenAuth(jwtauth.New("HS256", []byte(api.Global_Config.Mediux.Token), nil))

	r.Route("/api", func(r chi.Router) {

		r.Post("/login", routes_auth.Login)

		// Onboarding Routes
		r.Get("/onboarding/status", routes_config.OnboardingStatus)

		r.Group(func(r chi.Router) {

			r.Use(jwtauth.Verifier(routes_auth.GetTokenAuth()))
			r.Use(routes_auth.Authenticator)

			// Base API & Health Check Routes
			// Check if the API is up and running
			r.Get("/", routes_api.HealthCheck)
			r.Get("/health", routes_api.HealthCheck)

			r.Get("/health/status/mediaserver", routes_ms.GetStatus)

			// Config Routes
			r.Get("/config", routes_config.GetSanitizedConfig)
			r.Get("/config/reload", routes_config.ReloadConfig)
			r.Get("/config/mediaserver/type", routes_ms.GetType)
			r.Post("/config/get/mediaserver/sections", routes_ms.GetAllLibrariesOptions)
			r.Post("/config/update", routes_config.UpdateConfig)
			r.Post("/config/validate/mediaserver", routes_ms.ValidateNewInfo)
			r.Post("/config/validate/mediux", routes_mediux.ValidateToken)
			r.Post("/config/validate/sonarr", routes_sonarr_radarr.TestConnection)
			r.Post("/config/validate/radarr", routes_sonarr_radarr.TestConnection)
			r.Post("/config/validate/notification/discord", routes_notification.SendTest)
			r.Post("/config/validate/notification/gotify", routes_notification.SendTest)
			r.Post("/config/validate/notification/pushover", routes_notification.SendTest)

			// Log Routes
			r.Get("/logs", routes_logging.GetCurrentLogFileContents)
			r.Post("/logs/clear", routes_logging.ClearOldLogs)

			// Clear Temporary Images Route
			r.Post("/temp-images/clear", routes_tempimages.Clear)

			// Media Server Routes
			r.Get("/mediaserver/sections", routes_ms.GetAllSections)
			r.Get("/mediaserver/sections/items", routes_ms.GetAllSectionItems)
			r.Get("/mediaserver/item/{ratingKey}", routes_ms.GetItemContent)
			r.Get("/mediaserver/image/{ratingKey}/{imageType}", routes_ms.GetImage)
			r.Patch("/mediaserver/download/file", routes_ms.DownloadAndUpdate)

			// Database Routes
			r.Get("/db/get/all", routes_db.GetAllItems)
			r.Delete("/db/delete/mediaitem/{tmdbID}/{libraryTitle}", routes_db.DeleteItem)
			r.Patch("/db/update", routes_db.UpdateItem)
			r.Post("/db/add/item", routes_db.AddItem)
			r.Post("/db/force-recheck", routes_autodownload.ForceRecheckItem)

			// Mediux Routes
			r.Get("/mediux/sets/get/{itemType}/{librarySection}/{tmdbID}", routes_mediux.GetAllSets)
			r.Get("/mediux/sets/get_set/{setID}", routes_mediux.GetSetByID)
			r.Get("/mediux/image/{assetID}", routes_mediux.GetImage)
			r.Get("/mediux/user/following_hiding", routes_mediux.GetUserFollowingAndHiding)
			r.Get("/mediux/sets/get_user/sets/{username}", routes_mediux.GetAllUserSets)

		})

	})
}

func addOnboardingRoutes(r chi.Router) {

	r.Route("/api", func(r chi.Router) {
		r.Group(func(r chi.Router) {

			// Base API & Health Check Routes
			// Check if the API is up and running
			r.Get("/", routes_api.HealthCheck)
			r.Get("/health", routes_api.HealthCheck)

			// Onboarding Routes
			r.Get("/onboarding/status", routes_config.OnboardingStatus)

			// Config Routes
			r.Post("/config/get/mediaserver/sections", routes_ms.GetAllLibrariesOptions)
			r.Post("/config/update", routes_config.UpdateConfig)
			r.Post("/config/validate/mediaserver", routes_ms.ValidateNewInfo)
			r.Post("/config/validate/mediux", routes_mediux.ValidateToken)
			r.Post("/config/validate/sonarr", routes_sonarr_radarr.TestConnection)
			r.Post("/config/validate/radarr", routes_sonarr_radarr.TestConnection)
			r.Post("/config/validate/notification/discord", routes_notification.SendTest)
			r.Post("/config/validate/notification/gotify", routes_notification.SendTest)
			r.Post("/config/validate/notification/pushover", routes_notification.SendTest)

			// Finalize endpoint: call after last step when config is saved & valid
			r.Post("/onboarding/complete", func(w http.ResponseWriter, _ *http.Request) {
				if !(api.Global_Config_Loaded && api.Global_Config_Valid) {
					api.Util_Response_SendJson(w, 422, api.JSONResponse{
						Status: "error",
						Data:   "Config still invalid; cannot finalize.",
					})
					return
				}

				triggered := false
				onboardingFinalizeOnce.Do(func() {
					triggered = true
					if OnboardingComplete != nil {
						go OnboardingComplete()
					}
				})

				api.Util_Response_SendJson(w, 200, api.JSONResponse{
					Status: "success",
					Data: map[string]any{
						"finalizeTriggered": triggered,
					},
				})
			})

		})

	})

}
