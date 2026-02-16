package routes_ms

import (
	"aura/config"
	"aura/logging"
	"aura/mediaserver"
	"aura/utils/httpx"
	"net/http"
)

func GetLibrarySectionOptions(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Get Library Section Options", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	// Get the reqeuest body
	var msConfig config.Config_MediaServer
	Err := httpx.DecodeRequestBodyToJSON(ctx, r.Body, &msConfig, "Get Library Section Options - Decode Request Body")
	if Err.Message != "" {
		httpx.SendResponse(w, ld, nil)
		return
	}

	// Get all available library sections from the Media Server
	librarySections, Err := mediaserver.GetLibrarySections(ctx, &msConfig)
	if Err.Message != "" {
		httpx.SendResponse(w, ld, nil)
		return
	}

	if len(librarySections) == 0 {
		logAction.SetError("No library sections found", "No library sections could be retrieved from the Media Server", nil)
		httpx.SendResponse(w, ld, nil)
		return
	}

	httpx.SendResponse(w, ld, librarySections)
}
