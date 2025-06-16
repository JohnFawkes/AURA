package logging

import (
	"net/http"
)

// HTTPRequestLog represents the structure for logging HTTP request details.
// It includes information such as the time of the request, HTTP method used,
// response status, request path, size of the response in bytes, elapsed time
// for the request, and the IP address of the client.
type HTTPRequestLog struct {
	Time    string
	Method  string
	Status  string
	Path    string
	Bytes   string
	Elapsed string
	IP      string
}

// Log represents a log entry with a message and an elapsed time.
// Message contains the log message as a string.
// Elapsed indicates the time duration associated with the log entry as a string.
type Log struct {
	Message string
	Elapsed string
}

// ErrorLog represents a structured log entry that contains an error and its associated log details.
// It includes an error instance and a Log object to provide additional context for the error.
type ErrorLog struct {
	Err error
	Log Log // Log details associated with the error
}

type LogFormatter struct{}

type LogEntry struct {
	Request *http.Request
}

// StandardError represents error details for API responses
type StandardError struct {
	Message    string `json:"Message"`    // User-facing error message
	HelpText   string `json:"HelpText"`   // Helpful suggestion for the user
	Details    any    `json:"Details"`    // Additional details about the error
	Function   string `json:"Function"`   // Function where error occurred
	LineNumber int    `json:"LineNumber"` // Line number where error occurred
}

// Create a blank StandardError
func NewStandardError() StandardError {
	return StandardError{
		Message:    "",
		HelpText:   "",
		Function:   "",
		LineNumber: 0,
	}
}
