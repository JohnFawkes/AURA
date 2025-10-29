package api

import (
	"aura/internal/logging"
	"encoding/json"
	"net/http"
)

type JSONResponse struct {
	Status string                `json:"status"` // "success", "error" or "warn"
	Data   any                   `json:"data,omitempty"`
	Error  *logging.LogErrorInfo `json:"error,omitempty"`
}

func Util_Response_SendJSON(w http.ResponseWriter, log *logging.LogData, data any) {
	var response JSONResponse
	// Always check for error actions
	for _, action := range log.Actions {
		action.Complete()
		if action.Status == logging.StatusError && action.Error != nil {
			response.Error = action.Error
			if log.Status == "" {
				log.Status = logging.StatusError
			}
			break
		}
	}

	if log.Status == "" {
		log.Status = logging.StatusSuccess
	}
	response.Status = log.Status
	response.Data = data

	if log.Status == logging.StatusError {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
