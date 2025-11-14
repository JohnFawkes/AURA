package api

import (
	"aura/internal/logging"
	"context"
	"fmt"
)

type Interface_SonarrRadarr interface {

	// Test the connection to the Sonarr/Radarr instance
	TestConnection(ctx context.Context, app Config_SonarrRadarrApp) (bool, logging.LogErrorInfo)

	// Get the item ID from the TMDB ID
	GetItemInfoFromTMDBID(ctx context.Context, app Config_SonarrRadarrApp, tmdbID int) (any, logging.LogErrorInfo)

	// Get All Configured Tags
	GetAllTags(ctx context.Context, app Config_SonarrRadarrApp) ([]SonarrRadarrTag, logging.LogErrorInfo)

	// Add New Tags
	AddNewTags(ctx context.Context, app Config_SonarrRadarrApp, tags []string) ([]SonarrRadarrTag, logging.LogErrorInfo)

	// Handle tags for the Sonarr/Radarr instance
	HandleTags(ctx context.Context, app Config_SonarrRadarrApp, item MediaItem, selectedTypes []string) logging.LogErrorInfo
}

type SonarrApp struct{}
type RadarrApp struct{}

type SonarrRadarrTag struct {
	Label string `json:"label"`
	ID    int    `json:"id"`
}

func NewSonarrRadarrInterface(ctx context.Context, app Config_SonarrRadarrApp) (Interface_SonarrRadarr, logging.LogErrorInfo) {
	var interfaceSR Interface_SonarrRadarr
	switch app.Type {
	case "Sonarr":
		interfaceSR = &SonarrApp{}
	case "Radarr":
		interfaceSR = &RadarrApp{}
	default:
		_, logAction := logging.AddSubActionToContext(ctx, "Creating Sonarr/Radarr Interface", logging.LevelTrace)
		logAction.SetError(fmt.Sprintf("Unsupported Sonarr/Radarr Type: '%s'", app.Type),
			"Ensure the Sonarr/Radarr Type is set to either 'Sonarr' or 'Radarr' in the config file",
			nil)
		return nil, *logAction.Error
	}

	// Make sure all required info is set
	Err := SR_MakeSureAllInfoIsSet(ctx, app)
	if Err.Message != "" {
		return nil, Err
	}

	return interfaceSR, logging.LogErrorInfo{}
}

func SR_MakeSureAllInfoIsSet(ctx context.Context, app Config_SonarrRadarrApp) logging.LogErrorInfo {
	// Make sure the Type, Library, URL, and API Key are set
	if app.Type == "" || app.Library == "" || app.URL == "" || app.APIKey == "" {
		_, logAction := logging.AddSubActionToContext(ctx, "Making Sure Sonarr/Radarr Info Is Set", logging.LevelTrace)
		logAction.SetError("Sonarr/Radarr App Info Incomplete",
			"Ensure the Sonarr/Radarr Type, Library, URL, and API Key are set in the config file",
			map[string]any{
				"Type":    app.Type,
				"Library": app.Library,
				"URL":     app.URL,
				"APIKey":  app.APIKey,
			})
		return *logAction.Error
	}
	return logging.LogErrorInfo{}
}
