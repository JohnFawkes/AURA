package database

import (
	"aura/internal/logging"
	"aura/internal/utils"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

func DeleteItemFromDatabase(w http.ResponseWriter, r *http.Request) {
	logging.LOG.Debug(r.URL.Path)
	startTime := time.Now()

	// Get the ratingKey from the URL
	ratingKey := chi.URLParam(r, "ratingKey")
	if ratingKey == "" {
		utils.SendErrorJSONResponse(w, http.StatusBadRequest, logging.ErrorLog{
			Err: fmt.Errorf("missing ratingKey in request"),
			Log: logging.Log{
				Message: "Missing ratingKey in request",
			},
		})
		return
	}

	logErr := DeleteMediaItemFromDatabase(ratingKey)
	if logErr.Err != nil {
		utils.SendErrorJSONResponse(w, http.StatusInternalServerError, logErr)
		return
	}

	utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
		Status:  "success",
		Message: "Item deleted successfully",
		Elapsed: utils.ElapsedTime(startTime),
		Data:    nil})
}

func DeleteMediaItemFromDatabase(ratingKey string) logging.ErrorLog {
	deleteMediaItemQuery := `
DELETE FROM Media_Item WHERE id = ?`
	_, err := db.Exec(deleteMediaItemQuery, ratingKey)
	if err != nil {
		return logging.ErrorLog{Err: err, Log: logging.Log{
			Message: "Failed to delete data from database",
		}}
	}

	logErr := DeletePosterSetFromDatabaseByMediaID(ratingKey)
	if logErr.Err != nil {
		return logErr
	}
	logging.LOG.Info(fmt.Sprintf("DB - MediaItem deleted successfully for item: %s", ratingKey))
	return logging.ErrorLog{}
}

func DeletePosterSetFromDatabaseByMediaID(ratingKey string) logging.ErrorLog {
	deletePosterSetQuery := `
DELETE FROM Poster_Sets WHERE media_item_id = ?`
	_, err := db.Exec(deletePosterSetQuery, ratingKey)
	if err != nil {
		return logging.ErrorLog{Err: err, Log: logging.Log{
			Message: "Failed to delete data from database",
		}}
	}
	logging.LOG.Info(fmt.Sprintf("DB - PosterSet deleted successfully for item: %s", ratingKey))
	return logging.ErrorLog{}
}

func DeletePosterSetFromDatabaseByID(posterSetID string) logging.ErrorLog {
	deletePosterSetQuery := `
DELETE FROM Poster_Sets WHERE id = ?`
	_, err := db.Exec(deletePosterSetQuery, posterSetID)
	if err != nil {
		return logging.ErrorLog{Err: err, Log: logging.Log{
			Message: "Failed to delete data from database",
		}}
	}
	logging.LOG.Info(fmt.Sprintf("DB - PosterSet deleted successfully for item: %s", posterSetID))
	return logging.ErrorLog{}
}
