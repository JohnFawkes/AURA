package routes_plex

import (
	"aura/internal/api"
	"aura/internal/logging"
	"net/http"
)

func PlexGetPinCodeAndIDHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Plex Get Pin Handler", logging.LevelDebug)
	ctx = logging.WithCurrentAction(ctx, logAction)

	// Get the values from Plex
	pinCode, plexID, logErr := api.Plex_GetPinCodeAndID(ctx)
	if logErr.Message != "" {
		logAction.SetError("Failed to get Plex pin", logErr.Message, logErr.Detail)
		api.Util_Response_SendJSON(w, ld, map[string]any{
			"error": logErr.Message,
		})
		return
	}

	api.Util_Response_SendJSON(w, ld, map[string]any{
		"pinCode": pinCode,
		"plexID":  plexID,
	})
}
