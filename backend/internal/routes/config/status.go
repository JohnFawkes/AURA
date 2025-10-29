package routes_config

import (
	"aura/internal/api"
	"aura/internal/logging"
	"net/http"
)

// onboardingStatus represents the current status of the onboarding process and app config
type onboardingStatus struct {
	ConfigLoaded bool       `json:"configLoaded"` // Whether a config file is loaded
	ConfigValid  bool       `json:"configValid"`  // Whether the loaded config is valid
	NeedsSetup   bool       `json:"needsSetup"`   // Whether the app needs initial setup
	CurrentSetup api.Config `json:"currentSetup"` // The current (sanitized) configuration
}

func GetConfigStatus(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Get Config Status", logging.LevelTrace)
	ctx = logging.WithCurrentAction(ctx, logAction)

	config := api.Global_Config // Make a copy of the current config for sanitization

	status := onboardingStatus{
		ConfigLoaded: api.Global_Config_Loaded,
		ConfigValid:  (api.Global_Config_Valid && api.Global_Config_MediuxValid && api.Global_Config_MediaServerValid),
		NeedsSetup:   !(api.Global_Config_Loaded && api.Global_Config_Valid && api.Global_Config_MediuxValid && api.Global_Config_MediaServerValid),
		CurrentSetup: config.Sanitize(ctx),
	}

	api.Util_Response_SendJSON(w, ld, status)
}
