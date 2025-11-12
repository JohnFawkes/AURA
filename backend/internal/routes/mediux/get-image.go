package routes_mediux

import (
	"aura/internal/api"
	"aura/internal/logging"
	"net/http"
	"time"
)

func GetImage(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Get Image From MediUX", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	actionGetQueryParams := logAction.AddSubAction("Get all query params", logging.LevelTrace)
	// Get the following information from the URL
	// Asset ID
	// Modified Date
	// Quality
	assetID := r.URL.Query().Get("assetID")
	modifiedDate := r.URL.Query().Get("modifiedDate")
	quality := r.URL.Query().Get("quality")

	if quality == "" {
		quality = "thumb"
	}
	if quality != "thumb" && quality != "original" && quality != "optimized" {
		actionGetQueryParams.SetError("Invalid Query Parameters", "The quality parameter provided is not valid. Must be one of: thumb, original, optimized",
			map[string]any{
				"assetID":      assetID,
				"modifiedDate": modifiedDate,
				"quality":      quality,
			})
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}
	var imageQuality api.MediuxImageQuality
	switch quality {
	case "thumb":
		imageQuality = api.MediuxImageQualityThumb
	case "optimized":
		imageQuality = api.MediuxImageQualityOptimized
	case "original":
		imageQuality = api.MediuxImageQualityOriginal
	}

	// Get the modified date from the URL query parameters
	var modifiedDateTime time.Time
	var err error
	if modifiedDate == "" || modifiedDate == "0" || modifiedDate == "undefined" {
		// Use today's date if the modified date is not provided
		modifiedDateTime = time.Now()
	} else {
		// Try to parse the modified date as an ISO 8601 timestamp
		modifiedDateTime, err = time.Parse(time.RFC3339, modifiedDate)
		if err != nil {
			actionGetQueryParams.SetError("Invalid Query Parameters", "The modified date provided is not a valid ISO 8601 timestamp",
				map[string]any{
					"assetID":      assetID,
					"modifiedDate": modifiedDate,
					"quality":      quality,
				})
			api.Util_Response_SendJSON(w, ld, nil)
			return
		}
	}
	// Format the date to be YYYYMMDDHHMMSS
	// Example: 2025-06-20T10:20:30Z -> 20250620102030
	formatDate := modifiedDateTime.Format("20060102150405")

	// Validate the asset ID and modified date
	if assetID == "" {
		actionGetQueryParams.SetError("Missing Query Parameters", "One or more required query parameters are missing",
			map[string]any{
				"assetID":      assetID,
				"modifiedDate": modifiedDate,
			})
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}
	actionGetQueryParams.Complete()

	imageData, imageType, Err := api.Mediux_GetImage(ctx, assetID, formatDate, imageQuality)
	if Err.Message != "" {
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}

	w.Header().Set("Content-Type", imageType)
	w.WriteHeader(http.StatusOK)
	w.Write(imageData)
}

func GetAvatarImage(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Get Avatar Image From MediUX", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	actionGetQueryParams := logAction.AddSubAction("Get all query params", logging.LevelTrace)
	// Get the following information from the URL
	// Avatar ID
	avatarID := r.URL.Query().Get("avatarID")
	if avatarID == "" {
		// If the avatarID is missing, check to see if there is a username parameter
		username := r.URL.Query().Get("username")
		if username == "" {
			actionGetQueryParams.SetError("Missing Query Parameters", "One or more required query parameters are missing",
				map[string]any{
					"avatarID": avatarID,
					"username": username,
				})
			api.Util_Response_SendJSON(w, ld, nil)
			return
		}

		// Get the avatar ID from the username
		// TODO: Implement this
		return
	}
	actionGetQueryParams.Complete()

	imageData, imageType, Err := api.Mediux_GetAvatarImage(ctx, avatarID)
	if Err.Message != "" {
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}

	w.Header().Set("Content-Type", imageType)
	w.WriteHeader(http.StatusOK)
	w.Write(imageData)
}
