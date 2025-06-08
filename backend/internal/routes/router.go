package routes

import (
	"aura/internal/database"
	"aura/internal/mediux"
	"aura/internal/notifications"
	"aura/internal/routes/health"
	"aura/internal/routes/middleware"
	tempimages "aura/internal/routes/temp-images"
	mediaserver "aura/internal/server"

	"github.com/go-chi/chi/v5"
)

func NewRouter() *chi.Mux {
	// Create a new router
	r := chi.NewRouter()

	// Configure the router with middlewares
	middleware.Configure_Middlewares(r)

	// Add the routes to the router
	AddRoutes(r)

	// If the route is not found, return a JSON response
	r.NotFound(health.NotFound)

	return r
}

func AddRoutes(r *chi.Mux) {

	r.Get("/", health.HealthCheck)

	r.Route("/api", func(r chi.Router) {
		// Base API Route: Check if the API is up and running
		r.Get("/", health.HealthCheck)

		// Health Check Routes
		r.Get("/health", health.HealthCheck)
		r.Get("/health/status/mediaserver", mediaserver.GetMediaServerStatus)
		r.Post("/health/status/notification", notifications.SendTestNotification)

		// Config Routes
		r.Get("/config", health.GetConfig)

		// Log Routes
		r.Get("/logs", health.GetCurrentLogFile)

		// Clear Temporary Images Route
		r.Post("/temp-images/clear", tempimages.ClearTempImages)

		// Media Server Routes
		r.Get("/mediaserver/type", health.GetMediaServerType)
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

		// Mediux Routes
		r.Get("/mediux/sets/get/{itemType}/{librarySection}/{ratingKey}/{tmdbID}", mediux.GetAllSets)
		r.Get("/mediux/sets/get_set/{setID}", mediux.GetShowSetByID)
		r.Get("/mediux/image/{assetID}", mediux.GetMediuxImage)
		r.Get("/mediux/user/following_hiding", mediux.GetUserFollowingAndHiding)
		r.Get("/mediux/sets/get_user/sets/{username}", mediux.GetAllUserSets)

	})
}
