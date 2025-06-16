package plex

import (
	"aura/internal/logging"
	"aura/internal/utils"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

func GetItemContent(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	logging.LOG.Trace(r.URL.Path)
	Err := logging.NewStandardError()

	// Get the SKU from the URL
	ratingKey := chi.URLParam(r, "ratingKey")
	if ratingKey == "" {

		Err.Message = "Missing rating key"
		Err.HelpText = "Ensure the URL contains a valid rating key parameter."
		Err.Details = fmt.Sprintf("Received ratingKey: %s", ratingKey)
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}

	itemInfo, Err := FetchItemContent(ratingKey)
	if Err.Message != "" {
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}

	if itemInfo.RatingKey == "" {

		Err.Message = "Item not found"
		Err.HelpText = "Ensure the rating key corresponds to an existing item in Plex."
		Err.Details = fmt.Sprintf("No item found for ratingKey: %s", ratingKey)
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}

	// Respond with a success message
	utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
		Status:  "success",
		Elapsed: utils.ElapsedTime(startTime),
		Data:    itemInfo,
	})
}
