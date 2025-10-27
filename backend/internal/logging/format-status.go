package logging

import (
	"strconv"

	"github.com/fatih/color"
)

// getLogStatus takes an HTTP status code as an integer and returns two strings:
//  1. A colored representation of the status code for terminal output, where the color
//     corresponds to the status code range:
//     - Green for 2xx (Success)
//     - Yellow for 3xx (Redirection)
//     - Bright Red for 4xx (Client Error)
//     - Red for 5xx (Server Error)
//     - White for other codes
//  2. A non-colored, fixed-length string representation of the status code.
//     If the status code is 0, it returns an empty space instead of "[0]".
//
// Parameters:
// - status: The HTTP status code as an integer.
// Returns:
// - A colored string representation of the status code.
// - A non-colored, fixed-length string representation of the status code.
func getLogStatus(status int) (string, string) {

	// Convert the status from int to string
	statusStr := "[" + strconv.Itoa(status) + "]"
	if statusStr == "[0]" {
		statusStr = " "
	}
	statusStr = fixStringLength(statusStr, 5)

	// Color the status code
	var colored string
	switch {
	case status >= 200 && status < 300:
		colored = color.GreenString(statusStr)
	case status >= 300 && status < 400:
		colored = color.YellowString(statusStr)
	case status >= 400 && status < 500:
		colored = color.New(color.FgHiRed).Sprint(statusStr)
	case status >= 500 && status < 600:
		colored = color.RedString(statusStr)
	default:
		colored = color.WhiteString(statusStr)
	}

	// Return the colored status and the non-colored status
	return colored, statusStr
}
