package routes_mediux

import (
	"aura/internal/api"
	"aura/internal/logging"
	"net/http"
	"strings"
)

func ValidateToken(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Validate MediUX Token", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	// Get the media server information from the request
	var mediuxInfo api.Config_Mediux
	Err := api.DecodeRequestBodyJSON(ctx, r.Body, &mediuxInfo, "Config_Mediux")
	if Err.Message != "" {
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}

	// If the token is masked, replace it with the existing token
	if strings.HasPrefix(mediuxInfo.Token, "***") {
		mediuxInfo.Token = api.Global_Config.Mediux.Token
	}

	Err = api.Mediux_ValidateToken(ctx, mediuxInfo.Token)
	if Err.Message != "" {
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}

	api.Util_Response_SendJSON(w, ld, map[string]bool{"valid": true})
}
