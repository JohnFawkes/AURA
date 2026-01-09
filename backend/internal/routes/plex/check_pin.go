package routes_plex

import (
	"aura/internal/api"
	"aura/internal/logging"
	"net/http"
)

func PlexCheckPinHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Plex Check Pin Handler", logging.LevelDebug)
	ctx = logging.WithCurrentAction(ctx, logAction)

	// Get the PIN from the request body
	plexID := r.URL.Query().Get("plexID")
	if plexID == "" {
		logAction.SetError("Missing PIN code", "No PIN code provided in request", map[string]any{
			"plexID": plexID,
		})
		api.Util_Response_SendJSON(w, ld, map[string]any{
			"error": "Missing PIN code",
		})
		return
	}

	// Check the PIN with Plex
	isValid, authToken, logErr := api.Plex_CheckPin(ctx, plexID)
	if logErr.Message != "" {
		logAction.SetError("Failed to check Plex pin", logErr.Message, logErr.Detail)
		api.Util_Response_SendJSON(w, ld, map[string]any{
			"error": logErr.Message,
		})
		return
	}

	var listOfPlexServerConnections []api.PlexServersResponse
	// If the PIN is valid and we have an auth token, then get a list of connections available
	if isValid && authToken != "" {
		listOfPlexServerConnections, logErr = api.Plex_GetServerConnections(ctx, authToken)
		if logErr.Message != "" {
			logAction.SetError("Failed to get Plex server connections", logErr.Message, logErr.Detail)
			api.Util_Response_SendJSON(w, ld, map[string]any{
				"error": logErr.Message,
			})
			return
		}
	}

	api.Util_Response_SendJSON(w, ld, map[string]any{
		"authenticated":        isValid,
		"authToken":            authToken,
		"connectionsAvailable": listOfPlexServerConnections,
	})
}
