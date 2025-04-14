package logging

import "github.com/fatih/color"

// getLogPath formats a given file path for logging purposes by truncating it
// to a fixed length and applying color formatting. It returns both the
// color-formatted path and the original truncated path.
//
// Parameters:
//   - path: The file path to be formatted.
//
// Returns:
//   - string: The color-formatted version of the truncated path.
//   - string: The non-colored truncated path.
func getLogPath(path string) (string, string) {

	path = fixStringLength(path, FunctionNameAndPathLength)

	// Color the path
	colored := color.CyanString(path)

	// Return the colored path and the non-colored path
	return colored, path
}
