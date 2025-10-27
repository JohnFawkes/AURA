package routes_auth

import (
	"aura/internal/api"
	"aura/internal/logging"
	"encoding/json"
	"fmt"
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
	startTime := time.Now()
	Err := logging.NewStandardError()

	if !api.Global_Config.Auth.Enabled {
		Err.Message = "Authentication disabled"
		Err.HelpText = "Enable Auth in the server config."
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		return
	}

	if GetTokenAuth() == nil {
		Err.Message = "Auth not initialized"
		Err.HelpText = "SetTokenAuth was not called at startup."
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		return
	}

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Err.Message = "Failed to decode request body"
		Err.HelpText = "Ensure the body is valid JSON: {\"password\":\"...\"}"
		Err.Details = fmt.Sprintf("Error: %s", err.Error())
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		return
	}

	// Compare password
	ok, err := argon2id.ComparePasswordAndHash(req.Password, api.Global_Config.Auth.Password)
	if err != nil {
		Err.Message = "Password verification error"
		Err.HelpText = "Stored hash may be invalid. Restart application and try again."
		Err.Details = fmt.Sprintf("Error: %s", err.Error())
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		return
	}
	if !ok {
		Err.Message = "Invalid credentials"
		Err.HelpText = "Password incorrect."
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
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
		Err.Message = "Failed to generate token"
		Err.HelpText = "Check JWT secret configuration."
		Err.Details = fmt.Sprintf("Error: %s", err.Error())
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		return
	}

	api.Util_Response_SendJson(w, http.StatusOK, api.JSONResponse{
		Status:  "success",
		Elapsed: api.Util_ElapsedTime(startTime),
		Data:    map[string]any{"token": signedToken},
	})
}
