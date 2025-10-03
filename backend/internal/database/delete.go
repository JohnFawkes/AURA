package database

import (
	"aura/internal/logging"
	"aura/internal/utils"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

func DeleteMediaItemFromDatabase(w http.ResponseWriter, r *http.Request) {
	logging.LOG.Debug(r.URL.Path)
	startTime := time.Now()
	Err := logging.NewStandardError()

	// Get the ratingKey from the URL
	ratingKey := chi.URLParam(r, "ratingKey")
	if ratingKey == "" {
		Err.Message = "Missing Rating Key in Request"
		Err.HelpText = "Ensure the request includes a valid ratingKey parameter."
		Err.Details = map[string]any{
			"error":     "Rating Key is empty",
			"ratingKey": ratingKey,
			"request":   r.URL.Path,
		}
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}

	Err = deleteMediaItemFromDatabase(ratingKey)
	if Err.Message != "" {
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}

	utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
		Status:  "success",
		Elapsed: utils.ElapsedTime(startTime),
		Data:    "success"})
}

func deleteMediaItemFromDatabase(ratingKey string) logging.StandardError {
	deleteMediaItemQuery := `
DELETE FROM SavedItems WHERE media_item_id = ?`
	_, err := db.Exec(deleteMediaItemQuery, ratingKey)
	if err != nil {
		Err := logging.NewStandardError()
		Err.Message = "Failed to delete Media Item from database"
		Err.HelpText = "Ensure the database connection is established and the query is correct."
		Err.Details = map[string]any{
			"error":     err.Error(),
			"query":     deleteMediaItemQuery,
			"ratingKey": ratingKey,
		}
		return Err
	}

	logging.LOG.Info(fmt.Sprintf("DB - Media Item deleted successfully for item: %s", ratingKey))
	return logging.StandardError{}
}

func deletePosterItemFromDatabase(ratingKey string, setID string) logging.StandardError {
	deletePosterItemQuery := `
DELETE FROM SavedItems WHERE media_item_id = ? AND poster_set_id = ?`
	_, err := db.Exec(deletePosterItemQuery, ratingKey, setID)
	if err != nil {
		Err := logging.NewStandardError()
		Err.Message = "Failed to delete Poster Item from database"
		Err.HelpText = "Ensure the database connection is established and the query is correct."
		Err.Details = map[string]any{
			"error":     err.Error(),
			"query":     deletePosterItemQuery,
			"ratingKey": ratingKey,
			"setID":     setID,
		}
		return Err
	}
	logging.LOG.Info(fmt.Sprintf("DB - Poster Item deleted successfully for item: %s - %s", ratingKey, setID))
	return logging.StandardError{}
}
