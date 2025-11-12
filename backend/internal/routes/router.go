package routes

import (
	"aura/internal/api"
	"aura/internal/logging"
	routes_api "aura/internal/routes/api"
	routes_auth "aura/internal/routes/auth"
	routes_autodownload "aura/internal/routes/autodownload"
	routes_config "aura/internal/routes/config"
	routes_db "aura/internal/routes/db"
	routes_download_queue "aura/internal/routes/download-queue"
	routes_logging "aura/internal/routes/logging"
	routes_ms "aura/internal/routes/mediaserver"
	routes_mediux "aura/internal/routes/mediux"
	middleware "aura/internal/routes/middleware"
	routes_notification "aura/internal/routes/notification"
	routes_search "aura/internal/routes/search"
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
	r.MethodNotAllowed(routes_api.MethodNotAllowed)

	return r
}

// AddRoutes adds all the routes to the given router.
func AddRoutes(r *chi.Mux) {

	// If Config not yet valid, only expose onboarding routes
	if !(api.Global_Config_Loaded && api.Global_Config_Valid) {
		logging.LOGGER.Info().Timestamp().Msg("Configuration Invalid or not loaded, adding onboarding routes only")
		addOnboardingRoutes(r)
		return
	} else {
		logging.LOGGER.Info().Timestamp().Msg("Configuration Valid, adding full routes")
	}

	routes_auth.SetTokenAuth(jwtauth.New("HS256", []byte(api.Global_Config.Mediux.Token), nil))

	// Basic route to check if the server is running
	r.Get("/", routes_api.HealthCheck)
	r.Get("/health", routes_api.HealthCheck)

	r.Route("/api", func(r chi.Router) {

		// Route for Login
		r.Post("/login", routes_auth.Login)

		// Route for Config Status
		r.Get("/config/status", routes_config.GetConfigStatus)

		r.Get("/search", routes_search.SearchHandler)

		r.Group(func(r chi.Router) {
			// Protected routes
			r.Use(jwtauth.Verifier(routes_auth.GetTokenAuth()))
			r.Use(routes_auth.Authenticator)

			// Config Routes
			r.Route("/config", func(r chi.Router) {
				r.Get("/", routes_config.GetSanitizedConfig)
				r.Get("/reload", routes_config.ReloadConfig)
				r.Post("/update", routes_config.UpdateConfig)

				r.Route("/validate", func(r chi.Router) {
					r.Post("/mediux", routes_mediux.ValidateToken)
					r.Post("/mediaserver", routes_ms.ValidateNewInfo)
					r.Post("/sonarr", routes_sonarr_radarr.TestConnection)
					r.Post("/radarr", routes_sonarr_radarr.TestConnection)
					r.Post("/notification", routes_notification.SendTest)
				})
			})

			// Logging Routes
			r.Route("/log", func(r chi.Router) {
				r.Get("/", routes_logging.GetLogContents)
				r.Post("/clear", routes_logging.ClearLogFile)
			})

			// Temp Image Routes
			r.Route("/temp-images", func(r chi.Router) {
				r.Post("/clear", routes_tempimages.ClearTempImages)
			})

			// Media Server Routes
			r.Route("/mediaserver", func(r chi.Router) {
				r.Get("/status", routes_ms.GetStatus)
				r.Get("/type", routes_ms.GetType)
				r.Post("/library-options", routes_ms.GetAllLibrariesOptions)
				r.Get("/sections", routes_ms.GetAllSections)
				r.Get("/sections/items", routes_ms.GetAllSectionItems)
				r.Get("/item", routes_ms.GetItemContent)
				r.Get("/image", routes_ms.GetImage)
				r.Patch("/download", routes_ms.DownloadAndUpdate)
				r.Get("/collection-items", routes_ms.GetCollectionItems)
				r.Post("/collection-children", routes_ms.GetAllCollectionChildren)
				r.Patch("/download-collection", routes_ms.DownloadAndUpdateCollection)
			})

			// Download Queue Routes
			r.Route("/download-queue", func(r chi.Router) {
				r.Get("/status", routes_download_queue.GetDownloadQueueLastStatus)
				r.Post("/add", routes_download_queue.AddToDownloadQueue)
				r.Get("/get-results", routes_download_queue.GetDownloadQueueResults)
				r.Delete("/delete", routes_download_queue.DeleteFromDownloadQueue)
			})

			// Mediux Routes
			r.Route("/mediux", func(r chi.Router) {
				r.Get("/sets", routes_mediux.GetAllSets)
				r.Get("/sets-by-user", routes_mediux.GetAllUserSets)
				r.Get("/set-by-id", routes_mediux.GetSetByID)
				r.Get("/image", routes_mediux.GetImage)
				r.Get("/avatar-image", routes_mediux.GetAvatarImage)
				r.Get("/user-follow-hiding", routes_mediux.GetUserFollowingAndHiding)
				r.Get("/check-link", routes_mediux.CheckMediuxLink)
			})

			// Database Routes
			r.Route("/db", func(r chi.Router) {
				r.Get("/get-all", routes_db.GetAllItems)
				r.Delete("/delete", routes_db.DeleteItem)
				r.Patch("/update", routes_db.UpdateItem)
				r.Post("/add", routes_db.AddItem)
				r.Post("/force-recheck", routes_autodownload.ForceRecheckItem)
			})

		})

	})
}

// addOnboardingRoutes adds onboarding routes to the given router.
func addOnboardingRoutes(r chi.Router) {

	// Base API & Health Check Routes
	// Check if the API is up and running
	r.Get("/", routes_api.HealthCheck)
	r.Get("/health", routes_api.HealthCheck)

	r.Route("/api", func(r chi.Router) {
		r.Group(func(r chi.Router) {

			// Config Routes
			r.Route("/config", func(r chi.Router) {
				r.Get("/", routes_config.GetSanitizedConfig)
				r.Get("/status", routes_config.GetConfigStatus)
				r.Post("/reload", routes_config.ReloadConfig)
				r.Post("/update", routes_config.UpdateConfig)

				r.Route("/validate", func(r chi.Router) {
					r.Post("/mediaserver", routes_ms.ValidateNewInfo)
					r.Post("/mediux", routes_mediux.ValidateToken)
					r.Post("/sonarr", routes_sonarr_radarr.TestConnection)
					r.Post("/radarr", routes_sonarr_radarr.TestConnection)
					r.Post("/notification", routes_notification.SendTest)
				})
			})

			// Logging Routes
			r.Route("/log", func(r chi.Router) {
				r.Get("/", routes_logging.GetLogContents)
				r.Post("/clear", routes_logging.ClearLogFile)
			})

			r.Post("/mediaserver/library-options", routes_ms.GetAllLibrariesOptions)

			// Finalize Onboarding Route
			r.Post("/onboarding/finalize", func(w http.ResponseWriter, r *http.Request) {
				ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
				logAction := ld.AddAction("Finalize Onboarding", logging.LevelInfo)
				ctx = logging.WithCurrentAction(ctx, logAction)

				if !(api.Global_Config_Loaded && api.Global_Config_Valid) {
					logAction.SetError("Configuration is not valid or not loaded", "", nil)
					ld.Status = logging.StatusError
					api.Util_Response_SendJSON(w, ld, nil)
					return
				}

				triggered := false
				onboardingFinalizeOnce.Do(func() {
					triggered = true
					if OnboardingComplete != nil {
						go OnboardingComplete()
					}
				})

				api.Util_Response_SendJSON(w, ld, map[string]any{
					"onboarding_finalized": triggered,
				})
			})

		})
	})
}
