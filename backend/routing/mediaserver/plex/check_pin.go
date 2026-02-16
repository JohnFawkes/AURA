package routes_plex

import (
	"aura/logging"
	"aura/mediaserver/plex"
	"aura/utils/httpx"
	"net/http"
)

func CheckPlexPin(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Check Plex ID For Auth", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	// Get the Plex ID from the request
	plexID := r.URL.Query().Get("plex_id")
	if plexID == "" {
		logAction.SetError("Plex ID is required", "No Plex ID provided in request", nil)
		httpx.SendResponse(w, ld, nil)
		return
	}

	// Check the Plex ID to see if the user has authenticated with Plex
	isValid, authToken, plexErr := plex.OAuth_CheckIDForAuth(ctx, plexID)
	if plexErr.Message != "" {
		httpx.SendResponse(w, ld, nil)
		return
	}

	// If the Plex ID is valid, get a list of available connections
	var listOfPlexServerConnections []plex.PlexServersResponse
	if isValid && authToken != "" {
		listOfPlexServerConnections, plexErr = plex.OAuth_GetServerConnections(ctx, authToken)
		if plexErr.Message != "" {
			httpx.SendResponse(w, ld, nil)
			return
		}
	}

	var response struct {
		Authenticated        bool                       `json:"authenticated"`
		AuthToken            string                     `json:"auth_token"`
		ConnectionsAvailable []plex.PlexServersResponse `json:"connections_available"`
	}
	response.Authenticated = isValid
	response.AuthToken = authToken
	response.ConnectionsAvailable = listOfPlexServerConnections

	httpx.SendResponse(w, ld, response)
}
