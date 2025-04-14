package logging

import (
	"net/http"
	"path"
	"regexp"
	"time"
)

// shouldSkipLogging determines whether logging should be skipped for a given HTTP request.
// It checks the request's URL path against a predefined list of patterns for paths to exclude from logging,
// such as static assets (e.g., .js, .css, .svg) and specific API endpoints.
// Additionally, it skips logging for HTTP OPTIONS requests.
//
// Parameters:
//   - request: A pointer to an http.Request object representing the incoming HTTP request.
//
// Returns:
//   - A boolean value: true if logging should be skipped, false otherwise.
func shouldSkipLogging(request *http.Request) bool {
	// Define patterns for paths to skip logging
	skipPatterns := []string{
		`\.js$`,
		`\.css$`,
		`manifest\.json$`,
		`\.svg$`,
		`logo.*$`,
		`^/assets/.*$`,
		`^/favicon.ico$`,
		`^/api/plex/image/.*$`,
	}

	// Check if the path matches any skip pattern
	for _, pattern := range skipPatterns {
		matched, err := regexp.MatchString(pattern, request.URL.Path)
		if err == nil && matched {
			return true
		}
	}

	return request.Method == "OPTIONS"
}

// getTodayLogFile generates the file path for today's log file.
// It performs the following steps:
//  1. Retrieves the current date and time in the "America/New_York" time zone.
//  2. Formats the date as "YYYYMMDD".
//  3. Reads the log directory path from the configuration file.
//     If the variable is not set, it defaults to "/tmp/logs".
//  4. Constructs the log file path by combining the log directory path and the formatted date
//     with a ".log" extension.
//
// Returns:
// - A string representing the full path to today's log file.
func getTodayLogFile() string {
	// Get the current time in the format "2006/01/02 15:04:05"
	dt := time.Now().Format("2006/01/02 15:04:05")

	// Change time zone to America/New_York
	loc, _ := time.LoadLocation("America/New_York")
	tz, _ := time.ParseInLocation("2006/01/02 15:04:05", dt, loc)

	// Format the time in the format "2006/01/02 15:04:05"
	formattedDT := tz.Format("20060102")

	// Use the Log Path and the formatted date to create the log file path
	filepath := path.Join(LogFolder, formattedDT+".log")

	// Return the log file for today's date
	return filepath
}
