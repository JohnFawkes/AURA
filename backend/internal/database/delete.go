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

	// Get the ratingKey from the URL
	ratingKey := chi.URLParam(r, "ratingKey")
	if ratingKey == "" {
		utils.SendErrorJSONResponse(w, http.StatusBadRequest, logging.ErrorLog{
			Err: fmt.Errorf("bad request"),
			Log: logging.Log{
				Message: "Missing ratingKey in request",
			},
		})
		return
	}

	logErr := deleteMediaItemFromDatabase(ratingKey)
	if logErr.Err != nil {
		utils.SendErrorJSONResponse(w, http.StatusInternalServerError, logErr)
		return
	}

	utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
		Status:  "success",
		Message: "Item deleted successfully",
		Elapsed: utils.ElapsedTime(startTime),
		Data:    "success"})
}

func deleteMediaItemFromDatabase(ratingKey string) logging.ErrorLog {
	deleteMediaItemQuery := `
DELETE FROM SavedItems WHERE media_item_id = ?`
	_, err := db.Exec(deleteMediaItemQuery, ratingKey)
	if err != nil {
		return logging.ErrorLog{Err: err, Log: logging.Log{
			Message: "Failed to delete Media Item from database",
		}}
	}

	logging.LOG.Info(fmt.Sprintf("DB - Media Item deleted successfully for item: %s", ratingKey))
	return logging.ErrorLog{}
}

func deletePosterItemFromDatabase(ratingKey string, setID string) logging.ErrorLog {
	deletePosterItemQuery := `
DELETE FROM SavedItems WHERE media_item_id = ? AND poster_set_id = ?`
	_, err := db.Exec(deletePosterItemQuery, ratingKey, setID)
	if err != nil {
		return logging.ErrorLog{Err: err, Log: logging.Log{
			Message: "Failed to delete Poster Item from database",
		}}
	}
	logging.LOG.Info(fmt.Sprintf("DB - Poster Item deleted successfully for item: %s - %s", ratingKey, setID))
	return logging.ErrorLog{}
}
