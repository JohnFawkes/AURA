package auth

import (
	"aura/internal/config"
	"aura/internal/logging"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/jwtauth/v5"
)

var tokenAuth *jwtauth.JWTAuth

// SetTokenAuth should be called at startup with the chosen secret.
func SetTokenAuth(t *jwtauth.JWTAuth) {
	tokenAuth = t
}

// TokenAuth accessor (optional use)
func TokenAuth() *jwtauth.JWTAuth {
	return tokenAuth
}

func Authenticator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Skip if auth globally disabled
		if !config.Global.Auth.Enabled {
			next.ServeHTTP(w, r)
			return
		}

		// Public login route
		if r.URL.Path == "/api/login" ||
			strings.HasPrefix(r.URL.Path, "/api/mediaserver/image/") ||
			strings.HasPrefix(r.URL.Path, "/api/mediux/image/") {
			next.ServeHTTP(w, r)
			return
		}

		if tokenAuth == nil {
			SendNotAuthenticatedResponse(w, "Auth not initialized")
			logging.LOG.Error("Authentication error: tokenAuth is nil")
			return
		}

		// jwtauth.Verifier MUST already have run to populate context
		_, claims, err := jwtauth.FromContext(r.Context())
		if err != nil {
			SendNotAuthenticatedResponse(w, "Invalid or expired token")
			logging.LOG.Error(fmt.Sprintf("Authentication error: %v", err))
			return
		}

		if sub, _ := claims["sub"].(string); sub == "" {
			SendNotAuthenticatedResponse(w, "Invalid token")
			logging.LOG.Error("Authentication error: missing sub claim")
			return
		}

		// Optional: still ensure header shape
		authz := r.Header.Get("Authorization")
		if authz == "" || !strings.HasPrefix(authz, "Bearer ") {
			SendNotAuthenticatedResponse(w, "Token is empty")
			logging.LOG.Error("Authentication error: Token header missing")
			return
		}

		// All good
		next.ServeHTTP(w, r)
	})
}

func SendNotAuthenticatedResponse(w http.ResponseWriter, message string) {
	resp := map[string]any{"message": message}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(resp)
}
