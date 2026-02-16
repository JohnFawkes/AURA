package routes_auth

import (
	"aura/database"
	"aura/logging"
	"context"

	"github.com/go-chi/jwtauth/v5"
)

// TokenAuth is the global JWTAuth instance for handling JWT tokens.
var TokenAuth *jwtauth.JWTAuth

// SetTokenAuth sets the global JWTAuth instance for handling JWT tokens.
func SetTokenAuth(t *jwtauth.JWTAuth) {
	TokenAuth = t
}

func GetTokenAuthSecret() (secret string, Err logging.LogErrorInfo) {
	ctx, ld := logging.CreateLoggingContext(context.Background(), "Application Auth Token Secret Retrieval")
	defer ld.Log()

	logAction := ld.AddAction("Auth Setup", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)
	defer logAction.Complete()

	// Query the database to get the token secret
	secret, Err = database.GetAuthTokenSecret(ctx)
	if Err.Message != "" {
		return "", Err
	}
	return secret, Err
}
