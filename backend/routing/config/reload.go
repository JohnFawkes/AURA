package routes_config

import (
	"aura/config"
	"aura/logging"
	"aura/utils/httpx"
	"net/http"
)

func Reload(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Reload Config", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	// Reload the config file
	config.LoadYAML(ctx)

	// Print the config details (sanitized)
	config.Current.PrintDetails()

	// Sanitize the config before sending it back
	sanitizedConfig := config.Current.SanitizeConfig(ctx)

	status := AppConfigStatus{
		ConfigLoaded:    config.Loaded,
		ConfigValid:     (config.Valid && config.MediuxValid && config.MediaServerValid),
		NeedsSetup:      !(config.Loaded && config.Valid && config.MediuxValid && config.MediaServerValid),
		CurrentSetup:    *sanitizedConfig,
		MediaServerName: config.MediaServerName,
	}

	httpx.SendResponse(w, ld, status)
}
