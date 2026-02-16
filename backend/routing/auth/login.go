package routes_auth

import (
	"aura/config"
	"aura/logging"
	"aura/utils/httpx"
	"net/http"
	"time"

	"github.com/alexedwards/argon2id"
)

type loginRequest struct {
	Password string `json:"password"`
}

// Status godoc
// @Summary      Auth Login
// @Description  Obtain a JWT token by providing valid credentials
// @Tags         auth
// @Produce      json
// @Success      200  {object}  routes_auth.loginRequest
// @Router       /api/login [post]
func Login(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("User Login", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	if !config.Current.Auth.Enabled {
		httpx.SendResponse(w, ld, "Authentication is disabled")
		return
	}

	if TokenAuth == nil {
		logAction.SetError("Authentication not configured", "The authentication system is not properly configured", nil)
		httpx.SendResponse(w, ld, nil)
		return
	}

	var loginReq loginRequest
	Err := httpx.DecodeRequestBodyToJSON(ctx, r.Body, &loginReq, "Login Request")
	if Err.Message != "" {
		httpx.SendResponse(w, ld, nil)
		return
	}

	// Compare password
	ok, err := argon2id.ComparePasswordAndHash(loginReq.Password, config.Current.Auth.Password)
	if err != nil || !ok {
		logAction.SetError("Invalid credentials", "The provided password is incorrect", map[string]any{
			"error": err,
		})
		httpx.SendResponse(w, ld, nil)
		return
	}

	// Build claims
	claims := map[string]any{
		"sub": "aura",
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	}

	// Use jwtauth to create token (consistent with verifier)
	_, signedToken, err := TokenAuth.Encode(claims)
	if err != nil {
		logAction.SetError("Failed to generate token", "An error occurred while generating the JWT token", map[string]any{
			"error": err,
		})
		httpx.SendResponse(w, ld, nil)
		return
	}

	logAction.AppendResult("token_generated", true)

	var response struct {
		Token string `json:"token"`
	}
	response.Token = signedToken
	httpx.SendResponse(w, ld, response)
}
