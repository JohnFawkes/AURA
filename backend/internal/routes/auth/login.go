package routes_auth

import (
	"aura/internal/api"
	"aura/internal/logging"
	"net/http"
	"time"

	"github.com/alexedwards/argon2id"
)

// loginRequest represents the expected JSON structure for login requests.
// It contains the password field.
type loginRequest struct {
	Password string `json:"password"`
}

// Login handles user login requests.
//
// Method: POST
//
// Endpoint: /api/login
//
// It expects a JSON body with the password field.
//
// If authentication is successful, it responds with a JWT token.
//
// If authentication fails, it responds with a 401 Unauthorized error.
func Login(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("User Login", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	if !api.Global_Config.Auth.Enabled {
		api.Util_Response_SendJSON(w, ld, "Authentication is disabled")
		return
	}

	if GetTokenAuth() == nil {
		logAction.SetError("Auth not initialized", "The authentication system is not set up", nil)
		api.Util_Response_SendJSON(w, ld, "Authentication system not initialized")
		return
	}

	var loginReq loginRequest
	Err := api.DecodeRequestBodyJSON(ctx, r.Body, &loginReq, "loginRequest")
	if Err.Message != "" {
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}

	// Compare password
	ok, err := argon2id.ComparePasswordAndHash(loginReq.Password, api.Global_Config.Auth.Password)
	if err != nil || !ok {
		logAction.SetError("Invalid credentials", "The provided password is incorrect", map[string]any{
			"error": err,
		})
		api.Util_Response_SendJSON(w, ld, "Invalid password")
		return
	}

	// Build claims
	claims := map[string]any{
		"sub": "aura",
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	}

	// Use jwtauth to create token (consistent with verifier)
	_, signedToken, err := GetTokenAuth().Encode(claims)
	if err != nil {
		logAction.SetError("Failed to generate token", "An error occurred while generating the JWT token", map[string]any{
			"error": err,
		})
		api.Util_Response_SendJSON(w, ld, "Failed to generate token")
		return
	}
	logAction.AppendResult("token_generated", true)
	api.Util_Response_SendJSON(w, ld, map[string]any{"token": signedToken})
}
