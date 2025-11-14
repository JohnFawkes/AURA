package api

import (
	"aura/internal/logging"
	"context"
)

func (config *Config) PrintDetails() {
	ctx, ld := logging.CreateLoggingContext(context.Background(), "Config - Print Details")
	logAction := ld.AddAction("Print Config Details", logging.LevelDebug)
	ctx = logging.WithCurrentAction(ctx, logAction)
	defer logAction.Complete()

	event := logging.LOGGER.Debug() // default level

	sanitizedConfig := config.Sanitize(ctx)

	event.Timestamp().
		Interface("Authentication Details", sanitizedConfig.Auth).
		Interface("Logging", sanitizedConfig.Logging).
		Interface("Media Server", sanitizedConfig.MediaServer).
		Interface("MediUX", sanitizedConfig.Mediux).
		Interface("Auto Download", sanitizedConfig.AutoDownload).
		Interface("Images", sanitizedConfig.Images).
		Interface("TMDB", sanitizedConfig.TMDB).
		Interface("Labels and Tags", sanitizedConfig.LabelsAndTags).
		Interface("Notifications", sanitizedConfig.Notifications).
		Interface("Sonarr and Radarr Apps", sanitizedConfig.SonarrRadarr).
		Msg("Configuration Details")
}
