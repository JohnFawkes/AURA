package api

import (
	"aura/internal/logging"
)

type JSONResponse struct {
	Status  string                 `json:"status"`          // success, error, or warning
	Elapsed string                 `json:"elapsed"`         // Time taken to process request
	Data    any                    `json:"data,omitempty"`  // Response data (if any)
	Error   *logging.StandardError `json:"error,omitempty"` // Error details (if any)
}
