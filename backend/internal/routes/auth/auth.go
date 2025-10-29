package routes_auth

import (
	"aura/internal/api"
	"aura/internal/logging"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/jwtauth/v5"
)

// TokenAuth is the global JWTAuth instance for handling JWT tokens.
var TokenAuth *jwtauth.JWTAuth

// SetTokenAuth sets the global JWTAuth instance for handling JWT tokens.
func SetTokenAuth(t *jwtauth.JWTAuth) {
	TokenAuth = t
}

// GetTokenAuth returns the global JWTAuth instance for handling JWT tokens.
func GetTokenAuth() *jwtauth.JWTAuth {
	return TokenAuth
}

// Authenticator is a middleware that checks for a valid JWT token in the Authorization header.
// If the token is valid, it calls the next handler; otherwise, it responds with a 401 Unauthorized error.
//
// It skips authentication for the following routes:
//   - /api/login
//   - /api/mediaserver/image/*
//   - /api/mediux/image/*
//
// If authentication is globally disabled in the configuration, it allows all requests to pass through.
func Authenticator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip if auth globally disabled
		if !api.Global_Config.Auth.Enabled {
			next.ServeHTTP(w, r)
			return
		}

		// Public login route
		if r.URL.Path == "/api/login" ||
			strings.HasPrefix(r.URL.Path, "/api/mediaserver/image") ||
			strings.HasPrefix(r.URL.Path, "/api/mediux/image") {
			next.ServeHTTP(w, r)
			return
		}

		ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
		logAction := ld.AddAction("Authenticate Request", logging.LevelInfo)
		ctx = logging.WithCurrentAction(ctx, logAction)
		defer logAction.Complete()

		// Ensure TokenAuth is initialized
		if TokenAuth == nil {
			sendNotAuthenticatedResponse(w, "Auth not initialized")
			logAction.SetError("Auth not initialized", "The authentication system is not set up", nil)
			return
		}

		// jwtauth.Verifier MUST already have run to populate context
		_, claims, err := jwtauth.FromContext(r.Context())
		if err != nil {
			sendNotAuthenticatedResponse(w, "Invalid or expired token")
			logAction.SetError("Invalid or expired token", err.Error(), nil)
			return
		}

		if sub, _ := claims["sub"].(string); sub == "" {
			sendNotAuthenticatedResponse(w, "Invalid token")
			logAction.SetError("Invalid token", "Token missing 'sub' claim", nil)
			return
		}

		// Ensure header shape
		authz := r.Header.Get("Authorization")
		if authz == "" || !strings.HasPrefix(authz, "Bearer ") {
			sendNotAuthenticatedResponse(w, "Invalid Authorization header")
			logAction.SetError("Invalid Authorization header", "Authorization header missing or malformed", nil)
			return
		}

		// Token is valid, proceed to next handler
		next.ServeHTTP(w, r)
	})
}

// sendNotAuthenticatedResponse sends a 401 Unauthorized response with a JSON message.
func sendNotAuthenticatedResponse(w http.ResponseWriter, message string) {
	resp := map[string]any{"message": message}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(resp)
}
