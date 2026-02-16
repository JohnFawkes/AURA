package routes_plex

import (
	"aura/logging"
	"aura/mediaserver/plex"
	"aura/utils/httpx"
	"net/http"
)

func GetPlexPinAndID(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Get Plex Pin and ID", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	// Get Pin Code and ID from Plex API

	plexPin, plexID, plexErr := plex.OAuth_GetPinAndID(ctx)
	if plexErr.Message != "" {
		httpx.SendResponse(w, ld, plexErr)
		return
	}

	var response struct {
		PlexPin string `json:"plex_pin"`
		PlexID  int64  `json:"plex_id"`
	}
	response.PlexPin = plexPin
	response.PlexID = plexID

	httpx.SendResponse(w, ld, response)
}
