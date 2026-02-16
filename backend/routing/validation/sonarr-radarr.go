package routes_validation

import (
	"aura/config"
	"aura/logging"
	sonarr_radarr "aura/sonarr-radarr"
	"aura/utils/httpx"
	"net/http"
)

func ValidateSonarrRadarrInfo(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Validate Sonarr/Radarr Info", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	// Get the Sonarr/Radarr information from the request
	var srInfo config.Config_SonarrRadarrApp
	Err := httpx.DecodeRequestBodyToJSON(ctx, r.Body, &srInfo, "Sonarr/Radarr Info")
	if Err.Message != "" {
		httpx.SendResponse(w, ld, nil)
		return
	}

	fullToken := srInfo.ApiToken
	if config.IsMaskedField(srInfo.ApiToken) {
		// If the token is masked, we need to grab it from the existing config
		for _, existingApp := range config.Current.SonarrRadarr.Applications {
			if existingApp.Type == srInfo.Type && existingApp.URL == srInfo.URL {
				fullToken = existingApp.ApiToken
				break
			}
		}
	}
	srInfo.ApiToken = fullToken

	_, Err = sonarr_radarr.TestConnection(ctx, srInfo)
	if Err.Message != "" {
		httpx.SendResponse(w, ld, nil)
		return
	}

	srInfo.ApiToken = config.MaskToken(srInfo.ApiToken)
	httpx.SendResponse(w, ld, srInfo)
}
