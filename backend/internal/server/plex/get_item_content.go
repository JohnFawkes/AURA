package plex

import (
	"fmt"
	"net/http"
	"poster-setter/internal/logging"
	"poster-setter/internal/utils"
	"time"

	"github.com/go-chi/chi/v5"
)

func GetItemContent(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	logging.LOG.Trace(r.URL.Path)

	// Get the SKU from the URL
	ratingKey := chi.URLParam(r, "ratingKey")
	if ratingKey == "" {
		utils.SendErrorJSONResponse(w, http.StatusInternalServerError, logging.ErrorLog{Err: fmt.Errorf("missing rating key"),
			Log: logging.Log{
				Message: "Missing rating key in URL",
				Elapsed: utils.ElapsedTime(startTime),
			},
		})
		return
	}

	itemInfo, logErr := FetchItemContent(ratingKey)
	if logErr.Err != nil {
		utils.SendErrorJSONResponse(w, http.StatusInternalServerError, logErr)
		return
	}

	if itemInfo.RatingKey == "" {
		utils.SendErrorJSONResponse(w, http.StatusInternalServerError, logging.ErrorLog{Err: fmt.Errorf("no item found with the given rating key"),
			Log: logging.Log{
				Message: "No item found with the given rating key",
				Elapsed: utils.ElapsedTime(startTime),
			},
		})
		return
	}

	// Respond with a success message
	utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
		Status:  "success",
		Message: "Retrieved item content from Plex",
		Elapsed: utils.ElapsedTime(startTime),
		Data:    itemInfo,
	})
}
