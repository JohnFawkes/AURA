package logging

import (
	"sync"
	"time"
)

const (
	StatusSuccess = "success"
	StatusError   = "error"
	StatusWarn    = "warn"

	LevelTrace = "trace"
	LevelDebug = "debug"
	LevelInfo  = "info"
	LevelWarn  = "warn"
	LevelError = "error"
)

// LogData is the base element used for both HTTP routes and background jobs.
type LogData struct {
	Status              string        `json:"status"`          // Status: "success", "error", or "warn"
	Name                string        `json:"name"`            // Human name for the overall operation: "Add to DB", "Fetch All Media Items" or Route.Path
	Timestamp           time.Time     `json:"timestamp"`       // When the operation started
	ElapsedMicroseconds int64         `json:"elapsed_us"`      // Elapsed time in microseconds
	Route               *LogRouteInfo `json:"route,omitempty"` // Optional route metadata (nil for background jobs)
	Actions             []*LogAction  `json:"actions"`         // Slice of actions within this operation
	mu                  sync.Mutex    // Mutex to protect concurrent access to Actions
}

// RouteInfo holds request-specific metadata populated by the Chi middleware.
type LogRouteInfo struct {
	Method        string              `json:"method"`           // HTTP method (e.g. GET, POST)
	Path          string              `json:"path"`             // Registered route path (e.g. /api/v1/media/{id})
	Params        map[string][]string `json:"params,omitempty"` // Query/path/route params: map[string][]string to preserve multiple values
	IP            string              `json:"ip,omitempty"`     // Client IP address (middleware can capture)
	ResponseBytes int64               `json:"response_bytes"`   // Number of bytes written in the response (middleware can capture)
}

// LogAction is a single step inside an operation.
type LogAction struct {
	Name                string         `json:"name"`                  // Human name for the action: "Query DB", "Call External API"
	Status              string         `json:"status"`                // "success", "error", or "warn"
	Level               string         `json:"level,omitempty"`       // Optional level for UI filtering (e.g. "trace", "debug","info","warn","error")
	Warnings            map[string]any `json:"warnings,omitempty"`    // Warning message (if any)
	Error               *LogErrorInfo  `json:"error,omitempty"`       // Optional error details (omitted when nil)
	Timestamp           time.Time      `json:"timestamp"`             // When the action started (useful for ordering)
	ElapsedMicroseconds int64          `json:"elapsed_us"`            // Elapsed time for the action in microseconds
	Result              map[string]any `json:"result,omitempty"`      // Optional result data
	SubActions          []*LogAction   `json:"sub_actions,omitempty"` // Optional nested sub-actions
	mu                  sync.Mutex     // Mutex to protect concurrent access to SubActions
}

// LogErrorInfo contains structured error information.
type LogErrorInfo struct {
	Function   string         `json:"function,omitempty"`    // Function name where the error occurred
	LineNumber int            `json:"line_number,omitempty"` // Line number in the source code
	Message    string         `json:"message"`               // Error message
	Detail     map[string]any `json:"detail,omitempty"`      // Optional detailed error information
	Help       string         `json:"help,omitempty"`        // Optional help or remediation steps
}
