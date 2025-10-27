package api

import (
	"aura/internal/logging"
	"net/http"
)

func Mediux_ValidateToken(token string) logging.StandardError {
	// Create a new StandardError instance
	Err := logging.NewStandardError()
	logging.LOG.Debug("Validating MediUX token")
	if token == "" {
		Err.Message = "MediUX token is empty"
		Err.HelpText = "Please provide a valid MediUX token in the configuration file"
		return Err
	}

	// Make a GET request to the MediUX API to validate the token
	url := "https://images.mediux.io/users/me"
	response, body, err := MakeHTTPRequest(url, "GET", nil, 10, nil, "Mediux")
	if err.Message != "" {
		Err.Message = "Failed to validate MediUX token"
		Err.HelpText = "Ensure the MediUX service is reachable and the token is valid."
		return Err
	}

	if response.StatusCode != http.StatusOK {
		Err.Message = "Invalid MediUX token"
		Err.HelpText = "Ensure the MediUX token is correct and has not expired."
		Err.Details = map[string]any{
			"status_code": response.StatusCode,
			"response":    string(body),
		}
		return Err
	}

	logging.LOG.Info("Successfully validated MediUX token")
	return logging.StandardError{}
}
