package routes_mediux

import (
	"aura/internal/api"
	"aura/internal/logging"
	"fmt"
	"net/http"
)

func CheckMediuxLink(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Check MediUX Link", logging.LevelTrace)
	ctx = logging.WithCurrentAction(ctx, logAction)

	// Parse query parameters
	itemType := r.URL.Query().Get("itemType")
	tmdbID := r.URL.Query().Get("tmdbID")

	// Validate input
	if itemType == "" || tmdbID == "" {
		logAction.SetError("Missing required query parameters", "Both 'itemType' and 'tmdbID' must be provided",
			map[string]any{
				"itemType": itemType,
				"tmdbID":   tmdbID,
			})
		api.Util_Response_SendJSON(w, ld, map[string]any{"exists": false, "url": ""})
		return
	}

	resp, err := http.Head(fmt.Sprintf("https://mediux.io/%s/%s", itemType, tmdbID))
	if err != nil {
		logAction.SetError("Failed to check MediUX link", "Error occurred while making HEAD request",
			map[string]any{
				"error": err.Error(),
				"url":   fmt.Sprintf("https://mediux.io/%s/%s", itemType, tmdbID),
			})
		api.Util_Response_SendJSON(w, ld, map[string]any{"exists": false, "url": ""})
		return
	}
	defer resp.Body.Close()

	exists := resp.StatusCode == http.StatusOK
	var urlString string
	if exists {
		urlString = fmt.Sprintf("https://mediux.io/%s/%s", itemType, tmdbID)
	} else {
		urlString = fmt.Sprintf("https://mediux.pro/%ss/%s", itemType, tmdbID)
	}

	api.Util_Response_SendJSON(w, ld, map[string]any{
		"exists": exists,
		"url":    urlString,
	})
}
