package api

import (
	"aura/internal/logging"
	"fmt"
)

type Interface_SonarrRadarr interface {

	// Test the connection to the Sonarr/Radarr instance
	TestConnection(app Config_SonarrRadarrApp) (bool, logging.StandardError)

	// Get the item ID from the TMDB ID
	GetItemInfoFromTMDBID(app Config_SonarrRadarrApp, tmdbID int) (any, logging.StandardError)

	// Get All Configured Tags
	GetAllTags(app Config_SonarrRadarrApp) ([]SonarrRadarrTag, logging.StandardError)

	// Add New Tags
	AddNewTags(app Config_SonarrRadarrApp, tags []string) ([]SonarrRadarrTag, logging.StandardError)

	// Handle tags for the Sonarr/Radarr instance
	HandleTags(app Config_SonarrRadarrApp, item MediaItem) logging.StandardError
}

type SonarrApp struct{}
type RadarrApp struct{}

type SonarrRadarrTag struct {
	Label string `json:"label"`
	ID    int    `json:"id"`
}

func SR_GetSonarrRadarrInterface(app Config_SonarrRadarrApp) (Interface_SonarrRadarr, logging.StandardError) {
	var interfaceSR Interface_SonarrRadarr
	switch app.Type {
	case "Sonarr":
		interfaceSR = &SonarrApp{}
	case "Radarr":
		interfaceSR = &RadarrApp{}
	default:
		Err := logging.NewStandardError()
		Err.Message = "Unsupported Sonarr/Radarr Type"
		Err.HelpText = "Ensure the Sonarr/Radarr Type is set to either 'Sonarr' or 'Radarr' in the configuration."
		return nil, Err
	}

	return interfaceSR, logging.StandardError{}
}

func SR_MakeSureAllInfoIsSet(app Config_SonarrRadarrApp) logging.StandardError {
	Err := logging.NewStandardError()
	// Make sure the Type is set
	if app.Type == "" {
		Err.Message = "Sonarr/Radarr Type is not set"
		Err.HelpText = "Please set the Type in the configuration file"
		Err.Details = map[string]any{
			"error": "Type is empty",
			"app":   app,
		}
		logging.LOG.ErrorWithLog(Err)
		return Err
	}

	// Make sure that the Library is set
	if app.Library == "" {
		Err.Message = fmt.Sprintf("%s Library is not set", app.Type)
		Err.HelpText = "Please set the Library in the configuration file"
		Err.Details = map[string]any{
			"error": "Library is empty",
			"app":   app,
		}
		logging.LOG.ErrorWithLog(Err)
		return Err
	}

	// Make sure that the URL is set
	if app.URL == "" {
		Err.Message = fmt.Sprintf("%s URL is not set", app.Type)
		Err.HelpText = "Please set the URL in the configuration file"
		Err.Details = map[string]any{
			"error": "URL is empty",
			"app":   app,
		}
		logging.LOG.ErrorWithLog(Err)
		return Err
	}

	// Make sure that the API Key is set
	if app.APIKey == "" {
		Err.Message = fmt.Sprintf("%s API Key is not set", app.Type)
		Err.HelpText = "Please set the API Key in the configuration file"
		Err.Details = map[string]any{
			"error": "API Key is empty",
			"app":   app,
		}
		logging.LOG.ErrorWithLog(Err)
		return Err
	}
	return logging.StandardError{}
}
