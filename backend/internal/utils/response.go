package utils

import (
	"aura/internal/logging"
	"encoding/json"
	"net/http"
)

type JSONResponse struct {
	Status  string                 `json:"status"`          // success, error, or warning
	Elapsed string                 `json:"elapsed"`         // Time taken to process request
	Data    any                    `json:"data,omitempty"`  // Response data (if any)
	Error   *logging.StandardError `json:"error,omitempty"` // Error details (if any)
}

func SendErrorResponse(w http.ResponseWriter, elapsed string, err logging.StandardError) {
	if err.Function == "" {
		err.Function = GetFunctionName()
	}
	if err.LineNumber == 0 {
		err.LineNumber = GetLineNumber()
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	response := JSONResponse{
		Status:  "error",
		Elapsed: elapsed,
		Error:   &err,
	}
	// Log the error using the logging package
	logging.LOG.Error(err.Message)
	// Encode the response as JSON and send it
	if err := json.NewEncoder(w).Encode(response); err != nil {
		// Handle encoding error if necessary
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func SendJsonResponse(w http.ResponseWriter, statusCode int, response JSONResponse) {

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Encode the data into JSON and send it
	if err := json.NewEncoder(w).Encode(response); err != nil {
		// Handle encoding error if necessary
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
