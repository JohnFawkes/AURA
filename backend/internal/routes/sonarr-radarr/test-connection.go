package routes_sonarr_radarr

import (
	"aura/internal/api"
	"aura/internal/logging"
	"aura/internal/masking"
	"net/http"
	"strings"
)

func TestConnection(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Test Sonarr/Radarr Connection", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	// Get the Sonarr/Radarr information from the request
	var srInfo api.Config_SonarrRadarrApp
	Err := api.DecodeRequestBodyJSON(ctx, r.Body, &srInfo, "Config Sonarr/Radarr Info")
	if Err.Message != "" {
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}

	fullToken := srInfo.APIKey
	if strings.HasPrefix(srInfo.APIKey, "***") {
		// If the API key is masked, get the actual key from the global config
		for _, existingApp := range api.Global_Config.SonarrRadarr.Applications {
			if existingApp.Type == srInfo.Type && existingApp.URL == srInfo.URL {
				fullToken = existingApp.APIKey
				break
			}
		}
	}
	srInfo.APIKey = fullToken

	_, Err = api.SR_CallTestConnection(ctx, srInfo)
	if Err.Message != "" {
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}

	srInfo.APIKey = masking.Masking_Token(srInfo.APIKey)
	api.Util_Response_SendJSON(w, ld, srInfo)
}
