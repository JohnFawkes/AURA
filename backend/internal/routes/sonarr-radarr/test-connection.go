package routes_sonarr_radarr

import (
	"aura/internal/api"
	"aura/internal/logging"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

func TestConnection(w http.ResponseWriter, r *http.Request) {
	logging.LOG.Trace(r.URL.Path)
	startTime := time.Now()
	Err := logging.NewStandardError()

	// Get the Sonarr/Radarr information from the request
	var srInfo api.Config_SonarrRadarrApp
	if err := json.NewDecoder(r.Body).Decode(&srInfo); err != nil {
		Err.Message = "Failed to decode request body"
		Err.HelpText = "Ensure the request body is valid JSON"
		Err.Details = map[string]any{
			"error": err.Error(),
		}
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
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

	_, Err = api.SR_CallTestConnection(srInfo)
	if Err.Message != "" {
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		return
	}

	// Respond with a success message
	api.Util_Response_SendJson(w, http.StatusOK, api.JSONResponse{
		Status:  "success",
		Elapsed: api.Util_ElapsedTime(startTime),
		Data:    srInfo,
	})
}
