package route_auth

import (
	"aura/internal/auth"
	"aura/internal/config"
	"aura/internal/logging"
	"aura/internal/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/alexedwards/argon2id"
)

type loginRequest struct {
	Password string `json:"password"`
}

func HandleLogin(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	Err := logging.NewStandardError()

	if !config.Global.Auth.Enabled {
		Err.Message = "Authentication disabled"
		Err.HelpText = "Enable Auth in the server config."
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}

	if auth.TokenAuth() == nil {
		Err.Message = "Auth not initialized"
		Err.HelpText = "SetTokenAuth was not called at startup."
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Err.Message = "Failed to decode request body"
		Err.HelpText = "Ensure the body is valid JSON: {\"password\":\"...\"}"
		Err.Details = fmt.Sprintf("Error: %s", err.Error())
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}

	// Compare password
	ok, err := argon2id.ComparePasswordAndHash(req.Password, config.Global.Auth.Password)
	if err != nil {
		Err.Message = "Password verification error"
		Err.HelpText = "Stored hash may be invalid. Restart application and try again."
		Err.Details = fmt.Sprintf("Error: %s", err.Error())
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}
	if !ok {
		Err.Message = "Invalid credentials"
		Err.HelpText = "Password incorrect."
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}

	// Build claims
	claims := map[string]any{
		"sub": "aura",
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	}

	// Use jwtauth to create token (consistent with verifier)
	_, signedToken, err := auth.TokenAuth().Encode(claims)
	if err != nil {
		Err.Message = "Failed to generate token"
		Err.HelpText = "Check JWT secret configuration."
		Err.Details = fmt.Sprintf("Error: %s", err.Error())
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}

	utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
		Status:  "success",
		Elapsed: utils.ElapsedTime(startTime),
		Data:    map[string]any{"token": signedToken},
	})
}
