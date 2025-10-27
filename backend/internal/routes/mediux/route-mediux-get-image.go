package routes_mediux

import (
	"aura/internal/api"
	"aura/internal/logging"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

/*
Route_Mediux_GetImage handles the API request to fetch an image from Mediux based on asset ID, modified date, and quality.

Takes in the following:

- assetID as a URL parameter

- modifiedDate as a query parameter (ISO 8601 format, e.g., 2023-10-01T12:00:00Z). If not provided, defaults to today's date.

- quality as a query parameter (either "thumb", "original", or "optimized"). Defaults to "thumb" if not provided.

Returns the image data with the appropriate Content-Type header.
*/
func GetImage(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	logging.LOG.Trace(r.URL.Path)
	Err := logging.NewStandardError()

	// Get the asset ID from the URL
	assetID := chi.URLParam(r, "assetID")
	if assetID == "" {
		Err.Message = "Missing asset ID in URL"
		Err.HelpText = "Ensure the asset ID is provided in the URL path."
		Err.Details = map[string]any{
			"error":   "Asset ID is empty",
			"request": r.URL.Path,
		}
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		return
	}

	// Get the modified date from the URL query parameters
	modifiedDate := r.URL.Query().Get("modifiedDate")
	var modifiedDateTime time.Time
	var err error
	if modifiedDate == "" || modifiedDate == "0" || modifiedDate == "undefined" {
		// Use today's date if the modified date is not provided
		modifiedDateTime = time.Now()
	} else {
		// Try to parse the modified date as an ISO 8601 timestamp
		modifiedDateTime, err = time.Parse(time.RFC3339, modifiedDate)
		if err != nil {
			Err.Message = "Invalid modified date format"
			Err.HelpText = "Ensure the modified date is in ISO 8601 format (e.g., 2023-10-01T12:00:00Z)."
			Err.Details = map[string]any{
				"modifiedDate": modifiedDate,
			}
			api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
			return
		}
	}
	// Format the date to be YYYYMMDDHHMMSS
	// Example: 2025-06-20T10:20:30Z -> 20250620102030
	formatDate := modifiedDateTime.Format("20060102150405")

	// Get Quality from the URL query parameters
	qualityParam := r.URL.Query().Get("quality")
	if qualityParam == "" {
		// Default to "thumb" if quality is not provided
		qualityParam = "thumb"
	}
	// Check if the quality is valid
	if qualityParam != "thumb" && qualityParam != "original" && qualityParam != "optimized" {
		Err.Message = "Invalid quality parameter"
		Err.HelpText = "Ensure the quality parameter is either 'thumb', 'original', or 'optimized'."
		Err.Details = map[string]any{
			"quality": qualityParam,
		}
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		return
	}

	// If the image does not exist, then get it from Mediux
	imageData, imageType, Err := api.Mediux_GetImage(assetID, formatDate, qualityParam)
	if Err.Message != "" {
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		return
	}

	w.Header().Set("Content-Type", imageType)
	w.WriteHeader(http.StatusOK)
	w.Write(imageData)
}
