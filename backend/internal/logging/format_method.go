package logging

import "github.com/fatih/color"

// getLogMethod takes an HTTP method as input and returns two strings:
//  1. A colored version of the method name, formatted to a fixed length, where the color
//     corresponds to the HTTP method (e.g., GET is green, POST is blue, etc.).
//  2. A non-colored version of the method name, also formatted to the same fixed length.
//
// Parameters:
// - method: A string representing the HTTP method (e.g., "GET", "POST").
//
// Returns:
// - A string containing the colored and formatted method name.
// - A string containing the non-colored and formatted method name.
//
// The function uses ANSI color codes for coloring the method name based on its type.
// If the method is not one of the predefined types (GET, POST, DELETE, PUT, PATCH),
// it returns the method name without coloring.
func getLogMethod(method string) (string, string) {

	// Color the Method of the request
	var colored string
	length := LevelAndMethodLength
	switch method {
	case "GET":
		colored = color.GreenString(fixStringLength(method, length))
	case "POST":
		colored = color.BlueString(fixStringLength(method, length))
	case "DELETE":
		colored = color.RedString(fixStringLength(method, length))
	case "PUT":
		colored = color.YellowString(fixStringLength(method, length))
	case "PATCH":
		colored = color.MagentaString(fixStringLength(method, length))
	default:
		colored = method
	}

	// Return the colored method and the non-colored method
	return colored, fixStringLength(method, length)
}
