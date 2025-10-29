package routes_middleware

import (
	"aura/internal/logging"
	"net/http"
	"regexp"
	"strings"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/rs/cors"
)

/*
	Configure_Middlewares

Sets up the middleware stack for the given router.

Middlewares included:
- Custom Logging Middleware: Logs requests using a custom log formatter.
- CORS Middleware: Allows CORS for all origins (replace this with specific origins).
- RealIP Middleware: Gets the client's real public IP address from the request headers.
- StripSlashes Middleware: Strips slashes to no slash URL versions.
- Panic Recovery Middleware: Recovers from panics and returns a 500 error.

Parameters:

- r: The chi.Mux router to configure middlewares for.
*/
func Configure_Middlewares(r *chi.Mux) {

	AllowedOrigins := []string{"*"}

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

	// Logging Middleware: Custom logging middleware
	r.Use(LoggingMiddleware)

	// Middleware for recovering from panics
	r.Use(middleware.Recoverer)
}

type responseWriterWithBytes struct {
	http.ResponseWriter
	bytesWritten int64
}

func (w *responseWriterWithBytes) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.bytesWritten += int64(n)
	return n, err
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		wrapped := &responseWriterWithBytes{ResponseWriter: w}

		ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)

		// Skip logging for certain paths/methods
		if logging.ShouldSkipLogging(r, ld) {
			next.ServeHTTP(w, r)
			return
		}

		ld.Route = &logging.LogRouteInfo{
			Method: r.Method,
			Path:   r.URL.Path,
			IP:     getLogIP(r),
		}

		if len(r.URL.Query()) > 0 {
			ld.Route.Params = r.URL.Query()
		}

		ctx = logging.WithLogData(r.Context(), ld)
		next.ServeHTTP(wrapped, r.WithContext(ctx))

		// Set response bytes after handler
		ld.Route.ResponseBytes = wrapped.bytesWritten
		ld.Complete()
		ld.Log()
	})
}

func removeExtraSlashes(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.ReplaceAll(r.URL.Path, "//", "/")
		next.ServeHTTP(w, r)
	})
}

func getLogIP(Request *http.Request) string {

	// Get the IP address of the client
	ip := Request.RemoteAddr
	if forwarded := Request.Header.Get("X-Forwarded-For"); forwarded != "" {
		ip = forwarded // If the X-Forwarded-For header is present, use it instead
	}

	// Remove the port number from the IP address
	ip = regexp.MustCompile(`:\d+$`).ReplaceAllString(ip, "")

	// If the ip is ::1 (change it to localhost)
	if ip == "[::1]" || ip == "::1" {
		ip = "localhost"
	}

	// Return the IP address
	return ip
}
