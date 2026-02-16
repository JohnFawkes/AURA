package routes_validation

import (
	"aura/config"
	"aura/logging"
	"aura/mediux"
	"aura/utils/httpx"
	"net/http"
)

func ValidateMediuxInfo(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Validate Mediux Token", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	// Get the MediUX Info from the request body
	var mediuxInfo config.Config_Mediux
	Err := httpx.DecodeRequestBodyToJSON(ctx, r.Body, &mediuxInfo, "Mediux Info")
	if Err.Message != "" {
		httpx.SendResponse(w, ld, nil)
		return
	}

	// If the MediUX Token is masked, retrieve the actual token from the config
	if config.IsMaskedField(mediuxInfo.ApiToken) {
		mediuxInfo.ApiToken = config.Current.Mediux.ApiToken
	}

	isValid, Err := mediux.ValidateToken(ctx, mediuxInfo.ApiToken)
	if Err.Message != "" {
		httpx.SendResponse(w, ld, nil)
		return
	}

	var response struct {
		Valid   bool   `json:"valid"`
		Message string `json:"message"`
	}
	response.Valid = isValid
	response.Message = "Successfully validated Mediux token"

	httpx.SendResponse(w, ld, response)
}
