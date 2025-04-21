package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"poster-setter/internal/logging"
	"strings"
)

type JSONResponse struct {
	Status  string `json:"status,omitempty"`
	Message string `json:"message,omitempty"`
	Elapsed string `json:"elapsed,omitempty"`
	Data    any    `json:"data,omitempty"`
}

type SSEMessage struct {
	Response JSONResponse `json:"response,omitempty"`
	Progress SSEProgress  `json:"progress,omitempty"`
}

type SSEProgress struct {
	Value    int    `json:"value,omitempty"`
	Text     string `json:"text,omitempty"`
	NextStep string `json:"nextStep,omitempty"`
}

func SendErrorJSONResponse(w http.ResponseWriter, statusCode int, err logging.ErrorLog) {
	SendJsonResponse(w, statusCode, JSONResponse{
		Status:  "error",
		Message: fmt.Sprintf("%s - %s", err.Log.Message, err.Err.Error()),
		Elapsed: err.Log.Elapsed,
		Data:    nil,
	})
}

func SendJsonResponse(w http.ResponseWriter, statusCode int, response JSONResponse) {
	// Set the Content-Type header to application/json
	w.Header().Set("Content-Type", "application/json")

	// Set the status code
	w.WriteHeader(statusCode)

	// Set the status field based on the status code if it's not already set
	if response.Status == "" {
		if statusCode == http.StatusOK {
			response.Status = "success"
		} else {
			response.Status = "error"
			if response.Message == "" {
				response.Message = http.StatusText(statusCode)
			}

		}
	}

	// If there is an error, log it
	if response.Status == "error" {
		logging.LOG.Error(response.Message)
	}

	// Make the first letter of the message uppercase
	if response.Message != "" {
		response.Message = strings.ToUpper(string(response.Message[0])) + response.Message[1:]
	}

	// Encode the data into JSON and send it
	if err := json.NewEncoder(w).Encode(response); err != nil {
		// Handle encoding error if necessary
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

}

func SendSSEResponse(w http.ResponseWriter, flusher http.Flusher, message SSEMessage) {
	jsonData, err := json.Marshal(message)
	if err != nil {
		logging.LOG.Error("Failed to marshal SSE message")
		return
	}
	fmt.Fprintf(w, "data: %s\n\n", jsonData)
	flusher.Flush()
}
