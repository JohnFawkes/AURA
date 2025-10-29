package routes_ms

import (
	"aura/internal/api"
	"aura/internal/logging"
	"net/http"
)

func GetAllSections(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Get All Sections", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	// Fetch all sections from the media server
	allSections, Err := api.CallFetchLibrarySectionInfo(ctx)
	if Err.Message != "" {
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}

	if len(allSections) == 0 {
		logAction.SetError("No library sections found",
			"Ensure that the Media Server has library sections configured",
			nil)
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}

	api.Util_Response_SendJSON(w, ld, allSections)
}
