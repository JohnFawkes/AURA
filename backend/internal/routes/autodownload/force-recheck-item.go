package routes_autodownload

import (
	"aura/internal/api"
	"aura/internal/logging"
	"encoding/json"
	"net/http"
	"time"
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
	logging.LOG.Trace(r.URL.Path)
	startTime := time.Now()
	Err := logging.NewStandardError()

	var requestBody struct {
		Item api.DBMediaItemWithPosterSets
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		Err.Message = "Failed to decode request body"
		Err.HelpText = "Ensure the request body is a valid JSON object matching the expected structure."
		Err.Details = map[string]any{
			"error": err.Error(),
			"body":  r.Body,
		}
		logging.LOG.ErrorWithLog(Err)
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		return
	}
	item := requestBody.Item

	results := api.AutoDownload_CheckItem(item)

	// If no warnings, send a success response
	api.Util_Response_SendJson(w, http.StatusOK, api.JSONResponse{
		Status:  "success",
		Elapsed: api.Util_ElapsedTime(startTime),
		Data:    results,
	})
}
