package routes

import (
	"net/http"
	"os"
	"path/filepath"
	"poster-setter/internal/database"
	"poster-setter/internal/mediux"
	"poster-setter/internal/routes/health"
	"poster-setter/internal/routes/middleware"
	tempimages "poster-setter/internal/routes/temp-images"
	mediaserver "poster-setter/internal/server"

	"github.com/go-chi/chi/v5"
)

func NewRouter() *chi.Mux {
	// Create a new router
	r := chi.NewRouter()

	// Configure the router with middlewares
	middleware.Configure_Middlewares(r)

	// Add the routes to the router
	AddRoutes(r)

	// Serve static files
	ServeStaticFiles(r)

	// If the route is not found, return a JSON response
	r.NotFound(health.NotFound)

	return r
}

func AddRoutes(r *chi.Mux) {
	r.Route("/api", func(r chi.Router) {
		// Base API Route: Check if the API is up and running
		r.Get("/", health.HealthCheck)

		// Config Routes
		r.Get("/config", health.GetConfig)

		// Log Routes
		r.Get("/logs", health.GetCurrentLogFile)

		// Clear Temporary Images Route
		r.Post("/temp-images/clear", tempimages.ClearTempImages)

		// Media Server Routes
		r.Get("/mediaserver/sections/all", mediaserver.GetAllSections)
		r.Get("/mediaserver/item/{ratingKey}", mediaserver.GetItemContent)
		r.Get("/mediaserver/image/{ratingKey}/{imageType}", mediaserver.GetImageFromMediaServer)
		r.Post("/mediaserver/update/send", mediaserver.GetUpdateSetFromClient)
		r.Get("/mediaserver/update/set/{ratingKey}", mediaserver.UpdateItemPosters)

		// Database Routes
		r.Get("/db/get/all", database.GetAllItems)
		r.Delete("/db/delete/{ratingKey}", database.DeleteItemFromDatabase)
		r.Patch("/db/update/{ratingKey}", database.UpdateSelectedTypesForItem)

		// Mediux Routes
		r.Get("/mediux/sets/get/{itemType}/{tmdbID}", mediux.GetAllSets)
		r.Get("/mediux/image/{assetID}", mediux.GetMediuxImage)

	})
}

func ServeStaticFiles(r *chi.Mux) {
	// Get the current working directory
	workingDir, err := os.Getwd()
	if err != nil {
		panic("Failed to get current working directory: " + err.Error())
	}

	// Define the path to the static files directory
	staticDir := filepath.Join(workingDir, "..", "frontend", "dist")

	// Check if the directory exists
	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		panic("Static files directory not found: " + staticDir)
	}

	// Serve static files
	fs := http.FileServer(http.Dir(staticDir))
	r.Handle("/*", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		http.StripPrefix("/", fs)
		// Check if the requested file exists
		filePath := filepath.Join(staticDir, req.URL.Path)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			// If the file doesn't exist, serve index.html
			http.ServeFile(w, req, filepath.Join(staticDir, "index.html"))
			return
		}
		// Otherwise, serve the requested file
		fs.ServeHTTP(w, req)
	}))
}
