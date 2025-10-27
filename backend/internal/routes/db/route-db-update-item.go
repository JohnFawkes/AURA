package routes_db

import (
	"aura/internal/api"
	"aura/internal/logging"
	"encoding/json"
	"net/http"
	"time"
)

func UpdateItem(w http.ResponseWriter, r *http.Request) {
	logging.LOG.Debug(r.URL.Path)
	startTime := time.Now()

	// Get the request body
	var saveItem api.DBMediaItemWithPosterSets

	// Decode the request body into the struct
	err := json.NewDecoder(r.Body).Decode(&saveItem)
	if err != nil {
		Err := logging.NewStandardError()
		Err.Message = "Failed to decode request body"
		Err.HelpText = "Ensure the request body is a valid JSON object matching the expected structure."
		Err.Details = map[string]any{
			"error": err.Error(),
			"body":  r.Body,
		}
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		return
	}

	for _, posterSet := range saveItem.PosterSets {
		if posterSet.ToDelete {
			Err := api.DB_Delete_PosterSet(posterSet.PosterSetID, saveItem.MediaItem.TMDB_ID, saveItem.MediaItem.LibraryTitle)
			if Err.Message != "" {
				api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
				return
			}
		} else {
			Err := api.DB_InsertAllInfoIntoTables(saveItem)
			if Err.Message != "" {
				api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
				return
			}
		}
	}

	api.Util_Response_SendJson(w, http.StatusOK, api.JSONResponse{
		Status:  "success",
		Elapsed: api.Util_ElapsedTime(startTime),
		Data:    "success"})
}
