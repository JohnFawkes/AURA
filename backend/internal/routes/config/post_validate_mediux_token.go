package route_config

import (
	"aura/internal/config"
	"aura/internal/logging"
	"aura/internal/modals"
	"aura/internal/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

func ValidateMediuxToken(w http.ResponseWriter, r *http.Request) {
	logging.LOG.Trace(r.URL.Path)
	startTime := time.Now()
	Err := logging.NewStandardError()

	// Get the media server information from the request
	var mediuxInfo modals.Config_Mediux
	if err := json.NewDecoder(r.Body).Decode(&mediuxInfo); err != nil {
		Err.Message = "Failed to decode request body"
		Err.HelpText = "Ensure the request body is valid JSON"
		Err.Details = fmt.Sprintf("Error: %v", err)
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}

	// Store the current values for Token
	currentToken := ""
	if config.Global.Mediux.Token != "" {
		currentToken = config.Global.Mediux.Token
	}

	// Change the Global values temporarily
	if !strings.HasPrefix(mediuxInfo.Token, "***") {
		config.Global.Mediux.Token = mediuxInfo.Token
	}

	// Restore the previous values
	defer func() {
		config.Global.Mediux.Token = currentToken
	}()

	Err = utils.ValidateMediuxToken(mediuxInfo.Token)
	if Err.Message != "" {
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		// Restore the previous values
		config.Global.Mediux.Token = currentToken
		return
	}

	// Respond with a success message
	utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
		Status:  "success",
		Elapsed: utils.ElapsedTime(startTime),
		Data:    "New MediUX token is ok",
	})
}
