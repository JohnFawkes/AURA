package routes

import (
	"aura/internal/auth"
	"aura/internal/config"
	"aura/internal/database"
	"aura/internal/download"
	"aura/internal/logging"
	"aura/internal/mediux"
	route_auth "aura/internal/routes/auth"
	route_config "aura/internal/routes/config"
	route_health "aura/internal/routes/health"
	route_logging "aura/internal/routes/logging"
	"aura/internal/routes/middleware"
	routes_onboarding "aura/internal/routes/onboarding"
	route_tempimages "aura/internal/routes/temp-images"
	mediaserver "aura/internal/server"
	"aura/internal/utils"
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
	r.NotFound(route_health.RouteNotFound)

	return r
}

func AddRoutes(r *chi.Mux) {

	r.Get("/", route_health.HealthCheck)

	// If config not yet valid, only expose onboarding endpoints.
	if !(config.ConfigLoaded && config.ConfigValid) {
		logging.LOG.Warn("Configuration not valid - only onboarding routes available")
		addOnboardingRoutes(r)
		return
	}

	auth.SetTokenAuth(jwtauth.New("HS256", []byte(config.Global.Mediux.Token), nil))

	r.Route("/api", func(r chi.Router) {

		r.Post("/login", route_auth.HandleLogin)

		// Onboarding Routes
		r.Get("/onboarding/status", routes_onboarding.GetOnboardingStatus)

		r.Group(func(r chi.Router) {

			r.Use(jwtauth.Verifier(auth.TokenAuth()))
			r.Use(auth.Authenticator)

			// Base API Route: Check if the API is up and running
			r.Get("/", route_health.HealthCheck)

			// Health Check Routes
			r.Get("/health", route_health.HealthCheck)
			r.Get("/health/status/mediaserver", route_health.GetMediaServerStatus)
			r.Post("/health/test/notification", route_health.SendTestNotification)

			// Config Routes
			r.Get("/config", route_config.GetConfig)
			r.Get("/config/mediaserver/type", route_config.GetMediaServerType)
			r.Post("/config/update", route_config.UpdateConfig)
			r.Post("/config/validate/mediaserver", route_config.ValidateMediaServerNewInfoConnection)
			r.Post("/config/validate/mediux", route_config.ValidateMediuxToken)

			// Log Routes
			r.Get("/logs", route_logging.GetCurrentLogFile)
			r.Post("/logs/clear", route_logging.ClearLogOldFiles)

			// Clear Temporary Images Route
			r.Post("/temp-images/clear", route_tempimages.ClearTempImages)

			// Media Server Routes
			r.Get("/mediaserver/sections", mediaserver.GetAllSections)
			r.Get("/mediaserver/sections/items", mediaserver.GetAllSectionItems)
			r.Get("/mediaserver/item/{ratingKey}", mediaserver.GetItemContent)
			r.Get("/mediaserver/image/{ratingKey}/{imageType}", mediaserver.GetImageFromMediaServer)
			r.Patch("/mediaserver/download/file", mediaserver.DownloadAndUpdate)

			// Database Routes
			r.Get("/db/get/all", database.GetAllItems)
			r.Delete("/db/delete/mediaitem/{ratingKey}", database.DeleteMediaItemFromDatabase)
			r.Patch("/db/update", database.UpdateSavedSetTypesForItem)
			r.Post("/db/add/item", mediaserver.AddItemToDatabase)
			r.Post("/db/force/recheck", download.ForceRecheckItem)

			// Mediux Routes
			r.Get("/mediux/sets/get/{itemType}/{librarySection}/{ratingKey}/{tmdbID}", mediux.GetAllSets)
			r.Get("/mediux/sets/get_set/{setID}", mediux.GetSetByID)
			r.Get("/mediux/image/{assetID}", mediux.GetMediuxImage)
			r.Get("/mediux/user/following_hiding", mediux.GetUserFollowingAndHiding)
			r.Get("/mediux/sets/get_user/sets/{username}", mediux.GetAllUserSets)

		})

	})
}

func addOnboardingRoutes(r chi.Router) {

	r.Route("/api", func(r chi.Router) {
		r.Group(func(r chi.Router) {

			// Base API Route: Check if the API is up and running
			r.Get("/", route_health.HealthCheck)

			// Health Check Routes
			r.Get("/health", route_health.HealthCheck)

			// Onboarding Routes
			r.Get("/onboarding/status", routes_onboarding.GetOnboardingStatus)

			// Config Routes
			r.Post("/config/validate/mediaserver", route_config.ValidateMediaServerNewInfoConnection)
			r.Post("/config/validate/mediux", route_config.ValidateMediuxToken)
			r.Post("/config/update", route_config.UpdateConfig)

			// Finalize endpoint: call after last step when config is saved & valid
			r.Post("/onboarding/complete", func(w http.ResponseWriter, _ *http.Request) {
				if !(config.ConfigLoaded && config.ConfigValid) {
					utils.SendJsonResponse(w, 422, utils.JSONResponse{
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

				utils.SendJsonResponse(w, 200, utils.JSONResponse{
					Status: "success",
					Data: map[string]any{
						"finalizeTriggered": triggered,
					},
				})
			})

		})

	})

}
