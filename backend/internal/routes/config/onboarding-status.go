package routes_config

import (
	"aura/internal/api"
	"aura/internal/logging"

	"net/http"
	"time"
)

// onboardingStatus represents the current status of the onboarding process and app config
type onboardingStatus struct {
	ConfigLoaded bool       `json:"configLoaded"` // Whether a config file is loaded
	ConfigValid  bool       `json:"configValid"`  // Whether the loaded config is valid
	NeedsSetup   bool       `json:"needsSetup"`   // Whether the app needs initial setup
	CurrentSetup api.Config `json:"currentSetup"` // The current (sanitized) configuration
}

// OnboardingStatus handles requests to check the onboarding status of the application.
//
// Method: GET
//
// Endpoint: /api/onboarding/status
//
// It responds with a JSON object containing the onboarding status, including whether the configuration is loaded,
// valid, if setup is needed, and the current sanitized configuration.
func OnboardingStatus(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	logging.LOG.Trace(r.URL.Path)

	config := api.Global_Config // Make a copy of the current config for sanitization
	config.ValidateConfig()

	status := onboardingStatus{
		ConfigLoaded: api.Global_Config_Loaded,
		ConfigValid:  (api.Global_Config_Valid && api.Global_Config_MediuxValid && api.Global_Config_MediaServerValid),
		NeedsSetup:   !(api.Global_Config_Loaded && api.Global_Config_Valid && api.Global_Config_MediuxValid && api.Global_Config_MediaServerValid),
		CurrentSetup: config.Sanitize(),
	}

	api.Util_Response_SendJson(w, http.StatusOK, api.JSONResponse{
		Status:  "success",
		Elapsed: api.Util_ElapsedTime(start),
		Data:    status,
	})
}
