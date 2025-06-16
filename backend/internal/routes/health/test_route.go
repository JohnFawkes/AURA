package health

import (
	"aura/internal/logging"
	"aura/internal/utils"
	"net/http"
	"time"
)

func TestRouteError(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Create a new StandardError with details about the error
	Err := logging.NewStandardError()

	Err.Message = "This is a test error for the TestRouteError function"
	Err.HelpText = "This error is intentionally triggered to test error handling in the TestRouteError function."
	Err.Details = "Something went wrong while processing the request."

	utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)

}

func TestRoute(w http.ResponseWriter, r *http.Request) {

	startTime := time.Now()

	utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
		Status:  "success",
		Elapsed: utils.ElapsedTime(startTime),
		Data:    "good",
	})

}
