package middleware

import (
	"net/http"
	"poster-setter/internal/logging"
	"strings"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/rs/cors"
)

func Configure_Middlewares(r *chi.Mux) {

	// Custom Logging Middleware: Log requests using a custom log formatter
	r.Use(middleware.RequestLogger(&logging.LogFormatter{}))

	AllowedOrigins := []string{"http://localhost:3000"}
	AllowedOrigins = append(AllowedOrigins, "http://10.1.1.30:3000")

	// CORS Middleware: Allow CORS for all origins (replace this with specific origins)
	cors := cors.New(cors.Options{
		AllowedOrigins:   AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "DELETE", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	})
	r.Use(cors.Handler)

	// RealIP Middleware: Get the client's real public IP address from the request headers
	r.Use(middleware.RealIP)

	// StripSlashes Middleware: Strip slashes to no slash URL versions
	r.Use(middleware.StripSlashes)

	// Custom middleware to remove extra slashes
	r.Use(removeExtraSlashes)

	// Middleware for recovering from panics
	r.Use(middleware.Recoverer)
}

func removeExtraSlashes(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.Replace(r.URL.Path, "//", "/", -1)
		next.ServeHTTP(w, r)
	})
}
