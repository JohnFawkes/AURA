package routes_autodownload

import (
	"aura/internal/api"
	"aura/internal/logging"
	"net/http"
)

// ForceRecheckItem handles requests to force a recheck of a specific media item.
//
// Method: POST
//
// Endpoint: /api/db/force-recheck
//
// It expects a JSON body containing the media item details.
//
// If the recheck is successful, it responds with a success message and the results.
//
// If there is an error, it responds with a JSON error message.
func ForceRecheckItem(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Force Recheck Item", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	var requestBody struct {
		Item          api.DBMediaItemWithPosterSets
		LargeDataSize bool
	}
	Err := api.DecodeRequestBodyJSON(ctx, r.Body, &requestBody, "DBMediaItemWithPosterSets")
	if Err.Message != "" {
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}

	item := requestBody.Item
	largeDataSize := requestBody.LargeDataSize
	logAction.AppendResult("large", largeDataSize)

	results := api.AutoDownload_CheckItem(ctx, item)

	api.Util_Response_SendJSON(w, ld, results)
}
