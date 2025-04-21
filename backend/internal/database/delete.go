package database

import (
	"fmt"
	"net/http"
	"poster-setter/internal/logging"
	"poster-setter/internal/utils"
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

	logErr := DeleteFromDatabase(ratingKey)
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

func DeleteFromDatabase(ratingKey string) logging.ErrorLog {
	query := `
DELETE FROM auto_downloader WHERE id = ?`
	_, err := db.Exec(query, ratingKey)
	if err != nil {
		return logging.ErrorLog{Err: err, Log: logging.Log{
			Message: "Failed to delete data from database",
		}}
	}
	return logging.ErrorLog{}
}
